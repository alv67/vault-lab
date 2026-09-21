package model

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TransactionFilter narrows the paginated transactions read. The zero value
// means "no filter". Type is the raw transaction type string (validated),
// From/To are date-only UTC midnights and inclusive, matched against the
// transaction date cast to a calendar date.
type TransactionFilter struct {
	Type    string
	AssetID *uuid.UUID
	From    *time.Time
	To      *time.Time
}

// ValidTransactionType reports whether s names one of the transaction types
// the schema CHECK constraint allows.
func ValidTransactionType(s string) bool {
	switch TransactionType(s) {
	case TxBuy, TxSell, TxDividend, TxSplit, TxFee:
		return true
	}
	return false
}

// TransactionDateLayout is the strict YYYY-MM-DD spelling accepted by the
// transaction list date filters.
const TransactionDateLayout = "2006-01-02"

// ParseTransactionFilter builds a TransactionFilter from endpoint query
// values. Every parameter is optional (absent or empty means unfiltered);
// any malformed value yields a static error naming the offending parameter
// so the handler can answer 400 without echoing user input.
func ParseTransactionFilter(values url.Values) (TransactionFilter, error) {
	var f TransactionFilter
	if raw := strings.TrimSpace(values.Get("type")); raw != "" {
		if !ValidTransactionType(raw) {
			return TransactionFilter{}, errors.New("invalid type")
		}
		f.Type = raw
	}
	if raw := strings.TrimSpace(values.Get("asset_id")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return TransactionFilter{}, errors.New("invalid asset_id")
		}
		f.AssetID = &id
	}
	if raw := strings.TrimSpace(values.Get("from")); raw != "" {
		d, err := ParseTransactionDate(raw)
		if err != nil {
			return TransactionFilter{}, fmt.Errorf("invalid from: %w", err)
		}
		f.From = &d
	}
	if raw := strings.TrimSpace(values.Get("to")); raw != "" {
		d, err := ParseTransactionDate(raw)
		if err != nil {
			return TransactionFilter{}, fmt.Errorf("invalid to: %w", err)
		}
		f.To = &d
	}
	return f, nil
}

// ParseTransactionDate parses a strict YYYY-MM-DD date into UTC midnight.
// The Format round-trip rejects lenient spellings like 2024-1-5 that
// time.Parse would otherwise accept.
func ParseTransactionDate(raw string) (time.Time, error) {
	d, err := time.Parse(TransactionDateLayout, raw)
	if err != nil {
		return time.Time{}, errors.New("date must be YYYY-MM-DD")
	}
	if d.Format(TransactionDateLayout) != raw {
		return time.Time{}, errors.New("date must be YYYY-MM-DD")
	}
	return d, nil
}
