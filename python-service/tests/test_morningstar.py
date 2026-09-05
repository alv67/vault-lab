import json
import pytest
import requests
from fastapi.testclient import TestClient

from app import morningstar
from app.main import app
from app.morningstar import (
    MorningstarDataError,
    MorningstarWafError,
    _jwt_expiry,
    _parse_countries,
    _parse_regions,
    _parse_sectors,
    has_market_suffix,
    resolve_market_isin,
    fetch_morningstar_exposure,
)

client = TestClient(app)

ISIN = "IE00B4L5Y983"


class FakeResponse:
    def __init__(self, payload, status_code=200):
        self._payload = payload
        self.status_code = status_code

    def json(self):
        return self._payload

    def raise_for_status(self):
        if self.status_code >= 400:
            raise requests.HTTPError(f"{self.status_code}")


def fake_search_payload(isin=ISIN, security_id="0P0000TSO8"):
    return {
        "results": [
            {
                "type": "security",
                "value": {
                    "isin": isin,
                    "name": "iShares Core MSCI World UCITS ETF",
                    "securityID": security_id,
                    "investmentType": "FE",
                },
            }
        ]
    }


def fake_sector_payload():
    return {
        "EQUITY": {
            "fundPortfolio": {
                "portfolioDate": "2026-07-31",
                "technology": 25.46,
                "financialServices": 15.2,
                "energy": 40.0,
            }
        }
    }


def fake_country_payload():
    return {
        "fundPortfolio": {
            "countries": [
                {"name": "unitedStates", "percent": 58.85},
                {"name": "japan", "percent": "5.92"},
                {"name": "southKorea", "percent": 0.2346},
            ]
        }
    }


def fake_region_payload():
    return {
        "fundPortfolio": {
            "portfolioDate": "2026-07-31",
            "masterPortfolioId": "0P0000TSO8",
            "northAmerica": 62.4,
            "europeDeveloped": 15.1,
            "japan": 6.4,
            "asiaDeveloped": 5.0,
            "unitedKingdom": 4.2,
            "asiaEmerging": 3.2,
            "australasia": 2.3,
            "latinAmerica": 1.1,
            "africaMiddleEast": 0.9,
            "europeEmerging": 0.8,
        }
    }


def patch_session(
    monkeypatch, country_payload=None, sector_payload=None, search_payload=None, region_payload=None
):
    country_payload = fake_country_payload() if country_payload is None else country_payload
    sector_payload = fake_sector_payload() if sector_payload is None else sector_payload
    search_payload = fake_search_payload() if search_payload is None else search_payload
    region_payload = fake_region_payload() if region_payload is None else region_payload

    def fake_requests_get(url, *args, **kwargs):
        if "/api/v2/search" in url:
            return FakeResponse(search_payload)
        if "portfolio/v2/sector/" in url:
            return FakeResponse(sector_payload)
        if "regionalSectorIncludeCountries" in url:
            return FakeResponse(country_payload)
        if "portfolio/regionalSector/" in url:
            return FakeResponse(region_payload)
        raise AssertionError(f"unexpected URL: {url}")

    monkeypatch.setattr("app.morningstar._MORNINGSTAR_BEARER", "test-bearer-token")
    monkeypatch.setattr("app.morningstar._MORNINGSTAR_COOKIES", "{}")
    monkeypatch.setattr("app.morningstar.requests.get", fake_requests_get)


# ---------- JWT helper ----------

def test_jwt_expiry_extracts_exp_claim():
    # header.payload.signature with an exp claim far in the future
    token = "a.b.c"
    assert _jwt_expiry(token) == 0.0
    # build a real JWT-like token
    import base64

    def b64(s):
        return base64.urlsafe_b64encode(s.encode()).rstrip(b"=").decode()

    claims = b64('{"exp": 2000000000}')
    token = b64('{"alg":"none"}') + "." + claims + ".sig"
    assert _jwt_expiry(token) == 2000000000.0


# ---------- parser unit tests ----------

