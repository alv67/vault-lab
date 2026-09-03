/**
 * Canonical list of ISO-3166 alpha-2 country codes used by the backend for
 * zero-filled exposure. The array defines the display order.
 */
export const CANONICAL_COUNTRIES: string[] = [
  // North America
  'CA', 'US',
  // Latin America & Caribbean
  'AR', 'BO', 'BR', 'CL', 'CO', 'CR', 'CU', 'DO', 'EC', 'GT', 'HN', 'MX',
  'NI', 'PA', 'PE', 'PY', 'SV', 'UY', 'VE',
  // Western Europe
  'AT', 'BE', 'CH', 'CY', 'DE', 'DK', 'ES', 'FI', 'FR', 'GB', 'GR', 'IE',
  'IS', 'IT', 'LU', 'MT', 'NL', 'NO', 'PT', 'SE',
  // Eastern Europe
  'BG', 'CZ', 'EE', 'HR', 'HU', 'LT', 'LV', 'PL', 'RO', 'RS', 'RU', 'SI',
  'SK', 'TR', 'UA',
  // Middle East & Africa
  'AE', 'BH', 'DZ', 'EG', 'IL', 'JO', 'KE', 'KW', 'LB', 'MA', 'NG', 'OM',
  'QA', 'SA', 'TN', 'ZA',
  // Oceania
  'AU', 'HK', 'JP', 'NZ', 'SG',
  // Asia
  'BD', 'CN', 'ID', 'IN', 'KR', 'LK', 'MY', 'PH', 'PK', 'TH', 'TW', 'VN',
]

/**
 * Full ISO-3166 alpha-2 → display name map for all canonical countries.
 */
export const COUNTRY_NAMES: Record<string, string> = {
  AE: 'United Arab Emirates',
  AR: 'Argentina',
  AT: 'Austria',
  AU: 'Australia',
  BD: 'Bangladesh',
  BE: 'Belgium',
  BH: 'Bahrain',
  BO: 'Bolivia',
  BR: 'Brazil',
  BG: 'Bulgaria',
  CA: 'Canada',
  CH: 'Switzerland',
  CL: 'Chile',
  CN: 'China',
  CO: 'Colombia',
  CR: 'Costa Rica',
  CU: 'Cuba',
  CY: 'Cyprus',
  CZ: 'Czech Republic',
  DE: 'Germany',
  DK: 'Denmark',
  DO: 'Dominican Republic',
  DZ: 'Algeria',
  EC: 'Ecuador',
  EE: 'Estonia',
  EG: 'Egypt',
  ES: 'Spain',
  FI: 'Finland',
  FR: 'France',
  GB: 'United Kingdom',
  GR: 'Greece',
  GT: 'Guatemala',
  HK: 'Hong Kong',
  HR: 'Croatia',
  HN: 'Honduras',
  HU: 'Hungary',
  ID: 'Indonesia',
  IE: 'Ireland',
  IL: 'Israel',
  IN: 'India',
  IS: 'Iceland',
  IT: 'Italy',
  JP: 'Japan',
  JO: 'Jordan',
  KE: 'Kenya',
  KR: 'South Korea',
  KW: 'Kuwait',
  LB: 'Lebanon',
  LK: 'Sri Lanka',
  LT: 'Lithuania',
  LU: 'Luxembourg',
  LV: 'Latvia',
  MA: 'Morocco',
  MT: 'Malta',
  MX: 'Mexico',
  MY: 'Malaysia',
  NG: 'Nigeria',
  NI: 'Nicaragua',
  NL: 'Netherlands',
  NO: 'Norway',
  NZ: 'New Zealand',
  OM: 'Oman',
  PA: 'Panama',
  PE: 'Peru',
  PH: 'Philippines',
  PK: 'Pakistan',
  PL: 'Poland',
  PT: 'Portugal',
  PY: 'Paraguay',
  QA: 'Qatar',
  RO: 'Romania',
  RS: 'Serbia',
  RU: 'Russia',
  SA: 'Saudi Arabia',
  SE: 'Sweden',
  SG: 'Singapore',
  SI: 'Slovenia',
  SK: 'Slovakia',
  SV: 'El Salvador',
  TH: 'Thailand',
  TN: 'Tunisia',
  TR: 'Turkey',
  TW: 'Taiwan',
  UA: 'Ukraine',
  US: 'United States',
  UY: 'Uruguay',
  VE: 'Venezuela',
  VN: 'Vietnam',
  ZA: 'South Africa',
}

/** Lookup friendly name for an ISO code; falls back to the code itself. */
export function countryDisplayName(code: string): string {
  return COUNTRY_NAMES[code] ?? code
}
