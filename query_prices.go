package mtgjson

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PriceQuery provides methods to query card price data.
// Prices come from AllPricesToday.json.gz, flattened and loaded into DuckDB.
type PriceQuery struct {
	conn   *Connection
	cache  *CacheManager
	loaded bool
}

func newPriceQuery(conn *Connection, cache *CacheManager) *PriceQuery {
	return &PriceQuery{conn: conn, cache: cache}
}

func (q *PriceQuery) ensure(ctx context.Context) error {
	if q.loaded {
		return nil
	}
	if q.conn.HasView("prices_today") {
		q.loaded = true
		return nil
	}
	path, err := q.cache.EnsureJSON(ctx, "all_prices_today")
	if err != nil {
		slog.Warn("Price data not available", "error", err)
		q.loaded = true
		return nil
	}
	if err := loadPricesToDuckDB(ctx, path, q.conn); err != nil {
		return err
	}
	q.loaded = true
	return nil
}

// Get returns full price data for a card UUID as a nested map.
// Returns nil if no price data exists.
func (q *PriceQuery) Get(ctx context.Context, uuid string) (map[string]any, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}
	if !q.conn.HasView("prices_today") {
		return nil, nil
	}
	rows, err := q.conn.Execute(ctx,
		"SELECT * FROM prices_today WHERE uuid = $1 ORDER BY source, provider, category, finish, date",
		uuid)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	// Reconstruct nested structure from flat rows
	result := make(map[string]any)
	for _, r := range rows {
		src, _ := r["source"].(string)
		prov, _ := r["provider"].(string)
		cat, _ := r["category"].(string)
		fin, _ := r["finish"].(string)
		date, _ := r["date"].(string)
		price := r["price"]
		currency, _ := r["currency"].(string)
		if currency == "" {
			currency = "USD"
		}

		srcMap := ensureNestedMap(result, src)
		provMap := ensureNestedMap(srcMap, prov)
		provMap["currency"] = currency
		catMap := ensureNestedMap(provMap, cat)
		finMap := ensureNestedMap(catMap, fin)
		finMap[date] = price
	}
	return result, nil
}

// Today returns the latest prices for a card UUID.
func (q *PriceQuery) Today(ctx context.Context, uuid string, opts ...PriceFilterOption) ([]map[string]any, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}
	if !q.conn.HasView("prices_today") {
		return nil, nil
	}
	cfg := &priceFilter{}
	for _, opt := range opts {
		opt(cfg)
	}

	parts := []string{
		"SELECT * FROM prices_today",
		"WHERE uuid = $1",
		"AND date = (SELECT MAX(p2.date) FROM prices_today p2 WHERE p2.uuid = $1)",
	}
	params := []any{uuid}
	idx := 2

	if cfg.provider != "" {
		parts = append(parts, fmt.Sprintf("AND provider = $%d", idx))
		params = append(params, cfg.provider)
		idx++
	}
	if cfg.finish != "" {
		parts = append(parts, fmt.Sprintf("AND finish = $%d", idx))
		params = append(params, cfg.finish)
		idx++
	}
	if cfg.category != "" {
		parts = append(parts, fmt.Sprintf("AND category = $%d", idx))
		params = append(params, cfg.category)
	}

	return q.conn.Execute(ctx, strings.Join(parts, " "), params...)
}

// History returns price history for a card UUID.
func (q *PriceQuery) History(ctx context.Context, uuid string, opts ...PriceHistoryOption) ([]map[string]any, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}
	if !q.conn.HasView("prices_today") {
		return nil, nil
	}
	cfg := &priceHistoryConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	parts := []string{"SELECT * FROM prices_today WHERE uuid = $1"}
	params := []any{uuid}
	idx := 2

	if cfg.provider != "" {
		parts = append(parts, fmt.Sprintf("AND provider = $%d", idx))
		params = append(params, cfg.provider)
		idx++
	}
	if cfg.finish != "" {
		parts = append(parts, fmt.Sprintf("AND finish = $%d", idx))
		params = append(params, cfg.finish)
		idx++
	}
	if cfg.category != "" {
		parts = append(parts, fmt.Sprintf("AND category = $%d", idx))
		params = append(params, cfg.category)
		idx++
	}
	if cfg.dateFrom != "" {
		parts = append(parts, fmt.Sprintf("AND date >= CAST($%d AS DATE)", idx))
		params = append(params, cfg.dateFrom)
		idx++
	}
	if cfg.dateTo != "" {
		parts = append(parts, fmt.Sprintf("AND date <= CAST($%d AS DATE)", idx))
		params = append(params, cfg.dateTo)
	}
	parts = append(parts, "ORDER BY date ASC")

	return q.conn.Execute(ctx, strings.Join(parts, " "), params...)
}

