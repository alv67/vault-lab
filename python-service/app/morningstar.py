"""Morningstar exposure resolver (custom, no mstarpy).

Morningstar's public pages sit behind an AWS WAF challenge and their legacy
`global.morningstar.com` JSON endpoints (used by mstarpy) are hard-blocked. This
resolver instead:

1. Bootstrap phase (lazy, cached): launches a headless Chromium under Xvfb,
   loads www.morningstar.com until the AWS WAF JS challenge auto-resolves, then
   reads the `aws-waf-token`/session cookies and fetches the MAAS bearer token
   from /api/v2/stores/maas/token. The bearer token is a JWT that authorizes the
   SAL service and expires after ~1 hour.
2. ISIN -> securityId lookup: GET /api/v2/search?q={isin} (works with the WAF
   cookies alone).
3. Data phase (plain requests, using the cached bearer + cookies):
      * https://www.us-api.morningstar.com/sal/sal-service/etf/portfolio/v2/sector/{sid}/data
        -> sector exposure, bucketed by asset class (EQUITY/FIXEDINCOME/...)
      * https://www.us-api.morningstar.com/sal/sal-service/etf/portfolio/regionalSectorIncludeCountries/{sid}/data
        -> country exposure ({fundPortfolio.countries[].name/percent}) — the
           payload already contains the full country list (51 entries, many zero);
           the website only paginates the client-side table 10 rows at a time.
      * https://www.us-api.morningstar.com/sal/sal-service/etf/portfolio/regionalSector/{sid}/data
        -> official region breakdown ({fundPortfolio.{regionKey: weight}}); the
           keys map 1:1 to canonical VaultLab region names.

Return-shape notes (confirmed from the live SAL service):
  * countries: name is a Morningstar camelCase country key (e.g. "southKorea",
    "czechRepublic", "unitedStates"); percent is a 0-100 number.
  * sectors: top-level buckets "EQUITY", "FIXEDINCOME" (note: no underscore in
    the live payload), each holding fundPortfolio.{camelCase GICS key: weight}.
    A sector bucket with all weights <= 1.0 is normalized to percentages.
"""

from __future__ import annotations

import base64
import json
import logging
import os
import threading
import time
from contextlib import contextmanager

import requests

from . import scraper as _scraper
from .schemas import EtfSearchResult, Exposure, ExposureRow

scraper_search_etf = _scraper.search_etf

logger = logging.getLogger(__name__)

# camelCase GICS sector key -> human-readable sector display name.
SECTOR_NAMES = {
    "basicMaterials": "Basic Materials",
    "communicationServices": "Communication Services",
    "consumerCyclical": "Consumer Cyclical",
    "consumerDefensive": "Consumer Defensive",
    "energy": "Energy",
    "financialServices": "Financial Services",
    "healthcare": "Healthcare",
    "industrials": "Industrials",
    "realEstate": "Real Estate",
    "technology": "Technology",
    "utilities": "Utilities",
}

# Sector buckets that may appear at the top level of the v2 sector payload.
SECTOR_BUCKETS = ("EQUITY", "FIXEDINCOME", "NOT_CLASSIFIED")

# Keys inside a fundPortfolio dict that are metadata, not sectors.
_NON_SECTOR_KEYS = {"portfolioDate", "name", "masterPortfolioId", "avgMarketCap"}

# Keys that represent the bond side of a sector breakdown rather than GICS
# sectors. For an equity fund these live in the FIXEDINCOME bucket alongside
# the (near-zero) corporate weight and must not be reported as sectors.
_NON_GICS_KEYS = {"cashAndEquivalents", "government", "municipal", "securitized", "derivative"}

_SAL_HOST = "https://www.us-api.morningstar.com/sal/sal-service/etf"
_WWW = "https://www.morningstar.com"