def test_parse_countries_dict_shape(monkeypatch):
    data = {
        "fundPortfolio": {
            "countries": [
                {"name": "United States", "percent": 58.85},
                {"name": "Japan", "percent": "5.92"},
            ]
        }
    }
    rows = _parse_countries(data)
    assert [r.name for r in rows] == ["United States", "Japan"]
    assert rows[0].weight == 58.85
    assert rows[1].weight == 5.92


def test_parse_countries_string_and_numeric_percent(monkeypatch):
    data = {
        "fundPortfolio": {
            "countries": [
                {"name": "United States", "percent": "58.85"},
                {"name": "Japan", "percent": 5.92},
            ]
        }
    }
    rows = _parse_countries(data)
    assert rows[0].weight == 58.85
    assert rows[1].weight == 5.92


def test_parse_countries_skips_missing_or_bad_weight():
    data = {
        "fundPortfolio": {
            "countries": [
                {"name": "United States", "percent": 58.85},
                {"name": "No Percent", "percent": None},
                {"name": "Bad Percent", "percent": "n/a"},
                "not a dict",
            ]
        }
    }
    rows = _parse_countries(data)
    assert [r.name for r in rows] == ["United States"]
    assert rows[0].weight == 58.85


def test_parse_countries_fraction_to_percentage():
    data = {
        "fundPortfolio": {
            "countries": [
                {"name": "United States", "percent": 0.5885},
                {"name": "Japan", "percent": 0.0592},
            ]
        }
    }
    rows = _parse_countries(data)
    assert rows[0].weight == 58.85
    assert rows[1].weight == 5.92


def test_parse_countries_morningstar_camelcase_names():
    data = {
        "fundPortfolio": {
            "countries": [
                {"name": "southKorea", "percent": 20.3},
                {"name": "hongKong", "percent": 1.2},
                {"name": "czechRepublic", "percent": 0.5},
                {"name": "japan", "percent": 8.0},
            ]
        }
    }
    rows = _parse_countries(data)
    assert [r.name for r in rows] == ["South Korea", "Hong Kong", "Czech Republic", "Japan"]


def test_parse_countries_empty_structures():
    assert _parse_countries({}) == []
    assert _parse_countries({"fundPortfolio": {"countries": []}}) == []
    assert _parse_countries({"fundPortfolio": {}}) == []
    assert _parse_countries(None) == []


def test_parse_regions_maps_names_and_skips_metadata():
    rows = _parse_regions(fake_region_payload())
    assert [r.name for r in rows] == [
        "North America",
        "Europe Developed",
        "Japan",
        "Asia Developed",
        "United Kingdom",
        "Asia Emerging",
        "Australasia",
        "Latin America",
        "Africa / Middle East",
        "Europe Emerging",
    ]
    assert rows[0].weight == 62.4
    assert rows[-1].weight == 0.8
    names = [r.name for r in rows]
    assert "portfolioDate" not in names
    assert "masterPortfolioId" not in names


def test_parse_regions_empty_structures():
    assert _parse_regions({}) == []
    assert _parse_regions({"fundPortfolio": {}}) == []
    assert _parse_regions({"fundPortfolio": {"portfolioDate": "2026-07-31"}}) == []
    assert _parse_regions({"fundPortfolio": "not a dict"}) == []
    assert _parse_regions(None) == []


def test_parse_sectors_dict_shape():
    data = {
        "EQUITY": {
            "fundPortfolio": {
                "technology": 35.46,
                "financialServices": 15.2,
                "portfolioDate": "2025-01-31",
                "name": "iShares Core MSCI World",
            }
        }
    }
    rows = _parse_sectors(data)
    names = [r.name for r in rows]
    assert "Technology" in names
    assert "Financial Services" in names
    assert len(rows) == 2
    assert next(r.weight for r in rows if r.name == "Technology") == 35.46


