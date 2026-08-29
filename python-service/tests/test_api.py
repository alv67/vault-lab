import pytest
import requests
from fastapi.testclient import TestClient

from app.main import app
from app.schemas import Exposure, ExposureRow

client = TestClient(app)


def test_healthz(monkeypatch):
    response = client.get("/healthz")
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


def test_get_exposure_ok(monkeypatch):
    def fake_fetch(isin):
        assert isin == "IE00B3RBWM25"
        return Exposure(
            countries=[
                ExposureRow(name="United States", weight=58.85),
                ExposureRow(name="Japan", weight=5.92),
            ],
            sectors=[],
        )

    monkeypatch.setattr("app.main.scraper.fetch_exposure", fake_fetch)
    response = client.get("/api/v1/etf/IE00B3RBWM25/exposure")
    assert response.status_code == 200
    body = response.json()
    assert body["countries"][0]["name"] == "United States"
    assert body["countries"][1]["weight"] == 5.92
    assert body["sectors"] == []


def test_get_exposure_upstream_error(monkeypatch):
    def fake_fetch(isin):
        raise requests.RequestException("connection refused")

    monkeypatch.setattr("app.main.scraper.fetch_exposure", fake_fetch)
    response = client.get("/api/v1/etf/IE00B3RBWM25/exposure")
    assert response.status_code == 502
    assert "detail" in response.json()


def test_get_exposure_empty(monkeypatch):
    def fake_fetch(isin):
        return Exposure(countries=[], sectors=[])

    monkeypatch.setattr("app.main.scraper.fetch_exposure", fake_fetch)
    response = client.get("/api/v1/etf/IE00B3RBWM25/exposure")
    assert response.status_code == 502


def test_get_holdings_stub(monkeypatch):
    response = client.get("/api/v1/etf/IE00B3RBWM25/holdings")
    assert response.status_code == 200
    body = response.json()
    assert body["isin"] == "IE00B3RBWM25"
    assert body["holdings"] == []