# Environment overrides (used by tests and the test stack to avoid a real WAF
# bootstrap): a pre-fetched bearer token and/or WAF cookies.
_MORNINGSTAR_BEARER = os.environ.get("MORNINGSTAR_BEARER") or ""
_MORNINGSTAR_COOKIES = os.environ.get("MORNINGSTAR_COOKIES") or ""

# Cache for the browser bootstrap: {bearer, cookies, expiry}.
_bootstrap_cache = {}
_bootstrap_lock = threading.Lock()


class MorningstarDataError(ValueError):
    """Raised when Morningstar exposure data cannot be retrieved or parsed."""


def _to_float(value):
    if value is None:
        return None
    try:
        return float(value)
    except (TypeError, ValueError):
        return None


def _maybe_to_percent(rows):
    """Scales a list of weights to percentages when they were 0-1 fractions.

    A genuine sub-1% weight could be mistaken for a fraction, so we only treat
    the whole list as fractions when the largest weight is <= 1.0 (and all values
    are non-negative).
    """
    if not rows:
        return rows
    weights = [row.weight for row in rows]
    if max(weights) <= 1.0 and all(w >= 0 for w in weights):
        return [ExposureRow(name=row.name, weight=round(row.weight * 100, 2)) for row in rows]
    return rows


# Known Morningstar camelCase country keys that need to become multi-word
# names so the backend geo.NormalizeCountry resolves them to ISO codes.
_MORNINGSTAR_COUNTRY_NAMES = {
    "unitedStates": "United States",
    "unitedStatesOfAmerica": "United States",
    "unitedKingdom": "United Kingdom",
    "southKorea": "South Korea",
    "southAfrica": "South Africa",
    "hongKong": "Hong Kong",
    "czechRepublic": "Czech Republic",
    "newZealand": "New Zealand",
    "saudiArabia": "Saudi Arabia",
    "unitedArabEmirates": "United Arab Emirates",
}

# Morningstar regionalSector fundPortfolio key -> canonical VaultLab region name.
# The backend consumes these names 1:1 (no geo derivation needed); any residual
# share not covered here is absorbed by the backend into "Other / Not Classified".
_MORNINGSTAR_REGION_NAMES = {
    "northAmerica": "North America",
    "latinAmerica": "Latin America",
    "unitedKingdom": "United Kingdom",
    "europeDeveloped": "Europe Developed",
    "europeEmerging": "Europe Emerging",
    "africaMiddleEast": "Africa / Middle East",
    "japan": "Japan",
    "australasia": "Australasia",
    "asiaDeveloped": "Asia Developed",
    "asiaEmerging": "Asia Emerging",
}


def _morningstar_country_name(key):
    """Converts a Morningstar country key to a readable name.

    Most keys are single words that already uppercase to the canonical name
    (e.g. "taiwan" -> "TAIWAN"). Compound camelCase keys are translated from the
    known map so the backend can resolve them.
    """
    key = str(key or "")
    if not key:
        return ""
    name = _MORNINGSTAR_COUNTRY_NAMES.get(key)
    if name:
        return name
    return key.title()


def _parse_countries(data):
    """Extracts country rows from the regionalSectorIncludeCountries shape."""
    if isinstance(data, dict):
        fund_portfolio = data.get("fundPortfolio", {})
        if isinstance(fund_portfolio, dict):
            countries = fund_portfolio.get("countries")
            if isinstance(countries, list):
                rows = []
                for item in countries:
                    if not isinstance(item, dict):
                        continue
                    name = _morningstar_country_name(item.get("name"))
                    weight = _to_float(item.get("percent"))
                    if name and weight is not None:
                        rows.append(ExposureRow(name=name, weight=weight))
                return _maybe_to_percent(rows)
    return []