def test_parse_sectors_fixed_income_and_other_buckets():
    data = {
        "EQUITY": {"fundPortfolio": {"technology": 40.0}},
        "FIXEDINCOME": {"fundPortfolio": {"utilities": 10.0}},
        "NOT_CLASSIFIED": {"fundPortfolio": {"energy": 6.0}},
    }
    # Only the bucket with the largest total sector weight is reported.
    rows = _parse_sectors(data)
    assert [r.name for r in rows] == ["Technology"]
    assert rows[0].weight == 40.0


def test_parse_sectors_prefers_equity_over_bond_side():
    data = {
        "EQUITY": {"fundPortfolio": {"technology": 41.0, "financialServices": 19.0}},
        "FIXEDINCOME": {
            "fundPortfolio": {
                "cashAndEquivalents": 90.0,
                "corporate": 9.8,
                "government": 0.0,
                "municipal": 0.0,
                "securitized": 0.0,
                "derivative": 0.0,
            }
        },
    }
    rows = _parse_sectors(data)
    assert {r.name for r in rows} == {"Technology", "Financial Services"}
    assert sum(r.weight for r in rows) == 60.0


def test_parse_sectors_bond_fund_uses_fixed_income_bucket():
    data = {
        "EQUITY": {"fundPortfolio": {}},
        "FIXEDINCOME": {"fundPortfolio": {"corporate": 60.0, "government": 40.0}},
    }
    rows = _parse_sectors(data)
    assert [r.name for r in rows] == ["Corporate"]
    assert rows[0].weight == 60.0


def test_parse_sectors_flat_fallback():
    data = {
        "technology": 35.46,
        "financialServices": 15.2,
    }
    rows = _parse_sectors(data)
    assert rows[0].name == "Technology"
    assert rows[0].weight == 35.46


def test_parse_sectors_fraction_to_percentage():
    data = {
        "EQUITY": {
            "fundPortfolio": {
                "technology": 0.3546,
                "financialServices": 0.152,
            }
        }
    }
    rows = _parse_sectors(data)
    assert rows[0].weight == 35.46
    assert rows[1].weight == 15.2


def test_parse_sectors_scales_per_bucket():
    data = {
        "EQUITY": {"fundPortfolio": {"technology": 0.5, "healthcare": 40.0}},
        "NOT_CLASSIFIED": {"fundPortfolio": {"energy": 0.4, "utilities": 10.0}},
    }
    # EQUITY wins (40.5 > 10.4); its max weight is 40 so no fraction scaling.
    rows = _parse_sectors(data)
    rows_by_name = {r.name: r.weight for r in rows}
    assert rows_by_name["Technology"] == 0.5
    assert rows_by_name["Healthcare"] == 40.0


def test_parse_sectors_bucket_all_fractions_scales():
    data = {
        "EQUITY": {"fundPortfolio": {"technology": 0.3546, "healthcare": 0.3}},
    }
    rows = _parse_sectors(data)
    rows_by_name = {r.name: r.weight for r in rows}
    assert rows_by_name["Technology"] == 35.46
    assert rows_by_name["Healthcare"] == 30.0


def test_parse_sectors_skips_non_gics_keys():
    data = {
        "FIXEDINCOME": {
            "fundPortfolio": {
                "cashAndEquivalents": 90.0,
                "government": 8.0,
                "municipal": 0.0,
                "securitized": 1.0,
                "derivative": 0.0,
            }
        }
    }
    assert _parse_sectors(data) == []


def test_parse_sectors_skips_metadata_and_empty():
    data = {"EQUITY": {"fundPortfolio": {"portfolioDate": "2025-01-31", "name": "x"}}}
    assert _parse_sectors(data) == []
    assert _parse_sectors({}) == []
    assert _parse_sectors({"EQUITY": {}}) == []
    assert _parse_sectors(None) == []


# ---------- fetch_morningstar_exposure tests ----------

def test_fetch_morningstar_exposure_sorts_descending(monkeypatch):
    patch_session(monkeypatch)
    exposure = fetch_morningstar_exposure(ISIN)
    assert exposure.isin == ISIN
    assert exposure.countries[0].name == "United States"
    # Country weights are kept as reported (not scaled to 100); the residual is
    # absorbed by the backend into the "Other / Not Classified" region.
    assert round(sum(c.weight for c in exposure.countries), 2) == 65.0
    assert exposure.countries[0].weight == 58.85
    assert exposure.sectors[0].name == "Energy"
    assert exposure.sectors[0].weight == 40.0