// PriceTrend returns price trend statistics for a card.
func (q *PriceQuery) PriceTrend(ctx context.Context, uuid string, opts ...PriceFilterOption) (*PriceTrend, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}
	if !q.conn.HasView("prices_today") {
		return nil, nil
	}
	cfg := &priceFilter{category: "retail"}
	for _, opt := range opts {
		opt(cfg)
	}

	parts := []string{
		"SELECT",
		"  MIN(price) AS min_price,",
		"  MAX(price) AS max_price,",
		"  ROUND(AVG(price), 2) AS avg_price,",
		"  MIN(date) AS first_date,",
		"  MAX(date) AS last_date,",
		"  COUNT(*) AS data_points",
		"FROM prices_today",
		"WHERE uuid = $1 AND category = $2",
	}
	params := []any{uuid, cfg.category}
	idx := 3

	if cfg.provider != "" {
		parts = append(parts, fmt.Sprintf("AND provider = $%d", idx))
		params = append(params, cfg.provider)
		idx++
	}
	if cfg.finish != "" {
		parts = append(parts, fmt.Sprintf("AND finish = $%d", idx))
		params = append(params, cfg.finish)
	}

	rows, err := q.conn.Execute(ctx, strings.Join(parts, " "), params...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	dp := scalarToInt(rows[0]["data_points"])
	if dp == 0 {
		return nil, nil
	}
	return &PriceTrend{
		MinPrice:   toFloat64(rows[0]["min_price"]),
		MaxPrice:   toFloat64(rows[0]["max_price"]),
		AvgPrice:   toFloat64(rows[0]["avg_price"]),
		FirstDate:  toDateStr(rows[0]["first_date"]),
		LastDate:   toDateStr(rows[0]["last_date"]),
		DataPoints: int64(dp),
	}, nil
}

// CheapestPrinting finds the cheapest printing of a card by name.
func (q *PriceQuery) CheapestPrinting(ctx context.Context, name string, opts ...PriceFilterOption) (map[string]any, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}
	if err := q.conn.EnsureViews(ctx, "cards"); err != nil {
		return nil, err
	}
	if !q.conn.HasView("prices_today") {
		return nil, nil
	}
	cfg := &priceFilter{provider: "tcgplayer", finish: "normal", category: "retail"}
	for _, opt := range opts {
		opt(cfg)
	}

	sql := "SELECT c.uuid, c.setCode, c.number, p.price, p.date " +
		"FROM cards c " +
		"JOIN prices_today p ON c.uuid = p.uuid " +
		"WHERE c.name = $1 AND p.provider = $2 " +
		"AND p.finish = $3 AND p.category = $4 " +
		"AND p.date = (SELECT MAX(p2.date) FROM prices_today p2 " +
		"WHERE p2.uuid = c.uuid AND p2.provider = $2 " +
		"AND p2.finish = $3 AND p2.category = $4) " +
		"ORDER BY p.price ASC " +
		"LIMIT 1"
	rows, err := q.conn.Execute(ctx, sql, name, cfg.provider, cfg.finish, cfg.category)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