def _parse_regions(data):
    """Extracts region rows from the regionalSector fundPortfolio shape.

    Only keys in _MORNINGSTAR_REGION_NAMES are reported; metadata keys such as
    portfolioDate/masterPortfolioId are skipped. Rows are sorted by weight
    descending.
    """
    if not isinstance(data, dict):
        return []
    fund_portfolio = data.get("fundPortfolio")
    if not isinstance(fund_portfolio, dict):
        return []
    rows = []
    for key, canonical_name in _MORNINGSTAR_REGION_NAMES.items():
        if key not in fund_portfolio:
            continue
        weight = _to_float(fund_portfolio.get(key))
        if weight is None:
            continue
        rows.append(ExposureRow(name=canonical_name, weight=weight))
    rows.sort(key=lambda r: r.weight, reverse=True)
    return rows


def _parse_sectors(data):
    """Extracts sector rows from the v2 sector shape.

    The payload is bucketed by asset class (EQUITY/FIXEDINCOME/...), each with a
    `fundPortfolio` dict of GICS sector key -> weight. Only ONE bucket carries the
    meaningful sector breakdown for the fund (EQUITY for equity funds, FIXEDINCOME
    for bond funds); the other bucket holds the complementary bond/cash side,
    which is not a sector split. We pick the bucket with the most sector weight.
    Fraction->percentage scaling is applied per bucket so a genuinely sub-1%
    weight is not wrongfully scaled because another bucket's weights are fractions.
    """
    if not isinstance(data, dict):
        return []
    candidate_buckets = []
    for bucket in SECTOR_BUCKETS:
        bucket_data = data.get(bucket)
        if isinstance(bucket_data, dict):
            fund_portfolio = bucket_data.get("fundPortfolio")
            rows = _sector_rows_from_fund_portfolio(fund_portfolio)
            if rows:
                candidate_buckets.append(rows)
    if not candidate_buckets:
        rows = _sector_rows_from_fund_portfolio(data)
        return _maybe_to_percent(rows)
    # Pick the bucket with the largest total sector weight.
    best = max(candidate_buckets, key=lambda rows: sum(r.weight for r in rows))
    return _maybe_to_percent(best)


def _sector_rows_from_fund_portfolio(fund_portfolio):
    if not isinstance(fund_portfolio, dict):
        return []
    rows = []
    for key, value in fund_portfolio.items():
        if key in _NON_SECTOR_KEYS or key in _NON_GICS_KEYS:
            continue
        weight = _to_float(value)
        if weight is None:
            continue
        display_name = SECTOR_NAMES.get(key, key.replace("_", " ").title())
        rows.append(ExposureRow(name=display_name, weight=weight))
    return rows


# ---------- browser bootstrap (WAF challenge + MAAS bearer token) ----------


def _jwt_expiry(token: str) -> float:
    """Returns the expiry timestamp of a JWT bearer token (0 if unparseable)."""
    try:
        payload = token.split(".")[1]
        payload += "=" * (-len(payload) % 4)
        claims = json.loads(base64.urlsafe_b64decode(payload))
        return float(claims.get("exp", 0))
    except Exception:
        return 0.0


def _chrome_options():
    from selenium.webdriver.chrome.options import Options

    options = Options()
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")
    options.add_argument("--disable-gpu")
    options.add_argument("--disable-blink-features=AutomationControlled")
    options.add_argument("--window-size=1280,1024")
    options.add_argument("--lang=en-US")
    options.add_argument(
        "--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
        "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36"
    )
    return options