def test_fetch_morningstar_exposure_includes_regions(monkeypatch):
    patch_session(monkeypatch)
    exposure = fetch_morningstar_exposure(ISIN)
    assert exposure.isin == ISIN
    assert exposure.regions[0].name == "North America"
    assert exposure.regions[0].weight == 62.4
    weights = [r.weight for r in exposure.regions]
    assert weights == sorted(weights, reverse=True)
    canonical = {
        "North America",
        "Latin America",
        "United Kingdom",
        "Europe Developed",
        "Europe Emerging",
        "Africa / Middle East",
        "Japan",
        "Australasia",
        "Asia Developed",
        "Asia Emerging",
    }
    assert all(r.name in canonical for r in exposure.regions)


def test_fetch_morningstar_exposure_picks_etf_security(monkeypatch):
    patch_session(
        monkeypatch,
        search_payload={
            "results": [
                {"value": {"isin": ISIN, "securityID": "XXX", "investmentType": "EQ"}},
                {"value": {"isin": ISIN, "securityID": "0P0000TSO8", "investmentType": "FE"}},
            ]
        },
    )
    exposure = fetch_morningstar_exposure(ISIN)
    assert exposure.countries[0].name == "United States"


def test_fetch_morningstar_exposure_empty_countries_raises(monkeypatch):
    patch_session(monkeypatch, country_payload={"fundPortfolio": {"countries": []}})
    with pytest.raises(MorningstarDataError):
        fetch_morningstar_exposure(ISIN)


def test_fetch_morningstar_exposure_upstream_error_raises(monkeypatch):
    def bad_get(url, *args, **kwargs):
        if "/api/v2/search" in url:
            return FakeResponse(fake_search_payload())
        if "regionalSectorIncludeCountries" in url:
            return FakeResponse({}, status_code=500)
        return FakeResponse(fake_sector_payload())

    monkeypatch.setattr("app.morningstar._MORNINGSTAR_BEARER", "test-bearer-token")
    monkeypatch.setattr("app.morningstar._MORNINGSTAR_COOKIES", "{}")
    monkeypatch.setattr("app.morningstar.requests.get", bad_get)
    with pytest.raises(MorningstarDataError):
        fetch_morningstar_exposure(ISIN)


def test_fetch_morningstar_exposure_isin_not_found(monkeypatch):
    monkeypatch.setattr("app.morningstar._MORNINGSTAR_BEARER", "test-bearer-token")
    monkeypatch.setattr("app.morningstar._MORNINGSTAR_COOKIES", "{}")
    monkeypatch.setattr(
        "app.morningstar.requests.get",
        lambda url, *a, **k: FakeResponse({"results": []}),
    )
    with pytest.raises(MorningstarDataError):
        fetch_morningstar_exposure(ISIN)


# ---------- WAF resilience (retry once with a fresh session) ----------

def test_fetch_morningstar_exposure_retries_with_fresh_session_on_challenge(monkeypatch):
    bootstraps = {"count": 0}
    sector_calls = {"count": 0}

    def fake_bootstrap():
        bootstraps["count"] += 1
        return f"bearer-{bootstraps['count']}", {}

    def fake_get(url, *a, **k):
        if "/api/v2/search" in url:
            return FakeResponse(fake_search_payload())
        if "portfolio/v2/sector/" in url:
            sector_calls["count"] += 1
            if sector_calls["count"] == 1:
                # Stale session: the WAF answers with an HTML challenge page.
                raise requests.exceptions.JSONDecodeError("Expecting value", "", 0)
            return FakeResponse(fake_sector_payload())
        if "regionalSectorIncludeCountries" in url:
            return FakeResponse(fake_country_payload())
        if "portfolio/regionalSector/" in url:
            return FakeResponse(fake_region_payload())
        raise AssertionError(f"unexpected URL: {url}")

    monkeypatch.setattr("app.morningstar._MORNINGSTAR_BEARER", "")
    monkeypatch.setattr("app.morningstar._MORNINGSTAR_COOKIES", "")
    monkeypatch.setattr("app.morningstar._bootstrap_cache", {})
    monkeypatch.setattr("app.morningstar._browser_bootstrap", fake_bootstrap)
    monkeypatch.setattr("app.morningstar.requests.get", fake_get)

    exposure = fetch_morningstar_exposure(ISIN)
    # The challenge forced exactly one invalidation + browser re-bootstrap.
    assert bootstraps["count"] == 2
    assert morningstar._bootstrap_cache["creds"]["bearer"] == "bearer-2"
    assert exposure.countries[0].name == "United States"
    assert exposure.regions[0].name == "North America"


