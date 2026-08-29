from app.scraper import (
    COUNTRIES,
    SECTORS,
    _extract_isin,
    _normalize_ticker,
    _parse_rows,
)

COUNTRY_HTML = """
<table class="table mb-0" data-testid="etf-holdings_countries_table">
  <tbody>
    <tr data-testid="etf-holdings_countries_row">
      <td data-testid="tl_etf-holdings_countries_value_name">United States</td>
      <td><div class="right ws"><span data-testid="tl_etf-holdings_countries_value_percentage">58.85%</span></div></td>
    </tr>
    <tr data-testid="etf-holdings_countries_row">
      <td data-testid="tl_etf-holdings_countries_value_name">Japan</td>
      <td><div class="right ws"><span data-testid="tl_etf-holdings_countries_value_percentage">5.92%</span></div></td>
    </tr>
    <tr data-testid="etf-holdings_countries_row">
      <td data-testid="tl_etf-holdings_countries_value_name">No Weight</td>
      <td><div class="right ws"><span></span></div></td>
    </tr>
  </tbody>
</table>
"""

SECTOR_HTML = """
<table class="table mb-0" data-testid="etf-holdings_sectors_table">
  <tbody>
    <tr data-testid="etf-holdings_sectors_row">
      <td data-testid="tl_etf-holdings_sectors_value_name">Technology</td>
      <td><div class="right ws"><span data-testid="tl_etf-holdings_sectors_value_percentage">43.04%</span></div></td>
    </tr>
    <tr data-testid="etf-holdings_sectors_row">
      <td data-testid="tl_etf-holdings_sectors_value_name">Finance</td>
      <td><div class="right ws"><span data-testid="tl_etf-holdings_sectors_value_percentage">22.38%</span></div></td>
    </tr>
  </tbody>
</table>
"""

PAGE_HTML = (
    "<html><body><script>window.GLOBALS = window.GLOBALS || {};"
    'var x = {"isin":"IE00BTJRMP35"};</script></body></html>'
)

FAQ_HTML = "<div>The ISIN of Xtrackers MSCI Emerging Markets UCITS ETF 1C is IE00BTJRMP35.</div>"


def test_parse_country_rows():
    rows = _parse_rows(COUNTRY_HTML, **COUNTRIES)
    assert [r.name for r in rows] == ["United States", "Japan"]
    assert rows[0].weight == 58.85
    assert rows[1].weight == 5.92


def test_parse_sector_rows():
    rows = _parse_rows(SECTOR_HTML, **SECTORS)
    assert [r.name for r in rows] == ["Technology", "Finance"]
    assert rows[0].weight == 43.04


def test_parse_skips_rows_without_weight():
    assert len(_parse_rows(COUNTRY_HTML, **COUNTRIES)) == 2


def test_parse_empty_document():
    assert _parse_rows("<html><body></body></html>", **COUNTRIES) == []


def test_parse_comma_decimal():
    html = COUNTRY_HTML.replace("5.92%", "5,92%")
    rows = _parse_rows(html, **COUNTRIES)
    assert rows[1].weight == 5.92


def test_extract_isin_from_globals():
    assert _extract_isin(PAGE_HTML) == "IE00BTJRMP35"


def test_extract_isin_from_faq():
    assert _extract_isin(FAQ_HTML) == "IE00BTJRMP35"


def test_extract_isin_empty():
    assert _extract_isin("<html></html>") == ""


def test_normalize_ticker():
    assert _normalize_ticker("SMEA.MI") == "SMEA"
    assert _normalize_ticker("EUNL.DE") == "EUNL"
    assert _normalize_ticker("CSPX.L") == "CSPX"
    assert _normalize_ticker("VWCE") == "VWCE"
    assert _normalize_ticker("  IWDA.DE  ") == "IWDA"
    assert _normalize_ticker("A.L") == "A.L"
    assert _normalize_ticker("") == ""


def test_search_etf_queries_normalized_ticker(monkeypatch):
    captured = {}

    class FakeResponse:
        def raise_for_status(self):
            pass

        def json(self):
            return {"etfs": [{"isin": "IE00B4K48X80", "name": "iShares Core MSCI Europe UCITS ETF EUR (Acc)"}]}

    def fake_get(url, params=None, **kwargs):
        captured["params"] = params
        return FakeResponse()

    monkeypatch.setattr("app.scraper.requests.get", fake_get)
    from app.scraper import search_etf

    results = search_etf("SMEA.MI")
    assert captured["params"]["query"] == "SMEA"
    assert results[0].isin == "IE00B4K48X80"