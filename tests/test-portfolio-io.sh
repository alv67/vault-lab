#!/usr/bin/env bash
#
# test-portfolio-io.sh — Regressione export/import portafoglio (compatibilità old file)
#
# Contesto: i documenti esportati da versioni precedenti dell'app non contengono
# i campi opzionali aggiunti in seguito (price_source, asset_class). L'import
# NON deve fallire con 500 (violazione della CHECK assets_price_source_check),
# ma deve importare il documento riempendo i valori mancanti con default sensati.
# L'export successivo deve riportare i campi (round-trip).
#
# Precondizioni:
#   - Stack di test isolato avviato (make test-e2e), backend raggiungibile.
#
# Uso:
#   ./tests/test-portfolio-io.sh [BASE_URL]
#   BASE_URL default: http://localhost:8080
#
# Esito: stampa PASS/FAIL per ogni check e exit code non-zero in caso di fallimento.
#
set -uo pipefail

BASE_URL="${1:-http://localhost:8080}"
API="$BASE_URL/api/v1"

PASS=0
FAIL=0

note()  { printf '\n\033[1m==> %s\033[0m\n' "$1"; }
ok()    { printf '  \033[32mPASS\033[0m  %s\n' "$1"; PASS=$((PASS+1)); }
bad()   { printf '  \033[31mFAIL\033[0m  %s\n' "$1"; FAIL=$((FAIL+1)); }
die()   { printf '\n\033[31mERRORE: %s\033[0m\n' "$1"; exit 1; }

# --- helper: login o registrazione e recupero access_token
login() { # $1 email  $2 password
  local body
  body=$(curl -s -X POST "$API/auth/login" \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"$1\",\"password\":\"$2\"}")
  jq -r '.access_token // empty' <<<"$body"
}

register() { # $1 email  $2 password  $3 name
  curl -s -X POST "$API/auth/register" \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"$1\",\"password\":\"$2\",\"name\":\"$3\"}" >/dev/null
}

# richiedi POST/GET mantenendo il codice HTTP sull'ultima riga
http() { # $1 method  $2 path  [$3 body]
  local method="$1" path="$2" body="${3:-}"
  if [ -n "$body" ]; then
    curl -s -X "$method" "$API$path" -H "Authorization: Bearer $TOK" \
      -H 'Content-Type: application/json' -d "$body" -w $'\n%{http_code}'
  else
    curl -s -X "$method" "$API$path" -H "Authorization: Bearer $TOK" -w $'\n%{http_code}'
  fi
}
code_of() { printf '%s' "${1##*$'\n'}"; }
body_of() { printf '%s' "${1%$'\n'*}"; }

# --- preparazione utente -------------------------------------------------------
TS=$(date +%s)
EMAIL="portfoli_io-$TS@test.local"
PW="password123"
TICKER_A="IOTST${TS}A"
TICKER_B="IOTST${TS}B"

note "Preparazione utente test"
register "$EMAIL" "$PW" "Portfolio IO"
TOK=$(login "$EMAIL" "$PW")
[ -n "$TOK" ] && ok "login utente import" || die "login utente fallito"

