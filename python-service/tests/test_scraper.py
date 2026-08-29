from app.scraper import _parse_country_rows

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


def test_parse_country_rows():
    rows = _parse_country_rows(COUNTRY_HTML)
    assert [r.name for r in rows] == ["United States", "Japan"]
    assert rows[0].weight == 58.85
    assert rows[1].weight == 5.92


def test_parse_skips_rows_without_weight():
    rows = _parse_country_rows(COUNTRY_HTML)
    assert len(rows) == 2


def test_parse_empty_document():
    assert _parse_country_rows("<html><body></body></html>") == []


def test_parse_comma_decimal():
    html = COUNTRY_HTML.replace("5.92%", "5,92%")
    rows = _parse_country_rows(html)
    assert rows[1].weight == 5.92