// CheapestPrintings finds the cheapest available printing of each card.
func (q *PriceQuery) CheapestPrintings(ctx context.Context, opts ...PriceListOption) ([]PricePrinting, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}
	if err := q.conn.EnsureViews(ctx, "cards"); err != nil {
		return nil, err
	}
	if !q.conn.HasView("prices_today") {
		return nil, nil
	}
	cfg := &priceListConfig{provider: "tcgplayer", finish: "normal", category: "retail", limit: 100}
	for _, opt := range opts {
		opt(cfg)
	}

	sql := fmt.Sprintf(
		"SELECT c.name, "+
			"  arg_min(c.setCode, p.price) AS cheapest_set, "+
			"  arg_min(c.number, p.price) AS cheapest_number, "+
			"  arg_min(c.uuid, p.price) AS cheapest_uuid, "+
			"  MIN(p.price) AS min_price "+
			"FROM cards c "+
			"JOIN prices_today p ON c.uuid = p.uuid "+
			"WHERE p.provider = $1 AND p.finish = $2 AND p.category = $3 "+
			"AND p.date = (SELECT MAX(date) FROM prices_today) "+
			"GROUP BY c.name "+
			"ORDER BY min_price ASC "+
			"LIMIT %d OFFSET %d", cfg.limit, cfg.offset)

	var result []PricePrinting
	if err := q.conn.ExecuteInto(ctx, &result, sql, cfg.provider, cfg.finish, cfg.category); err != nil {
		return nil, err
	}
	return result, nil
}

// MostExpensivePrintings finds the most expensive printing of each card.
func (q *PriceQuery) MostExpensivePrintings(ctx context.Context, opts ...PriceListOption) ([]ExpensivePrinting, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}
	if err := q.conn.EnsureViews(ctx, "cards"); err != nil {
		return nil, err
	}
	if !q.conn.HasView("prices_today") {
		return nil, nil
	}
	cfg := &priceListConfig{provider: "tcgplayer", finish: "normal", category: "retail", limit: 100}
	for _, opt := range opts {
		opt(cfg)
	}

	sql := fmt.Sprintf(
		"SELECT c.name, "+
			"  arg_max(c.setCode, p.price) AS priciest_set, "+
			"  arg_max(c.number, p.price) AS priciest_number, "+
			"  arg_max(c.uuid, p.price) AS priciest_uuid, "+
			"  MAX(p.price) AS max_price "+
			"FROM cards c "+
			"JOIN prices_today p ON c.uuid = p.uuid "+
			"WHERE p.provider = $1 AND p.finish = $2 AND p.category = $3 "+
			"AND p.date = (SELECT MAX(date) FROM prices_today) "+
			"GROUP BY c.name "+
			"ORDER BY max_price DESC "+
			"LIMIT %d OFFSET %d", cfg.limit, cfg.offset)

	var result []ExpensivePrinting
	if err := q.conn.ExecuteInto(ctx, &result, sql, cfg.provider, cfg.finish, cfg.category); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Functional option types ---

type priceFilter struct {
	provider string
	finish   string
	category string
}

// PriceFilterOption configures price query filters.
type PriceFilterOption func(*priceFilter)

// WithProvider filters by price provider (e.g. "tcgplayer", "cardmarket").
func WithPriceProvider(provider string) PriceFilterOption {
	return func(c *priceFilter) { c.provider = provider }
}

// WithFinish filters by card finish (e.g. "normal", "foil", "etched").
func WithPriceFinish(finish string) PriceFilterOption {
	return func(c *priceFilter) { c.finish = finish }
}

// WithCategory filters by price category ("retail" or "buylist").
func WithPriceCategory(category string) PriceFilterOption {
	return func(c *priceFilter) { c.category = category }
}

type priceHistoryConfig struct {
	provider string
	finish   string
	category string
	dateFrom string
	dateTo   string
}

// PriceHistoryOption configures price history query filters.
type PriceHistoryOption func(*priceHistoryConfig)

// WithHistoryProvider filters history by provider.
func WithHistoryProvider(provider string) PriceHistoryOption {
	return func(c *priceHistoryConfig) { c.provider = provider }
}

// WithHistoryFinish filters history by finish.
func WithHistoryFinish(finish string) PriceHistoryOption {
	return func(c *priceHistoryConfig) { c.finish = finish }
}

// WithHistoryCategory filters history by category.
func WithHistoryCategory(category string) PriceHistoryOption {
	return func(c *priceHistoryConfig) { c.category = category }
}

// WithDateFrom sets the start date filter (inclusive, YYYY-MM-DD).
func WithDateFrom(date string) PriceHistoryOption {
	return func(c *priceHistoryConfig) { c.dateFrom = date }
}

