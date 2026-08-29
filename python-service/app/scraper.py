import re
import time
import xml.etree.ElementTree as ET

import requests
from bs4 import BeautifulSoup

from .schemas import EtfSearchResult, Exposure, ExposureRow

JUSTETF_URL = "https://www.justetf.com/en/etf-profile.html?isin={isin}"
JUSTETF_SEARCH_URL = "https://www.justetf.com/api/etfs/quick-search"
TIMEOUT_SECONDS = 10
USER_AGENT = (
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
    "(KHTML, like Gecko) Chrome/124.0 Safari/537.36"
)

PAGE_HEADERS = {
    "User-Agent": USER_AGENT,
    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
}

COUNTRIES = {
    "row_testid": "etf-holdings_countries_row",
    "name_testid": "tl_etf-holdings_countries_value_name",
    "pct_testid": "tl_etf-holdings_countries_value_percentage",
}
SECTORS = {
    "row_testid": "etf-holdings_sectors_row",
    "name_testid": "tl_etf-holdings_sectors_value_name",
    "pct_testid": "tl_etf-holdings_sectors_value_percentage",
}

# Wicket AJAX behavior paths that reveal the full tables (the 'Show more' links).
COUNTRIES_WICKET_PATH = "holdingsSection-countries-loadMoreCountries"
SECTORS_WICKET_PATH = "holdingsSection-sectors-loadMoreSectors"

_WEIGHT_RE = re.compile(r"([\d]+(?:[.,]\d+)?)\s*%")
_ISIN_JS_RE = re.compile(r'"isin"\s*:\s*"([A-Z]{2}[A-Z0-9]{10})"')
_ISIN_FAQ_RE = re.compile(r"The ISIN of .*? is ([A-Z]{2}[A-Z0-9]{10})\.", re.IGNORECASE)


def _parse_weight(text):
    match = _WEIGHT_RE.search(text or "")
    if not match:
        return None
    return float(match.group(1).replace(",", "."))


def _extract_isin(html):
    match = _ISIN_JS_RE.search(html or "")
    if match:
        return match.group(1)
    match = _ISIN_FAQ_RE.search(html or "")
    return match.group(1) if match else ""


def _parse_rows(html, row_testid, name_testid, pct_testid):
    soup = BeautifulSoup(html, "html.parser")
    rows = []
    for tr in soup.select(f'[data-testid="{row_testid}"]'):
        name_el = tr.select_one(f'[data-testid="{name_testid}"]')
        pct_el = tr.select_one(f'[data-testid="{pct_testid}"]')
        name = name_el.get_text(" ", strip=True) if name_el else ""
        weight = _parse_weight(pct_el.get_text(" ", strip=True)) if pct_el else None
        if name and weight is not None:
            rows.append(ExposureRow(name=name, weight=round(weight, 2)))
    return rows


def _wicket_rows(session, isin, page_url, path, row_testid, name_testid, pct_testid):
    """Fetches a full holdings table through the JustETF Wicket AJAX 'Show more'.

    JustETF renders only the top rows server-side and expands the rest through a
    Wicket AJAX behavior on the load-more link. The behavior URL is stable across
    sessions; it must be called with the Wicket AJAX headers on the same session
    that loaded the profile page.
    """
    page_path = page_url.removeprefix("https://www.justetf.com/")
    ajax_url = (
        "https://www.justetf.com/en/etf-profile.html"
        f"?0-1.0-{path}&isin={isin}&_wicket=1&_={int(time.time() * 1000)}"
    )
    headers = {
        "User-Agent": USER_AGENT,
        "Accept": "application/xml, text/xml, */*; q=0.01",
        "X-Requested-With": "XMLHttpRequest",
        "Wicket-Ajax": "true",
        "Wicket-Ajax-BaseURL": page_path,
        "Referer": page_url,
    }
    response = session.get(ajax_url, headers=headers, timeout=TIMEOUT_SECONDS)
    response.raise_for_status()
    if "ajax-response" not in response.text[:300]:
        return None

    try:
        root = ET.fromstring(response.text)
        for component in root.findall("component"):
            cdata = component.text or ""
            if row_testid in cdata:
                return _parse_rows(cdata, row_testid, name_testid, pct_testid)
        return None
    except ET.ParseError:
        return None


def fetch_exposure(isin):
    """Best-effort ETF metadata from JustETF: ISIN, full country and sector
    breakdowns. Prefers the Wicket AJAX 'Show more' tables and falls back to the
    server-rendered top rows when those calls fail."""
    session = requests.Session()
    page = session.get(JUSTETF_URL.format(isin=isin), headers=PAGE_HEADERS, timeout=TIMEOUT_SECONDS)
    page.raise_for_status()

    try:
        countries = _wicket_rows(session, isin, page.url, COUNTRIES_WICKET_PATH, **COUNTRIES)
    except requests.RequestException:
        countries = None
    if not countries:
        countries = _parse_rows(page.text, **COUNTRIES)

    try:
        sectors = _wicket_rows(session, isin, page.url, SECTORS_WICKET_PATH, **SECTORS)
    except requests.RequestException:
        sectors = None
    if not sectors:
        sectors = _parse_rows(page.text, **SECTORS)

    return Exposure(
        isin=_extract_isin(page.text),
        countries=countries,
        sectors=sectors,
    )


def search_etf(query):
    """Resolves a ticker (or name fragment) to one or more ETFs via the
    JustETF quick-search API, returning their ISIN and name."""
    response = requests.get(
        JUSTETF_SEARCH_URL,
        params={
            "locale": "en",
            "currency": "EUR",
            "universeType": "PRIVATE",
            "universeCountry": "DE",
            "limit": 10,
            "page": 0,
            "query": query,
        },
        headers={"User-Agent": USER_AGENT, "Accept": "application/json"},
        timeout=TIMEOUT_SECONDS,
    )
    response.raise_for_status()
    data = response.json()
    results = []
    for entry in data.get("etfs", []):
        isin = entry.get("isin", "")
        if not isin:
            continue
        results.append(
            EtfSearchResult(
                isin=isin,
                name=entry.get("name", ""),
                ticker=entry.get("ticker", ""),
            )
        )
    return results