def test_fetch_morningstar_exposure_challenge_persists_raises(monkeypatch):
    search_calls = {"count": 0}
    bootstraps = {"count": 0}

    def fake_bootstrap():
        bootstraps["count"] += 1
        return f"bearer-{bootstraps['count']}", {}

    def fake_get(url, *a, **k):
        if "/api/v2/search" in url:
            search_calls["count"] += 1
            return FakeResponse(fake_search_payload())
        raise requests.exceptions.JSONDecodeError("Expecting value", "", 0)

    monkeypatch.setattr("app.morningstar._MORNINGSTAR_BEARER", "")
    monkeypatch.setattr("app.morningstar._MORNINGSTAR_COOKIES", "")
    monkeypatch.setattr("app.morningstar._bootstrap_cache", {})
    monkeypatch.setattr("app.morningstar._browser_bootstrap", fake_bootstrap)
    monkeypatch.setattr("app.morningstar.requests.get", fake_get)

    with pytest.raises(MorningstarWafError) as err:
        fetch_morningstar_exposure(ISIN)
    assert "WAF challenge could not be passed" in str(err.value)
    assert ISIN in str(err.value)
    # Exactly one retry of the whole flow, no more: search ran twice and each
    # attempt re-bootstrapped after the cache invalidation.
    assert search_calls["count"] == 2
    assert bootstraps["count"] == 2


def test_fetch_morningstar_exposure_connection_error_does_not_retry(monkeypatch):
    sessions = {"count": 0}

    def fake_credentials():
        sessions["count"] += 1
        return "tok", {}

    def fake_get(url, *a, **k):
        if "/api/v2/search" in url:
            return FakeResponse(fake_search_payload())
        raise requests.ConnectionError("connection reset by peer")

    monkeypatch.setattr("app.morningstar._session_credentials", fake_credentials)
    monkeypatch.setattr("app.morningstar.requests.get", fake_get)

    with pytest.raises(MorningstarDataError) as err:
        fetch_morningstar_exposure(ISIN)
    assert "upstream request failed" in str(err.value)
    # Non-WAF transport errors are wrapped immediately: no re-bootstrap loop.
    assert sessions["count"] == 1


# ---------- browser bootstrap resilience (Chrome session crashes) ----------

def test_session_credentials_retries_bootstrap_after_chrome_crash(monkeypatch):
    from selenium.common.exceptions import SessionNotCreatedException

    boots = {"count": 0}
    kills = {"count": 0}

    def fake_bootstrap():
        boots["count"] += 1
        if boots["count"] == 1:
            raise SessionNotCreatedException("session not created: Chrome instance exited")
        return "bearer-fresh", {"aws-waf-token": "tok"}

    monkeypatch.setattr("app.morningstar._MORNINGSTAR_BEARER", "")
    monkeypatch.setattr("app.morningstar._MORNINGSTAR_COOKIES", "")
    monkeypatch.setattr("app.morningstar._bootstrap_cache", {})
    monkeypatch.setattr("app.morningstar._browser_bootstrap", fake_bootstrap)
    monkeypatch.setattr(
        "app.morningstar._kill_stale_browsers",
        lambda: kills.__setitem__("count", kills["count"] + 1),
    )

    bearer, cookies = morningstar._session_credentials()
    assert bearer == "bearer-fresh"
    assert cookies == {"aws-waf-token": "tok"}
    # Exactly one cleanup between the crashed bootstrap and the successful one.
    assert boots["count"] == 2
    assert kills["count"] == 1
    assert morningstar._bootstrap_cache["creds"]["bearer"] == "bearer-fresh"


