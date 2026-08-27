// Package geo defines the canonical macro-regions and GICS sectors used to
// classify asset exposure, plus the mappings from country codes and provider
// sector names to those canonical values.
package geo

import "strings"

// Regions is the canonical list of macro-regions in display order.
var Regions = []string{
	"North America",
	"Latin America",
	"Europe Developed",
	"Europe Emerging",
	"Africa / Middle East",
	"Asia Developed",
	"Asia Emerging",
	"Other / Not Classified",
}

// GICSSectors is the canonical list of GICS sectors.
var GICSSectors = []string{
	"Energy",
	"Materials",
	"Industrials",
	"Consumer Discretionary",
	"Consumer Staples",
	"Health Care",
	"Financials",
	"Information Technology",
	"Communication Services",
	"Utilities",
	"Real Estate",
}

var regionByCountry = map[string]string{
	// North America
	"US": "North America",
	"CA": "North America",
	// Latin America
	"BR": "Latin America",
	"MX": "Latin America",
	"AR": "Latin America",
	"CL": "Latin America",
	"CO": "Latin America",
	"PE": "Latin America",
	"VE": "Latin America",
	"UY": "Latin America",
	"PY": "Latin America",
	"EC": "Latin America",
	"BO": "Latin America",
	"CR": "Latin America",
	"PA": "Latin America",
	"DO": "Latin America",
	"GT": "Latin America",
	"HN": "Latin America",
	"SV": "Latin America",
	"NI": "Latin America",
	"CU": "Latin America",
	// Europe Developed
	"GB": "Europe Developed",
	"FR": "Europe Developed",
	"DE": "Europe Developed",
	"IT": "Europe Developed",
	"ES": "Europe Developed",
	"CH": "Europe Developed",
	"NL": "Europe Developed",
	"BE": "Europe Developed",
	"AT": "Europe Developed",
	"SE": "Europe Developed",
	"DK": "Europe Developed",
	"NO": "Europe Developed",
	"FI": "Europe Developed",
	"IE": "Europe Developed",
	"PT": "Europe Developed",
	"LU": "Europe Developed",
	"IS": "Europe Developed",
	"GR": "Europe Developed",
	"MT": "Europe Developed",
	"CY": "Europe Developed",
	// Europe Emerging
	"PL": "Europe Emerging",
	"CZ": "Europe Emerging",
	"HU": "Europe Emerging",
	"TR": "Europe Emerging",
	"RU": "Europe Emerging",
	"RO": "Europe Emerging",
	"BG": "Europe Emerging",
	"SK": "Europe Emerging",
	"SI": "Europe Emerging",
	"HR": "Europe Emerging",
	"RS": "Europe Emerging",
	"UA": "Europe Emerging",
	"LT": "Europe Emerging",
	"LV": "Europe Emerging",
	"EE": "Europe Emerging",
	// Africa / Middle East
	"ZA": "Africa / Middle East",
	"NG": "Africa / Middle East",
	"EG": "Africa / Middle East",
	"SA": "Africa / Middle East",
	"AE": "Africa / Middle East",
	"IL": "Africa / Middle East",
	"KE": "Africa / Middle East",
	"MA": "Africa / Middle East",
	"DZ": "Africa / Middle East",
	"TN": "Africa / Middle East",
	"QA": "Africa / Middle East",
	"KW": "Africa / Middle East",
	"OM": "Africa / Middle East",
	"BH": "Africa / Middle East",
	"JO": "Africa / Middle East",
	"LB": "Africa / Middle East",
	// Asia Developed
	"JP": "Asia Developed",
	"AU": "Asia Developed",
	"NZ": "Asia Developed",
	"SG": "Asia Developed",
	"HK": "Asia Developed",
	// Asia Emerging
	"CN": "Asia Emerging",
	"IN": "Asia Emerging",
	"KR": "Asia Emerging",
	"TW": "Asia Emerging",
	"TH": "Asia Emerging",
	"ID": "Asia Emerging",
	"MY": "Asia Emerging",
	"PH": "Asia Emerging",
	"VN": "Asia Emerging",
	"LK": "Asia Emerging",
	"PK": "Asia Emerging",
	"BD": "Asia Emerging",
}

// RegionForCountry maps an ISO alpha-2 country code to its macro-region.
// Matching is case-insensitive; unknown or empty codes fall back to
// "Other / Not Classified".
func RegionForCountry(country string) string {
	if region, ok := regionByCountry[strings.ToUpper(strings.TrimSpace(country))]; ok {
		return region
	}
	return "Other / Not Classified"
}

var sectorAliases = map[string]string{
	"Technology":             "Information Technology",
	"Information Technology": "Information Technology",
	"Financial Services":     "Financials",
	"Financials":             "Financials",
	"Healthcare":             "Health Care",
	"Health Care":            "Health Care",
	"Consumer Cyclical":      "Consumer Discretionary",
	"Consumer Discretionary": "Consumer Discretionary",
	"Consumer Defensive":     "Consumer Staples",
	"Consumer Staples":       "Consumer Staples",
	"Communication Services": "Communication Services",
	"Basic Materials":        "Materials",
	"Materials":              "Materials",
}

// NormalizeSector maps a provider sector name to the canonical GICS name.
// Names absent from the alias map are returned unchanged.
func NormalizeSector(sector string) string {
	s := strings.TrimSpace(sector)
	if s == "" {
		return ""
	}
	if canonical, ok := sectorAliases[s]; ok {
		return canonical
	}
	return s
}

var sectorKeyToGICS = map[string]string{
	"realestate":             "Real Estate",
	"consumer_cyclical":      "Consumer Discretionary",
	"basic_materials":        "Materials",
	"consumer_defensive":     "Consumer Staples",
	"technology":             "Information Technology",
	"communication_services": "Communication Services",
	"financial_services":     "Financials",
	"utilities":              "Utilities",
	"industrials":            "Industrials",
	"energy":                 "Energy",
	"healthcare":             "Health Care",
}

// SectorKeyToGICS maps a Yahoo sectorWeightings key to the canonical GICS
// sector. Unknown keys return "".
func SectorKeyToGICS(key string) string {
	return sectorKeyToGICS[strings.ToLower(strings.TrimSpace(key))]
}