# --- import di un documento "vecchio formato" ----------------------------------
# price_source e asset_class volutamente ASSENTI (come nei file delle vecchie
# versioni); tipo ETF con valuta e ISIN compilati.
note "Import documento vecchio formato (senza price_source/asset_class)"
DOC=$(jq -n \
  --arg ta "$TICKER_A" --arg tb "$TICKER_B" '
  {
    document: {
      version: 1,
      exported_at: "2024-01-01T00:00:00Z",
      portfolio: {name: "Recovered legacy", description: "old file", currency: "USD"},
      assets: [
        {ticker: $ta, name: "Legacy Test ETF Alpha", type: "etf", currency: "EUR", isin: "IE00TESTAL01"},
        {ticker: $tb, name: "Legacy Test ETF Beta",  type: "etf", currency: "USD"}
      ],
      transactions: [
        {date: "2024-01-05T00:00:00Z", type: "buy", asset_ticker: $ta, quantity: "10",   price: "105.5", fees: "1.5"},
        {date: "2024-02-07T00:00:00Z", type: "buy", asset_ticker: $tb, quantity: "5",    price: "220.10"}
      ]
    },
    mode: "new",
    name: "Recovered legacy"
  }')
RESP=$(http POST /portfolios/import "$DOC")
CODE=$(code_of "$RESP"); IMPORT_BODY=$(body_of "$RESP")
if [ "$CODE" = "201" ]; then
  ok "import old-format -> 201 (era 500: regressione fixata)"
else
  bad "import old-format -> $CODE (atteso 201): $IMPORT_BODY"
fi
PID=$(jq -r '.id // empty' <<<"$IMPORT_BODY")
[ -n "$PID" ] || die "nessun portafoglio creato dall'import"

# --- verifica portafoglio e transazioni ----------------------------------------
note "Verifica portafoglio importato"
RESP=$(http GET "/portfolios/$PID")
CODE=$(code_of "$RESP"); PF=$(body_of "$RESP")
[ "$CODE" = "200" ] && ok "GET /portfolios/{id} -> 200" || bad "GET /portfolios/{id} -> $CODE (atteso 200)"
[ "$(jq -r '.name' <<<"$PF")" = "Recovered legacy" ] && ok "nome portafoglio corretto" || bad "nome portafoglio errato: $PF"

RESP=$(http GET "/portfolios/$PID/transactions")
CODE=$(code_of "$RESP"); TXS=$(body_of "$RESP")
[ "$CODE" = "200" ] && ok "GET transactions -> 200" || bad "GET transactions -> $CODE (atteso 200)"
# Since EPIC I.9 the endpoint returns a page envelope
# ({transactions,total,limit,offset}), not a bare array: count the rows inside.
TX_COUNT=$(jq '.transactions | length' <<<"$TXS")
[ "$TX_COUNT" = "2" ] && ok "2 transazioni importate" || bad "transazioni importate: $TX_COUNT (atteso 2)"

# --- default applicati agli asset creati ----------------------------------------
note "Asset importati ricevono i default (price_source=yahoo)"
RESP=$(http GET /assets)
ASSETS=$(body_of "$RESP")
for t in "$TICKER_A" "$TICKER_B"; do
  ps=$(jq -r --arg t "$t" '[.[] | select(.ticker==$t)][0].price_source // empty' <<<"$ASSETS")
  [ "$ps" = "yahoo" ] && ok "$t: price_source default 'yahoo'" || bad "$t: price_source='$ps' (atteso 'yahoo')"
  ac=$(jq -r --arg t "$t" '[.[] | select(.ticker==$t)][0].asset_class // empty' <<<"$ASSETS")
  [ -n "$ac" ] && ok "$t: asset_class valorizzato ('$ac')" || bad "$t: asset_class mancante"
  tp=$(jq -r --arg t "$t" '[.[] | select(.ticker==$t)][0].type // empty' <<<"$ASSETS")
  [ "$tp" = "etf" ] && ok "$t: type dal documento ('etf')" || bad "$t: type='$tp' (atteso 'etf')"
done

# --- round-trip: l'export deve riportare i campi --------------------------------
note "Round-trip: export contiene price_source"
RESP=$(http GET "/portfolios/$PID/export")
CODE=$(code_of "$RESP"); EXPORT=$(body_of "$RESP")
[ "$CODE" = "200" ] && ok "GET /portfolios/{id}/export -> 200" || bad "export -> $CODE (atteso 200)"
[ "$(jq -r '.version' <<<"$EXPORT")" = "1" ] && ok "export version = 1" || bad "export version errata: $(jq -r '.version' <<<"$EXPORT")"
for t in "$TICKER_A" "$TICKER_B"; do
  ps=$(jq -r --arg t "$t" '[.assets[] | select(.ticker==$t)][0].price_source // empty' <<<"$EXPORT")
  [ "$ps" = "yahoo" ] && ok "export $t: price_source presente ('yahoo')" || bad "export $t: price_source='$ps' (atteso 'yahoo')"
done

# --- gate di versione: documento più nuovo del supportato ------------------------
note "Import con version > 1 rifiutato con 400 chiaro"
FUTURE=$(jq -n --arg ta "$TICKER_A" '
  {
    document: {
      version: 99,
      exported_at: "2030-01-01T00:00:00Z",
      portfolio: {name: "From the future", currency: "USD"},
      assets: [{ticker: $ta, name: "Future ETF", type: "etf", currency: "USD"}],
      transactions: []
    },
    mode: "new"
  }')
RESP=$(http POST /portfolios/import "$FUTURE")
CODE=$(code_of "$RESP"); FUT_BODY=$(body_of "$RESP")
[ "$CODE" = "400" ] && ok "version 99 -> 400 (non 500)" || bad "version 99 -> $CODE (atteso 400)"
jq -e '.error | test("unsupported export version")' <<<"$FUT_BODY" >/dev/null 2>&1 \
  && ok "messaggio 'unsupported export version' presente" || bad "messaggio versione assente: $FUT_BODY"

# --- riepilogo -------------------------------------------------------------------
printf '\n\033[1mRisultato: %d PASS, %d FAIL\033[0m\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
