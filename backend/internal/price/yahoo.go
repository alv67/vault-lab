package price

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/amelamela/vault-lab/internal/model"
	"github.com/amelamela/vault-lab/internal/repository"
)

type YahooFetcher struct {
	repos           *repository.Repository
	client          *http.Client
	refreshInterval time.Duration
	historyCooldown map[uuid.UUID]time.Time
	splitCooldown   map[uuid.UUID]time.Time
	mu              sync.Mutex
}

func NewYahooFetcher(repos *repository.Repository, refreshInterval time.Duration) *YahooFetcher {
	return &YahooFetcher{
		repos:           repos,
		refreshInterval: refreshInterval,
		historyCooldown: map[uuid.UUID]time.Time{},
		splitCooldown:   map[uuid.UUID]time.Time{},
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type yahooChartQuote struct {
	Open   []*float64 `json:"open"`
	High   []*float64 `json:"high"`
	Low    []*float64 `json:"low"`
	Close  []*float64 `json:"close"`
	Volume []*int64   `json:"volume"`
}

type yahooChartResult struct {
	Meta struct {
		Symbol           string `json:"symbol"`
		Currency         string `json:"currency"`
		ExchangeName     string `json:"exchangeName"`
		FullExchangeName string `json:"fullExchangeName"`
		InstrumentType   string `json:"instrumentType"`
		LongName         string `json:"longName"`
		ShortName        string `json:"shortName"`
	} `json:"meta"`
	Timestamp  []int64           `json:"timestamp"`
	Indicators struct {
		Quote []yahooChartQuote `json:"quote"`
	} `json:"indicators"`
	Events *yahooChartEvents `json:"events"`
}

type yahooChartEvents struct {
	Splits map[string]yahooSplitEvent `json:"splits"`
}

type yahooSplitEvent struct {
	Date        int64   `json:"date"`
	Numerator   float64 `json:"numerator"`
	Denominator float64 `json:"denominator"`
	SplitRatio  string  `json:"splitRatio"`
}

type yahooChartResponse struct {
	Chart struct {
		Result []yahooChartResult `json:"result"`
		Error  interface{}        `json:"error"`
	} `json:"chart"`
}

type historyBar struct {
	Date  time.Time
	Close decimal.Decimal
}

type HistoryAsset struct {
	ID     uuid.UUID
	Ticker string
	From   time.Time
}

// FetchAll refreshes prices for every asset in the DB that is stale and keeps
// the USD->X FX rates fresh. Used by the worker.
func (f *YahooFetcher) FetchAll(ctx context.Context) error {
	assets, err := f.repos.Asset.List(ctx)
	if err != nil {
		return fmt.Errorf("list assets: %w", err)
	}

	refreshed, err := f.RefreshStale(ctx, assets)
	if err != nil {
		return err
	}
	if len(refreshed) > 0 {
		log.Info().Strs("symbols", refreshed).Msg("prices refreshed")
	} else {
		log.Info().Msg("no stale prices to refresh")
	}

	if err := f.RefreshFX(ctx); err != nil {
		log.Warn().Err(err).Msg("fx refresh failed")
	}
	return nil
}

// RefreshFX keeps USD->X rates for every currency used by assets up to date,
// using the same staleness/throttle logic as asset prices.
func (f *YahooFetcher) RefreshFX(ctx context.Context) error {
	currencies, err := f.repos.Asset.Currencies(ctx)
	if err != nil {
		return fmt.Errorf("list currencies: %w", err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now().UTC()
	for _, quote := range currencies {
		if quote == "USD" || quote == "" {
			continue
		}
		fetchedAt, err := f.repos.FX.FetchedAt(ctx, quote)
		if err == nil && fetchedAt.Add(f.refreshInterval).After(now) {
			continue
		}

		rate, err := f.fetchFX(ctx, quote)
		if err != nil {
			log.Warn().Err(err).Str("currency", quote).Msg("fx fetch failed")
			continue
		}
		if err := f.repos.FX.Upsert(ctx, "USD", quote, rate); err != nil {
			log.Warn().Err(err).Str("currency", quote).Msg("fx save failed")
			continue
		}
		log.Info().Str("currency", quote).Str("rate", rate.String()).Msg("fx rate updated")
	}
	return nil
}

func (f *YahooFetcher) fetchFX(ctx context.Context, quote string) (decimal.Decimal, error) {
	chart, err := f.fetchChart(ctx, "USD"+quote+"=X")
	if err != nil {
		return decimal.Zero, err
	}
	if len(chart.Indicators.Quote) == 0 {
		return decimal.Zero, fmt.Errorf("no quote indicators")
	}
	quoteData := chart.Indicators.Quote[0]

	idx := -1
	for i := len(chart.Timestamp) - 1; i >= 0; i-- {
		if i < len(quoteData.Close) && quoteData.Close[i] != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		return decimal.Zero, fmt.Errorf("no fx rate available")
	}
	return decimal.NewFromFloat(*quoteData.Close[idx]), nil
}

// fetchChart performs one Yahoo chart request and returns the parsed result.
func (f *YahooFetcher) fetchChart(ctx context.Context, ticker string) (*yahooChartResult, error) {
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=5d", ticker)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("yahoo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("yahoo returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result yahooChartResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w (body: %s)", err, string(body))
	}
	if result.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo error: %v", result.Chart.Error)
	}
	if len(result.Chart.Result) == 0 {
		return nil, fmt.Errorf("no chart data")
	}

	chart := &result.Chart.Result[0]
	return chart, nil
}

func (f *YahooFetcher) fetchChartRange(ctx context.Context, ticker string, from, to time.Time) ([]historyBar, error) {
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&period1=%d&period2=%d", ticker, from.Unix(), to.Unix())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("yahoo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("yahoo returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result yahooChartResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w (body: %s)", err, string(body))
	}
	if result.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo error: %v", result.Chart.Error)
	}
	if len(result.Chart.Result) == 0 {
		return nil, fmt.Errorf("no chart data")
	}

	chart := &result.Chart.Result[0]
	var bars []historyBar
	if len(chart.Indicators.Quote) > 0 {
		quote := chart.Indicators.Quote[0]
		for i, ts := range chart.Timestamp {
			if i < len(quote.Close) && quote.Close[i] != nil {
				bars = append(bars, historyBar{
					Date:  priceDate(ts),
					Close: decimal.NewFromFloat(*quote.Close[i]),
				})
			}
		}
	}
	if len(bars) == 0 {
		return nil, fmt.Errorf("no history bars")
	}
	return bars, nil
}

