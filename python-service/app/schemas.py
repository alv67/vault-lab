from pydantic import BaseModel, Field


class ExposureRow(BaseModel):
    name: str
    weight: float = Field(ge=0, le=100)


class Exposure(BaseModel):
    isin: str = ""
    countries: list[ExposureRow] = []
    sectors: list[ExposureRow] = []
    regions: list[ExposureRow] = []


class Holdings(BaseModel):
    isin: str
    holdings: list[dict] = []


class EtfSearchResult(BaseModel):
    isin: str
    name: str = ""
    ticker: str = ""