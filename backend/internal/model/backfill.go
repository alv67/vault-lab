package model

// MetaBackfillReport reports the result of the asset-metadata backfill endpoint.
type MetaBackfillReport struct {
	Processed      int      `json:"processed"`
	UpdatedCountry int      `json:"updated_country"`
	UpdatedSector  int      `json:"updated_sector"`
	Failed         int      `json:"failed"`
	Errors         []string `json:"errors"`
}