// EnsureHistory backfills daily closes from Yahoo for the given assets so the
// portfolio history series has market values from the first transaction to now.
func (f *YahooFetcher) EnsureHistory(ctx context.Context, assets []HistoryAsset) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now().UTC()
	for _, a := range assets {
		earliest, latest, err := f.repos.Price.MinMaxDate(ctx, a.ID)
		if err != nil {
			log.Warn().Err(err).Str("symbol", a.Ticker).Msg("history range failed")
			continue
		}

		var bars []historyBar
		switch {
		case latest == nil || earliest == nil || earliest.After(a.From.Add(30*24*time.Hour)):
			if last, ok := f.historyCooldown[a.ID]; ok && last.Add(f.refreshInterval).After(now) {
				continue
			}
			bars, err = f.fetchChartRange(ctx, a.Ticker, a.From, now)
			f.historyCooldown[a.ID] = now
		case latest.Add(f.refreshInterval).Before(now):
			bars, err = f.fetchChartRange(ctx, a.Ticker, *latest, now)
		default:
			continue
		}
		if err != nil {
			log.Warn().Err(err).Str("symbol", a.Ticker).Msg("history fetch failed")
			continue
		}
		for _, b := range bars {
			if _, err := f.repos.Price.Create(ctx, &model.Price{AssetID: a.ID, Date: b.Date, Close: b.Close, Source: "yahoo"}); err != nil {
				log.Warn().Err(err).Str("symbol", a.Ticker).Str("date", b.Date.Format("2006-01-02")).Msg("history save failed")
			}
		}
	}
	return nil
}

func (f *YahooFetcher) fetchSplits(ctx context.Context, ticker string) ([]model.Split, error) {
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1wk&events=split&range=max", ticker)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("yahoo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("yahoo returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result yahooChartResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w (body: %s)", err, string(body))
	}
	if result.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo error: %v", result.Chart.Error)
	}
	if len(result.Chart.Result) == 0 || result.Chart.Result[0].Events == nil {
		return []model.Split{}, nil
	}

	var splits []model.Split
	for _, e := range result.Chart.Result[0].Events.Splits {
		if e.Numerator == 0 || e.Denominator == 0 {
			continue
		}
		splits = append(splits, model.Split{
			Date:        time.Unix(e.Date, 0).UTC().Truncate(24 * time.Hour),
			Numerator:   decimal.NewFromFloat(e.Numerator),
			Denominator: decimal.NewFromFloat(e.Denominator),
		})
	}
	return splits, nil
}

// EnsureSplits fetches the stock split events for each asset and upserts them,
// re-checking the list on every call outside the per-asset cooldown. Upsert is
// idempotent on (asset_id, date) so re-fetching is safe.
func (f *YahooFetcher) EnsureSplits(ctx context.Context, assets []*model.Asset) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now().UTC()
	for _, a := range assets {
		if last, ok := f.splitCooldown[a.ID]; ok && last.Add(f.refreshInterval).After(now) {
			continue
		}
		splits, err := f.fetchSplits(ctx, a.Ticker)
		f.splitCooldown[a.ID] = now
		if err != nil {
			log.Warn().Err(err).Str("symbol", a.Ticker).Msg("split fetch failed")
			continue
		}
		for _, sp := range splits {
			sp.AssetID = a.ID
			if err := f.repos.Split.Upsert(ctx, &sp); err != nil {
				log.Warn().Err(err).Str("symbol", a.Ticker).Str("date", sp.Date.Format("2006-01-02")).Msg("split save failed")
			}
		}
	}
	return nil
}