@contextmanager
def _headless_browser():
    """Yields a Selenium Chrome driver running headless under Xvfb (if needed).

    Selenium Manager cannot resolve chromedriver on linux/aarch64, so the driver
    is pointed at the Debian-packaged /usr/bin/chromedriver explicitly. A virtual
    display is started when the DISPLAY env var is not already set.
    """
    from selenium import webdriver
    from selenium.webdriver.chrome.service import Service

    xvfb_proc = None
    if not os.environ.get("DISPLAY"):
        import subprocess

        xvfb_proc = subprocess.Popen(
            ["Xvfb", ":99", "-screen", "0", "1280x1024x24"],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        os.environ["DISPLAY"] = ":99"
        time.sleep(2)

    driver = webdriver.Chrome(options=_chrome_options(), service=Service("/usr/bin/chromedriver"))
    try:
        yield driver
    finally:
        try:
            driver.quit()
        except Exception:
            pass
        if xvfb_proc is not None:
            xvfb_proc.terminate()


def _load_until_real_page(driver, url, tries=4, timeout_per_try=30):
    """Loads a Morningstar page retrying until the AWS WAF challenge clears."""
    for attempt in range(tries):
        try:
            driver.get(url)
        except Exception as exc:  # pragma: no cover - defensive
            logger.warning("morningstar nav error: %s", exc)
        deadline = time.time() + timeout_per_try
        while time.time() < deadline:
            time.sleep(2)
            title = driver.title or ""
            if "Human Verification" not in title and "ERROR" not in title:
                return True
        logger.warning("morningstar WAF challenge not cleared (attempt %s)", attempt + 1)
    return False


def _browser_bootstrap():
    """Runs the WAF challenge and returns (bearer_token, cookies_dict)."""
    with _headless_browser() as driver:
        if not _load_until_real_page(driver, f"{_WWW}/"):
            raise MorningstarDataError(
                "Morningstar AWS WAF challenge could not be cleared (site blocked this IP)"
            )
        time.sleep(4)
        cookies = {c["name"]: c["value"] for c in driver.get_cookies()}
        token = driver.execute_async_script(
            """
            var cb = arguments[0];
            fetch('/api/v2/stores/maas/token', {credentials: 'include',
                  headers: {'Accept': 'application/json'}})
              .then(function(r){ return r.text(); })
              .then(function(t){ cb(t); })
              .catch(function(e){ cb('ERR:' + e); });
            """
        )
    if not token or token.startswith("ERR:"):
        raise MorningstarDataError("Could not obtain the Morningstar MAAS bearer token")
    return token, cookies


def _valid_token(token: str) -> bool:
    if not token:
        return False
    expiry = _jwt_expiry(token)
    return expiry == 0 or expiry > time.time() + 60


def _session_credentials():
    """Returns (bearer, cookies), re-bootstrapping the browser when needed."""
    if _MORNINGSTAR_BEARER:
        cookies = {}
        if _MORNINGSTAR_COOKIES:
            try:
                cookies = json.loads(_MORNINGSTAR_COOKIES)
            except ValueError:
                cookies = {}
        return _MORNINGSTAR_BEARER, cookies

    with _bootstrap_lock:
        cached = _bootstrap_cache.get("creds")
        if cached and _valid_token(cached["bearer"]):
            return cached["bearer"], cached["cookies"]
        bearer, cookies = _browser_bootstrap()
        _bootstrap_cache["creds"] = {"bearer": bearer, "cookies": cookies}
        logger.info("morningstar bootstrap refreshed (expiry=%s)", _jwt_expiry(bearer))
        return bearer, cookies


def _headers(bearer: str):
    return {
        "user-agent": (
            "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
            "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36"
        ),
        "accept": "application/json, text/plain, */*",
        "authorization": f"Bearer {bearer}",
        "referer": f"{_WWW}/",
        "origin": _WWW,
    }


def _search_security_id(isin: str, cookies: dict) -> str:
    """Resolves an ISIN to a Morningstar securityId via the v2 search API."""
    url = f"{_WWW}/api/v2/search"
    params = {"q": isin, "fields": "isin,name,securityId,investmentType", "limit": 10}
    search_headers = {
        "user-agent": _headers("")["user-agent"],
        "accept": "application/json, text/plain, */*",
        "referer": f"{_WWW}/",
        "origin": _WWW,
    }
    resp = requests.get(url, params=params, headers=search_headers, cookies=cookies, timeout=25)
    resp.raise_for_status()
    payload = resp.json()
    for item in payload.get("results", []):
        value = item.get("value", {}) if isinstance(item, dict) else {}
        if value.get("isin") == isin and value.get("investmentType") in ("FE", "FO", "FV", "FM"):
            return value["securityID"]
    if payload.get("results"):
        value = payload["results"][0].get("value", {})
        if value.get("isin") == isin:
            return value["securityID"]
    raise MorningstarDataError(f"ISIN {isin} not found on Morningstar")


# Ticker exchange suffix -> Morningstar exchange code. The suffix is the market
# a ticker is listed on (e.g. XMME.MI -> Borsa Italiana = XMIL); Morningstar
# exposes one quotation per market, and the ISIN can differ between markets (an
# ambiguous ticker like EQQQ resolves to different ISINs on different listings).
MARKET_SUFFIX_TO_EXCHANGE = {
    ".MI": "XMIL",  # Borsa Italiana
    ".DE": "XFRA",  # Frankfurt (XETRA also exists; prefer primary listing)
    ".L": "XLON",  # London
    ".SW": "XSWX",  # SIX Swiss
    ".AS": "XAMS",  # Euronext Amsterdam
    ".PA": "XPAR",  # Euronext Paris
    ".BR": "XBRU",  # Euronext Brussels
    ".LS": "XLIS",  # Euronext Lisbon
    ".ST": "XSTO",  # Nasdaq Stockholm
    ".CO": "XCPH",  # Nasdaq Copenhagen
    ".HE": "XHEL",  # Nasdaq Helsinki
    ".OL": "XOSL",  # Oslo
    ".VI": "XWBO",  # Vienna
    ".WA": "XWAR",  # Warsaw
    ".MX": "XMEX",  # Bolsa Mexicana
    ".DUB": "XDUB",  # Euronext Dublin
    ".TO": "XTSE",  # Toronto
    ".NZ": "XNZE",  # NZX
    ".AX": "XASX",  # ASX
}

# Asset types that resolve to a fund/ETF quotation in the Morningstar search.
_FUND_TYPES = ("FE", "FO", "FV", "FM")


def _search_morningstar(cookies: dict, query: str, fields: str, limit: int = 25) -> list:
    url = f"{_WWW}/api/v2/search"
    params = {"q": query, "fields": fields, "limit": limit}
    search_headers = {
        "user-agent": _headers("")["user-agent"],
        "accept": "application/json, text/plain, */*",
        "referer": f"{_WWW}/",
        "origin": _WWW,
    }
    resp = requests.get(url, params=params, headers=search_headers, cookies=cookies, timeout=25)
    resp.raise_for_status()
    return resp.json().get("results", [])


def _market_suffix(ticker: str) -> str:
    """Returns the trailing exchange suffix (e.g. '.MI') or '' when absent."""
    t = (ticker or "").strip()
    if len(t) < 3 or "." not in t:
        return ""
    base, sep, suffix = t.rpartition(".")
    if not base or not suffix or len(suffix) > 4:
        return ""
    return "." + suffix.upper()


def has_market_suffix(ticker: str) -> bool:
    """True when the ticker carries a recognized exchange suffix (e.g. XMME.MI)."""
    return _market_suffix(ticker) in MARKET_SUFFIX_TO_EXCHANGE


def resolve_market_isin(ticker: str, cookies: dict) -> tuple[str, str] | None:
    """Resolves (securityID, ISIN) for the specific market a ticker is listed on.

    Uses the trailing exchange suffix (e.g. XMME.MI -> XMIL) to pick the exact
    Morningstar quotation, since the ISIN can differ by market for ambiguous
    tickers. Returns None when the ticker has no recognized suffix or the market
    quotation cannot be found.
    """
    suffix = _market_suffix(ticker)
    exchange = MARKET_SUFFIX_TO_EXCHANGE.get(suffix)
    if not exchange:
        return None
    base_ticker = (ticker or "").strip()[: -len(suffix)]
    results = _search_morningstar(
        cookies, base_ticker, "isin,name,securityId,investmentType,ticker,exchange", limit=25
    )
    for item in results:
        value = item.get("value", {}) if isinstance(item, dict) else {}
        if value.get("exchange") == exchange and value.get("investmentType") in _FUND_TYPES:
            if value.get("isin") and value.get("securityID"):
                return value["securityID"], value["isin"]
    # Fall back to any fund quotation for the base ticker if the market is absent.
    for item in results:
        value = item.get("value", {}) if isinstance(item, dict) else {}
        if value.get("investmentType") in _FUND_TYPES:
            if value.get("isin") and value.get("securityID"):
                return value["securityID"], value["isin"]
    return None


def search_etf_morningstar(query: str) -> list:
    """Resolves a ticker (with optional market suffix) to its ISIN via Morningstar.

    Returns a single result (the market-specific quotation when a recognized
    suffix is present, otherwise the first fund match) in the shape the backend
    expects. Falls back to JustETF when Morningstar is unavailable or the query
    has no recognized market suffix.
    """
    try:
        _, cookies = _session_credentials()
        resolved = resolve_market_isin(query, cookies)
        if resolved:
            security_id, isin = resolved
            return [EtfSearchResult(isin=isin, name="", ticker=query)]
        # No recognized suffix: search by the query as-is (Morningstar returns the
        # primary listing). Only used when a bare ticker has no market suffix.
        results = _search_morningstar(cookies, query, "isin,name,securityId,investmentType,ticker", limit=10)
        for item in results:
            value = item.get("value", {}) if isinstance(item, dict) else {}
            if value.get("isin") and value.get("investmentType") in _FUND_TYPES:
                return [EtfSearchResult(isin=value["isin"], name=value.get("name", ""), ticker=value.get("ticker", "") or query)]
        return []
    except requests.RequestException:
        # Fall back to JustETF so a bare or unmapped ticker still resolves.
        return scraper_search_etf(query)


def _sal_get(bearer: str, cookies: dict, path: str, component: str) -> dict:
    url = f"{_SAL_HOST}/{path}"
    params = {
        "languageId": "en",
        "locale": "en",
        "clientId": "MDC",
        "benchmarkId": "mstarorcat",
        "component": component,
        "version": "4.71.0",
    }
    resp = requests.get(url, headers=_headers(bearer), params=params, cookies=cookies, timeout=30)
    resp.raise_for_status()
    return resp.json()


def fetch_morningstar_exposure(isin: str) -> Exposure:
    """Resolves a fund's country, sector, and region exposure from Morningstar.

    Raises MorningstarDataError if no usable country data can be retrieved so
    the endpoint can map it to a 502. Regions may be empty for funds without a
    regional breakdown and are not gated on.
    """
    bearer, cookies = _session_credentials()
    security_id = _search_security_id(isin, cookies)

    try:
        sector_data = _sal_get(
            bearer, cookies, f"portfolio/v2/sector/{security_id}/data", "sal-mip-sector-exposure"
        )
        country_data = _sal_get(
            bearer,
            cookies,
            f"portfolio/regionalSectorIncludeCountries/{security_id}/data",
            "sal-mip-country-exposure",
        )
        region_data = _sal_get(
            bearer, cookies, f"portfolio/regionalSector/{security_id}/data", "sal-mip-region"
        )
    except requests.RequestException as exc:
        raise MorningstarDataError(f"Morningstar upstream request failed: {exc}") from exc

    countries = _parse_countries(country_data)
    sectors = _parse_sectors(sector_data)
    regions = _parse_regions(region_data)

    if not countries:
        raise MorningstarDataError(f"No country data found for ISIN {isin}")

    # Country weights are kept as reported (they may not sum to 100 because
    # Morningstar buckets a residual share under "Other"); the backend absorbs
    # the residual into the "Other / Not Classified" region when deriving the
    # region dimension.
    countries.sort(key=lambda r: r.weight, reverse=True)
    sectors.sort(key=lambda r: r.weight, reverse=True)
    regions.sort(key=lambda r: r.weight, reverse=True)

    return Exposure(isin=isin, countries=countries, sectors=sectors, regions=regions)
