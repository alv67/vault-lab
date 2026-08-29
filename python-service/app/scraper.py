import re
import time
import xml.etree.ElementTree as ET

import requests
from bs4 import BeautifulSoup

from .schemas import Exposure, ExposureRow

JUSTETF_URL = "https://www.justetf.com/en/etf-profile.html?isin={isin}"
TIMEOUT_SECONDS = 10
USER_AGENT = (
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
    "(KHTML, like Gecko) Chrome/124.0 Safari/537.36"
)

PAGE_HEADERS = {
    "User-Agent": USER_AGENT,
    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
}

COUNTRIES_ROW_TESTID = "etf-holdings_countries_row"
COUNTRIES_NAME_TESTID = "tl_etf-holdings_countries_value_name"
COUNTRIES_PCT_TESTID = "tl_etf-holdings_countries_value_percentage"

_WEIGHT_RE = re.compile(r"([\d]+(?:[.,]\d+)?)\s*%")


def _parse_weight(text):
    match = _WEIGHT_RE.search(text or "")
    if not match:
        return None
    return float(match.group(1).replace(",", "."))


def _parse_country_rows(html):
    soup = BeautifulSoup(html, "html.parser")
    rows = []
    for tr in soup.select(f'[data-testid="{COUNTRIES_ROW_TESTID}"]'):
        name_el = tr.select_one(f'[data-testid="{COUNTRIES_NAME_TESTID}"]')
        pct_el = tr.select_one(f'[data-testid="{COUNTRIES_PCT_TESTID}"]')
        name = name_el.get_text(" ", strip=True) if name_el else ""
        weight = _parse_weight(pct_el.get_text(" ", strip=True)) if pct_el else None
        if name and weight is not None:
            rows.append(ExposureRow(name=name, weight=round(weight, 2)))
    return rows


def _fetch_full_countries(session, isin, page_url):
    """Fetches the full country breakdown via the Wicket AJAX 'Show more'.

    JustETF renders only the top countries server-side and expands the rest
    through a Wicket AJAX behavior on the load-more link. The behavior URL is
    stable across sessions; it must be called with the Wicket AJAX headers on
    the same session that loaded the profile page.
    """
    page_path = page_url.removeprefix("https://www.justetf.com/")
    ajax_url = (
        "https://www.justetf.com/en/etf-profile.html"
        f"?0-1.0-holdingsSection-countries-loadMoreCountries&isin={isin}"
        f"&_wicket=1&_={int(time.time() * 1000)}"
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
            if COUNTRIES_ROW_TESTID in cdata:
                return _parse_country_rows(cdata)
        return None
    except ET.ParseError:
        return None


def fetch_exposure(isin):
    """Best-effort country exposure for an ETF. Prefers the full country list
    from the Wicket AJAX endpoint and falls back to the server-rendered top
    countries when that fails."""
    session = requests.Session()
    page = session.get(JUSTETF_URL.format(isin=isin), headers=PAGE_HEADERS, timeout=TIMEOUT_SECONDS)
    page.raise_for_status()

    try:
        countries = _fetch_full_countries(session, isin, page.url)
    except requests.RequestException:
        countries = None
    if not countries:
        countries = _parse_country_rows(page.text)

    return Exposure(countries=countries, sectors=[])