def test_session_credentials_gives_up_after_two_crashed_bootstraps(monkeypatch):
    from selenium.common.exceptions import WebDriverException

    boots = {"count": 0}
    kills = {"count": 0}

    def fake_bootstrap():
        boots["count"] += 1
        raise WebDriverException("chrome not reachable")

    def expired_token():
        import base64

        claims = base64.urlsafe_b64encode(b'{"exp": 1}').rstrip(b"=").decode()
        return f"a.{claims}.c"

    monkeypatch.setattr("app.morningstar._MORNINGSTAR_BEARER", "")
    monkeypatch.setattr("app.morningstar._MORNINGSTAR_COOKIES", "")
    monkeypatch.setattr(
        "app.morningstar._bootstrap_cache",
        {"creds": {"bearer": expired_token(), "cookies": {}}},
    )
    monkeypatch.setattr("app.morningstar._browser_bootstrap", fake_bootstrap)
    monkeypatch.setattr(
        "app.morningstar._kill_stale_browsers",
        lambda: kills.__setitem__("count", kills["count"] + 1),
    )

    with pytest.raises(MorningstarDataError) as err:
        morningstar._session_credentials()
    assert "Chrome headless could not start" in str(err.value)
    # Bounded at two bootstraps (no infinite loop), each failure cleans up and
    # invalidates the cache so no poisoned credentials linger.
    assert boots["count"] == 2
    assert kills["count"] == 2
    assert "creds" not in morningstar._bootstrap_cache


def test_kill_stale_browsers_runs_pkill_and_swallows_errors(monkeypatch):
    calls = []

    def fake_run(cmd, **kwargs):
        calls.append((cmd, kwargs))
        raise OSError("pkill: command not found")

    monkeypatch.setattr(morningstar.subprocess, "run", fake_run)
    morningstar._kill_stale_browsers()  # must never raise

    cmd, kwargs = calls[0]
    assert cmd[:3] == ["pkill", "-9", "-f"]
    pattern = cmd[3]
    assert "chromedriver" in pattern and "chrome_crashpad" in pattern and "chromium" in pattern
    assert "python" not in pattern  # must not target the service process itself
    assert kwargs.get("check") is False
    assert kwargs.get("timeout") == 5


def test_exposure_flow_crashed_chrome_does_not_break_waf_retry(monkeypatch):
    # A WAF-challenge retry calls _session_credentials again; the browser
    # bootstrap there now goes through the crash-retry wrapper transparently.
    bootstraps = {"count": 0}

    def fake_bootstrap():
        bootstraps["count"] += 1
        return f"bearer-{bootstraps['count']}", {}

    monkeypatch.setattr("app.morningstar._MORNINGSTAR_BEARER", "")
    monkeypatch.setattr("app.morningstar._MORNINGSTAR_COOKIES", "")
    monkeypatch.setattr("app.morningstar._bootstrap_cache", {})
    monkeypatch.setattr("app.morningstar._browser_bootstrap", fake_bootstrap)
    kills = []
    monkeypatch.setattr(
        "app.morningstar._kill_stale_browsers", lambda: kills.append(True)
    )

    bearer, _ = morningstar._session_credentials()
    assert bearer == "bearer-1"
    assert kills == []  # healthy bootstrap must not trigger any cleanup

    morningstar._invalidate_bootstrap_cache()
    bearer, _ = morningstar._session_credentials()
    assert bearer == "bearer-2"
    assert kills == []


# ---------- endpoint tests ----------