// RefreshStaleForPortfolio refreshes prices for assets currently held in a
// portfolio. Used by POST /prices/refresh on page open.
func (f *YahooFetcher) RefreshStaleForPortfolio(ctx context.Context, portfolioID uuid.UUID) ([]string, error) {
	assets, err := f.repos.Portfolio.HeldAssets(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("list held assets: %w", err)
	}
	return f.RefreshStale(ctx, assets)
}

// RefreshStale fetches quotes from Yahoo only for assets whose stored close is
// stale (old close date) and that haven't been fetched recently. This keeps the
// number of calls to Yahoo Finance as low as possible.
func (f *YahooFetcher) RefreshStale(ctx context.Context, assets []*model.Asset) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(assets) == 0 {
		return nil, nil
	}

	ids := make([]uuid.UUID, 0, len(assets))
	for _, a := range assets {
		ids = append(ids, a.ID)
	}

	latest, err := f.repos.Price.FindLatestForAssets(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("find latest prices: %w", err)
	}

	now := time.Now().UTC()
	expected := latestExpectedClose(now)
	var stale []*model.Asset
	for _, a := range assets {
		lp, ok := latest[a.ID]
		if !ok {
			// No price stored yet: fetch.
			stale = append(stale, a)
			continue
		}
		if a.PriceFetchedAt != nil && a.PriceFetchedAt.Add(f.refreshInterval).After(now) {
			// Fetched recently: skip to respect throttling.
			continue
		}
		if !lp.Date.Before(expected) {
			// Stored close is already fresh enough.
			continue
		}
		stale = append(stale, a)
	}

	if len(stale) == 0 {
		return nil, nil
	}

	log.Info().
		Strs("symbols", tickers(stale)).
		Str("expected_close", expected.Format("2006-01-02")).
		Msg("refreshing stale prices")

	refreshed := f.fetchQuotes(ctx, stale)
	if len(refreshed) == 0 {
		return nil, nil
	}

	if err := f.repos.Asset.MarkPricesFetched(ctx, assetIDs(stale), now); err != nil {
		log.Warn().Err(err).Msg("failed to mark prices as fetched")
	}
	return refreshed, nil
}

// fetchQuotes fetches the latest OHLCV close for every asset, one Yahoo chart
// request per symbol, and persists them. Returns the symbols that were updated.
func (f *YahooFetcher) fetchQuotes(ctx context.Context, assets []*model.Asset) []string {
	var refreshed []string
	for i, a := range assets {
		if err := f.fetchQuote(ctx, a); err != nil {
			log.Warn().Err(err).Str("symbol", a.Ticker).Msg("quote fetch failed")
		} else {
			refreshed = append(refreshed, a.Ticker)
		}

		if i < len(assets)-1 {
			// Rate limiting between requests.
			time.Sleep(500 * time.Millisecond)
		}
	}
	return refreshed
}

func (f *YahooFetcher) fetchQuote(ctx context.Context, asset *model.Asset) error {
	chart, err := f.fetchChart(ctx, asset.Ticker)
	if err != nil {
		return err
	}
	if len(chart.Indicators.Quote) == 0 {
		return fmt.Errorf("no quote indicators")
	}
	quote := chart.Indicators.Quote[0]

	// Find the last complete daily bar (all values present).
	idx := -1
	for i := len(chart.Timestamp) - 1; i >= 0; i-- {
		if i < len(quote.Open) && i < len(quote.High) && i < len(quote.Low) &&
			i < len(quote.Close) && i < len(quote.Volume) &&
			quote.Open[i] != nil && quote.High[i] != nil && quote.Low[i] != nil &&
			quote.Close[i] != nil && quote.Volume[i] != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("no complete daily bar")
	}

	price := &model.Price{
		AssetID: asset.ID,
		Date:    priceDate(chart.Timestamp[idx]),
		Open:    decimal.NewFromFloat(*quote.Open[idx]),
		High:    decimal.NewFromFloat(*quote.High[idx]),
		Low:     decimal.NewFromFloat(*quote.Low[idx]),
		Close:   decimal.NewFromFloat(*quote.Close[idx]),
		Volume:  *quote.Volume[idx],
		Source:  "yahoo",
	}

	if _, err := f.repos.Price.Create(ctx, price); err != nil {
		return fmt.Errorf("save price: %w", err)
	}
	return nil
}

// priceDate returns the UTC calendar day the quote refers to.
func priceDate(marketTime int64) time.Time {
	return time.Unix(marketTime, 0).UTC().Truncate(24 * time.Hour)
}

// latestExpectedClose returns the most recent weekday (Mon-Fri) at UTC midnight.
// Used as the "latest close we should already have stored" reference point.
func latestExpectedClose(now time.Time) time.Time {
	d := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
		d = d.AddDate(0, 0, -1)
	}
	return d
}

func assetIDs(assets []*model.Asset) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(assets))
	for _, a := range assets {
		ids = append(ids, a.ID)
	}
	return ids
}

func tickers(assets []*model.Asset) []string {
	ts := make([]string, 0, len(assets))
	for _, a := range assets {
		ts = append(ts, a.Ticker)
	}
	return ts
}
