import requests
from fastapi import FastAPI, HTTPException

from . import scraper
from .schemas import Exposure, Holdings

app = FastAPI(title="VaultLab ETF metadata service", version="1.0.0")


@app.get("/healthz")
def healthz():
    return {"status": "ok"}


@app.get("/api/v1/etf/{isin}/exposure", response_model=Exposure)
def get_exposure(isin: str):
    try:
        exposure = scraper.fetch_exposure(isin)
    except requests.RequestException as exc:
        raise HTTPException(status_code=502, detail=f"upstream fetch failed: {exc}")
    except Exception as exc:
        raise HTTPException(status_code=502, detail=f"exposure parse failed: {exc}")
    if not exposure.countries:
        raise HTTPException(status_code=502, detail="no country data found for ISIN")
    return exposure


@app.get("/api/v1/etf/{isin}/holdings", response_model=Holdings)
def get_holdings(isin: str):
    return Holdings(isin=isin, holdings=[])