package price

import (
	"fmt"
	"net/http"
)

// FetchIssue describes a symbol (or currency) that could not be refreshed.
type FetchIssue struct {
	Symbol  string `json:"symbol"`
	Code    string `json:"code"` // "rate_limited" | "http_<status>" | "error"
	Message string `json:"message"`
}

// RefreshReport is the outcome of a price refresh run.
type RefreshReport struct {
	Refreshed   []string     `json:"refreshed"`
	Issues      []FetchIssue `json:"issues"`
	RateLimited bool         `json:"rate_limited"`
}

// issueCode maps an error to a stable machine-readable code: rate limiting for
// 401/403/429 responses, "http_<status>" for other HTTP statuses, "error"
// otherwise.
func issueCode(err error) string {
	switch s := statusOf(err); {
	case s == http.StatusUnauthorized || s == http.StatusForbidden || s == http.StatusTooManyRequests:
		return "rate_limited"
	case s > 0:
		return fmt.Sprintf("http_%d", s)
	default:
		return "error"
	}
}