def test_morningstar_exposure_ok(monkeypatch):
    patch_session(monkeypatch)
    response = client.get(f"/api/v1/etf/{ISIN}/morningstar-exposure")
    assert response.status_code == 200
    body = response.json()
    assert body["isin"] == ISIN
    assert body["countries"][0]["name"] == "United States"
    assert round(sum(c["weight"] for c in body["countries"]), 2) == 65.0
    assert body["sectors"][0]["name"] == "Energy"
    assert body["sectors"][0]["weight"] == 40.0
    assert body["regions"][0]["name"] == "North America"
    assert body["regions"][0]["weight"] == 62.4


def test_morningstar_exposure_empty_countries_502(monkeypatch):
    patch_session(monkeypatch, country_payload={"fundPortfolio": {"countries": []}})
    response = client.get(f"/api/v1/etf/{ISIN}/morningstar-exposure")
    assert response.status_code == 502
    assert "detail" in response.json()


def test_morningstar_exposure_parse_error_502(monkeypatch):
    def fake_get(url, *a, **k):
        return FakeResponse({})

    monkeypatch.setattr("app.morningstar._MORNINGSTAR_BEARER", "test-bearer-token")
    monkeypatch.setattr("app.morningstar._MORNINGSTAR_COOKIES", "{}")
    monkeypatch.setattr("app.morningstar.requests.get", fake_get)
    response = client.get(f"/api/v1/etf/{ISIN}/morningstar-exposure")
    assert response.status_code == 502


def test_morningstar_exposure_request_error_502(monkeypatch):
    def fake_get(url, *a, **k):
        raise requests.ConnectionError("connection refused")

    monkeypatch.setattr("app.morningstar._MORNINGSTAR_BEARER", "test-bearer-token")
    monkeypatch.setattr("app.morningstar._MORNINGSTAR_COOKIES", "{}")
    monkeypatch.setattr("app.morningstar.requests.get", fake_get)
    response = client.get(f"/api/v1/etf/{ISIN}/morningstar-exposure")
    assert response.status_code == 502
    assert "detail" in response.json()


def test_morningstar_exposure_waf_persists_502(monkeypatch):
    def fake_get(url, *a, **k):
        if "/api/v2/search" in url:
            return FakeResponse(fake_search_payload())
        raise requests.exceptions.JSONDecodeError("Expecting value", "", 0)

    monkeypatch.setattr("app.morningstar._session_credentials", lambda: ("tok", {}))
    monkeypatch.setattr("app.morningstar.requests.get", fake_get)
    response = client.get(f"/api/v1/etf/{ISIN}/morningstar-exposure")
    assert response.status_code == 502
    assert "WAF challenge could not be passed" in response.json()["detail"]


def test_morningstar_exposure_chrome_crash_502_clear_detail(monkeypatch):
    from selenium.common.exceptions import SessionNotCreatedException

    def fake_bootstrap():
        raise SessionNotCreatedException("session not created: Chrome instance exited")

    monkeypatch.setattr("app.morningstar._MORNINGSTAR_BEARER", "")
    monkeypatch.setattr("app.morningstar._MORNINGSTAR_COOKIES", "")
    monkeypatch.setattr("app.morningstar._bootstrap_cache", {})
    monkeypatch.setattr("app.morningstar._browser_bootstrap", fake_bootstrap)
    monkeypatch.setattr("app.morningstar._kill_stale_browsers", lambda: None)

    response = client.get(f"/api/v1/etf/{ISIN}/morningstar-exposure")
    assert response.status_code == 502
    # No raw "Message: session not created" leaking through the generic handler.
    assert "Chrome headless could not start" in response.json()["detail"]

# ---------- market-suffix resolution tests ----------

def test_has_market_suffix_recognizes_known_suffixes():
    assert has_market_suffix("XMME.MI")
    assert has_market_suffix("VWCE.DE")
    assert has_market_suffix("SWDA.L")
    assert not has_market_suffix("SWDA")
    assert not has_market_suffix("XMME.XX")  # unrecognized
    assert not has_market_suffix("AB")


