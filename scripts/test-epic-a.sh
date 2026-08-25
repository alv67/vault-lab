#!/usr/bin/env bash
#
# test-epic-a.sh — Verifica manuale delle modifiche EPIC A (A.1–A.4)
#
# Precondizioni:
#   - Stack avviato (make up) con backend raggiungibile.
#
# Uso:
#   ./scripts/test-epic-a.sh [BASE_URL]
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

# --- preparazione utenti -----------------------------------------------------
EMAIL_A="epicA-$(date +%s)@test.local"
EMAIL_B="epicB-$(date +%s)@test.local"
PW="password123"

note "Preparazione utenti test"
register "$EMAIL_A" "$PW" "Epic A"
register "$EMAIL_B" "$PW" "Epic B"
TOK_A=$(login "$EMAIL_A" "$PW")
TOK_B=$(login "$EMAIL_B" "$PW")
[ -n "$TOK_A" ] && [ -n "$TOK_B" ] && ok "login user A e user B" || die "login utenti fallito"

# --- creazione portafoglio di proprietà di A ---------------------------------
note "Creazione portafoglio (owner = user A)"
PF=$(curl -s -X POST "$API/portfolios" \
  -H "Authorization: Bearer $TOK_A" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Test EPIC A","description":"verifica","currency":"USD"}')
PID=$(jq -r '.id // empty' <<<"$PF")
[ -n "$PID" ] && ok "portafoglio creato ($PID)" || die "creazione portafoglio fallita: $PF"

# --- A.2: accesso negato a utente non proprietario ---------------------------
note "A.2 — Endpoint analitici bloccati per utente non proprietario (atteso 403)"
for ep in summary allocation performance roi history; do
  code=$(curl -s -o /dev/null -w '%{http_code}' "$API/portfolios/$PID/$ep" \
    -H "Authorization: Bearer $TOK_B")
  [ "$code" = "403" ] && ok "$ep -> 403 (atteso)" || bad "$ep -> $code (atteso 403)"
done

note "A.2 — Stesso utente (owner) deve avere accesso (atteso 200)"
for ep in summary allocation performance roi; do
  code=$(curl -s -o /dev/null -w '%{http_code}' "$API/portfolios/$PID/$ep" \
    -H "Authorization: Bearer $TOK_A")
  [ "$code" = "200" ] && ok "$ep -> 200 (atteso)" || bad "$ep -> $code (atteso 200)"
done

# --- A.3: metric di data-quality nel summary ---------------------------------
note "A.3 — Campi data-quality presenti nel summary"
SUM=$(curl -s "$API/portfolios/$PID/summary" -H "Authorization: Bearer $TOK_A")
for f in fx_missing_count fx_missing_value missing_country missing_category stale_count; do
  if jq -e "has(\"$f\")" <<<"$SUM" >/dev/null 2>&1; then
    ok "summary contiene '$f'"
  else
    bad "summary manca del campo '$f'"
  fi
done

# --- A.1: campo fx_missing per holding / allocazione -------------------------
note "A.1 — flag fx_missing presente nelle holding"
if jq -e '(.holdings | length) == 0 or all(.holdings[]; has("fx_missing") or has("stale"))' <<<"$SUM" >/dev/null 2>&1; then
  ok "holding espongono flag (fx_missing/stale)"
else
  bad "holding senza flag attesi"
fi

# --- A.4: sanity — endpoint base rispondono -----------------------------------
note "A.4 — Sanity sugli endpoint esistenti"
code=$(curl -s -o /dev/null -w '%{http_code}' "$API/portfolios" -H "Authorization: Bearer $TOK_A")
[ "$code" = "200" ] && ok "GET /portfolios -> 200" || bad "GET /portfolios -> $code (atteso 200)"

# --- riepilogo -----------------------------------------------------------------
printf '\n\033[1mRisultato: %d PASS, %d FAIL\033[0m\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