// WithDateTo sets the end date filter (inclusive, YYYY-MM-DD).
func WithDateTo(date string) PriceHistoryOption {
	return func(c *priceHistoryConfig) { c.dateTo = date }
}

type priceListConfig struct {
	provider string
	finish   string
	category string
	limit    int
	offset   int
}

// PriceListOption configures cheapest/most expensive printing queries.
type PriceListOption func(*priceListConfig)

// WithListProvider sets the provider for list queries.
func WithListProvider(provider string) PriceListOption {
	return func(c *priceListConfig) { c.provider = provider }
}

// WithListFinish sets the finish for list queries.
func WithListFinish(finish string) PriceListOption {
	return func(c *priceListConfig) { c.finish = finish }
}

// WithListCategory sets the category for list queries.
func WithListCategory(category string) PriceListOption {
	return func(c *priceListConfig) { c.category = category }
}

// WithListLimit sets the max results for list queries.
func WithListLimit(limit int) PriceListOption {
	return func(c *priceListConfig) { c.limit = limit }
}

// WithListOffset sets the offset for list query pagination.
func WithListOffset(offset int) PriceListOption {
	return func(c *priceListConfig) { c.offset = offset }
}

// --- Price data loading ---

func ensureNestedMap(parent map[string]any, key string) map[string]any {
	if v, ok := parent[key]; ok {
		if m, ok := v.(map[string]any); ok {
			return m
		}
	}
	m := make(map[string]any)
	parent[key] = m
	return m
}

func toDateStr(v any) string {
	switch d := v.(type) {
	case time.Time:
		return d.Format("2006-01-02")
	case string:
		return d
	default:
		return fmt.Sprint(v)
	}
}

func toFloat64(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case int:
		return float64(val)
	default:
		return 0
	}
}

// StreamFlattenPrices flattens nested price data to NDJSON rows written to w.
// Returns the number of rows written.
func StreamFlattenPrices(data map[string]any, w io.Writer) int {
	enc := json.NewEncoder(w)
	count := 0
	for uuid, formatsRaw := range data {
		formats, ok := formatsRaw.(map[string]any)
		if !ok {
			continue
		}
		for source, providersRaw := range formats { // paper, mtgo
			providers, ok := providersRaw.(map[string]any)
			if !ok {
				continue
			}
			for provider, priceDataRaw := range providers { // tcgplayer, cardmarket, etc.
				priceData, ok := priceDataRaw.(map[string]any)
				if !ok {
					continue
				}
				currency, _ := priceData["currency"].(string)
				if currency == "" {
					currency = "USD"
				}
				for _, categoryName := range []string{"buylist", "retail"} {
					categoryData, ok := priceData[categoryName].(map[string]any)
					if !ok {
						continue
					}
					for finish, datePricesRaw := range categoryData { // normal, foil, etched
						datePrices, ok := datePricesRaw.(map[string]any)
						if !ok {
							continue
						}
						for date, price := range datePrices {
							if price == nil {
								continue
							}
							row := map[string]any{
								"uuid":     uuid,
								"source":   source,
								"provider": provider,
								"currency": currency,
								"category": categoryName,
								"finish":   finish,
								"date":     date,
								"price":    toFloat64(price),
							}
							enc.Encode(row)
							count++
						}
					}
				}
			}
		}
	}
	return count
}

func loadPricesToDuckDB(ctx context.Context, path string, conn *Connection) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("mtgjson: open price file: %w", err)
	}
	defer f.Close()

	var reader io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		gr, err := gzip.NewReader(f)
		if err != nil {
			return fmt.Errorf("mtgjson: decompress price file: %w", err)
		}
		defer gr.Close()
		reader = gr
	}

	var raw map[string]any
	if err := json.NewDecoder(reader).Decode(&raw); err != nil {
		return fmt.Errorf("mtgjson: parse price JSON: %w", err)
	}

	data, _ := raw["data"].(map[string]any)
	if data == nil {
		return nil
	}

	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("mtgjson_prices_%d.ndjson", os.Getpid()))
	ndjson, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	count := StreamFlattenPrices(data, ndjson)
	ndjson.Close()

	if count > 0 {
		return conn.RegisterTableFromNdjson(ctx, "prices_today", tmpPath)
	}
	return nil
}