def test_resolve_market_isin_picks_matching_exchange(monkeypatch):
    from app.morningstar import _search_morningstar

    results = [
        {"value": {"exchange": "XFRA", "securityID": "0P0001B6W4", "isin": "IE00BTJRMP35", "investmentType": "FE", "ticker": "XMME"}},
        {"value": {"exchange": "XMIL", "securityID": "0P0001BPHH", "isin": "IE00BTJRMP35", "investmentType": "FE", "ticker": "XMME"}},
        {"value": {"exchange": "XMEX", "securityID": "0P0001RAGI", "isin": "IE00BTJRMP35", "investmentType": "FE", "ticker": "XMME"}},
    ]
    monkeypatch.setattr("app.morningstar._search_morningstar", lambda *a, **k: results)
    sid, isin = resolve_market_isin("XMME.DE", {})
    assert sid == "0P0001B6W4"
    assert isin == "IE00BTJRMP35"


def test_resolve_market_isin_fallback_to_any_fund(monkeypatch):
    from app.morningstar import _search_morningstar

    results = [
        {"value": {"exchange": "XLON", "securityID": "0P0001Q", "isin": "IE00X", "investmentType": "EQ", "ticker": "X"}},
        {"value": {"exchange": "XSWX", "securityID": "0P0001A", "isin": "IE00B4L5Y983", "investmentType": "FE", "ticker": "SWDA"}},
    ]
    monkeypatch.setattr("app.morningstar._search_morningstar", lambda *a, **k: results)
    # .SW not in the mock results -> fallback to first fund match
    sid, isin = resolve_market_isin("SWDA.SW", {})
    assert isin == "IE00B4L5Y983"


def test_resolve_market_isin_none_for_bare_ticker():
    assert resolve_market_isin("SWDA", {}) is None


def test_search_etf_morningstar_uses_resolve(monkeypatch):
    from app.schemas import EtfSearchResult

    monkeypatch.setattr("app.morningstar._session_credentials", lambda: ("tok", {}))
    monkeypatch.setattr(
        "app.morningstar.resolve_market_isin", lambda *a, **k: ("0P0001BPHH", "IE00BTJRMP35")
    )
    out = morningstar.search_etf_morningstar("XMME.MI")
    assert out[0].isin == "IE00BTJRMP35"


def test_search_etf_morningstar_retries_with_fresh_session_on_challenge(monkeypatch):
    attempts = {"count": 0}

    def fake_get(url, *a, **k):
        attempts["count"] += 1
        if attempts["count"] == 1:
            # Stale WAF session: challenge page instead of JSON.
            raise requests.exceptions.JSONDecodeError("Expecting value", "", 0)
        return FakeResponse(
            {
                "results": [
                    {
                        "value": {
                            "isin": ISIN,
                            "name": "iShares Core MSCI World UCITS ETF",
                            "investmentType": "FE",
                            "ticker": "SWDA",
                        }
                    }
                ]
            }
        )

    monkeypatch.setattr("app.morningstar._session_credentials", lambda: ("tok", {}))
    monkeypatch.setattr("app.morningstar.requests.get", fake_get)
    monkeypatch.setattr(
        "app.morningstar.scraper_search_etf",
        lambda q: pytest.fail("JustETF fallback must not run when the retry succeeds"),
    )

    out = morningstar.search_etf_morningstar("SWDA")
    assert attempts["count"] == 2
    assert out[0].isin == ISIN


def test_search_etf_morningstar_falls_back_when_challenge_persists(monkeypatch):
    from app.schemas import EtfSearchResult

    def fake_get(url, *a, **k):
        raise requests.exceptions.JSONDecodeError("Expecting value", "", 0)

    monkeypatch.setattr("app.morningstar._session_credentials", lambda: ("tok", {}))
    monkeypatch.setattr("app.morningstar.requests.get", fake_get)
    monkeypatch.setattr(
        "app.morningstar.scraper_search_etf",
        lambda q: [EtfSearchResult(isin="FAKEISIN", name="from justetf", ticker=q)],
    )

    out = morningstar.search_etf_morningstar("SWDA")
    assert out[0].name == "from justetf"
