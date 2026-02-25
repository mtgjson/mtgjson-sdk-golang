package queries

import (
	"context"
	"fmt"
	"strings"

	"github.com/mtgjson/mtgjson-sdk-go/db"
	"github.com/mtgjson/mtgjson-sdk-go/models"
)

// PriceQuery provides methods to query card price data.
// Prices come from AllPricesToday.parquet, registered as a DuckDB view.
type PriceQuery struct {
	conn *db.Connection
}

func NewPriceQuery(conn *db.Connection) *PriceQuery {
	return &PriceQuery{conn: conn}
}

func (q *PriceQuery) ensure(ctx context.Context) {
	_ = q.conn.EnsureViews(ctx, "all_prices_today")
}

func (q *PriceQuery) ensureHistory(ctx context.Context) {
	_ = q.conn.EnsureViews(ctx, "all_prices")
}

// Get returns full price data for a card UUID as a nested map.
// Returns nil if no price data exists.
func (q *PriceQuery) Get(ctx context.Context, uuid string) (map[string]any, error) {
	q.ensure(ctx)
	if !q.conn.HasView("all_prices_today") {
		return nil, nil
	}
	rows, err := q.conn.Execute(ctx,
		"SELECT * FROM all_prices_today WHERE uuid = $1 ORDER BY source, provider, price_type, finish, date",
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
		cat, _ := r["price_type"].(string)
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
	q.ensure(ctx)
	if !q.conn.HasView("all_prices_today") {
		return nil, nil
	}
	cfg := &priceFilter{}
	for _, opt := range opts {
		opt(cfg)
	}

	parts := []string{
		"SELECT * FROM all_prices_today",
		"WHERE uuid = $1",
		"AND date = (SELECT MAX(p2.date) FROM all_prices_today p2 WHERE p2.uuid = $1)",
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
	if cfg.priceType != "" {
		parts = append(parts, fmt.Sprintf("AND price_type = $%d", idx))
		params = append(params, cfg.priceType)
	}

	return q.conn.Execute(ctx, strings.Join(parts, " "), params...)
}

// History returns price history for a card UUID.
func (q *PriceQuery) History(ctx context.Context, uuid string, opts ...PriceHistoryOption) ([]map[string]any, error) {
	q.ensureHistory(ctx)
	if !q.conn.HasView("all_prices") {
		return nil, nil
	}
	cfg := &priceHistoryConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	parts := []string{"SELECT * FROM all_prices WHERE uuid = $1"}
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
	if cfg.priceType != "" {
		parts = append(parts, fmt.Sprintf("AND price_type = $%d", idx))
		params = append(params, cfg.priceType)
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
func (q *PriceQuery) PriceTrend(ctx context.Context, uuid string, opts ...PriceFilterOption) (*models.PriceTrend, error) {
	q.ensureHistory(ctx)
	if !q.conn.HasView("all_prices") {
		return nil, nil
	}
	cfg := &priceFilter{priceType: "retail"}
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
		"FROM all_prices",
		"WHERE uuid = $1 AND price_type = $2",
	}
	params := []any{uuid, cfg.priceType}
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
	dp := db.ScalarToInt(rows[0]["data_points"])
	if dp == 0 {
		return nil, nil
	}
	return &models.PriceTrend{
		MinPrice:   db.ToFloat64(rows[0]["min_price"]),
		MaxPrice:   db.ToFloat64(rows[0]["max_price"]),
		AvgPrice:   db.ToFloat64(rows[0]["avg_price"]),
		FirstDate:  db.ToDateStr(rows[0]["first_date"]),
		LastDate:   db.ToDateStr(rows[0]["last_date"]),
		DataPoints: int64(dp),
	}, nil
}

// CheapestPrinting finds the cheapest printing of a card by name.
func (q *PriceQuery) CheapestPrinting(ctx context.Context, name string, opts ...PriceFilterOption) (map[string]any, error) {
	q.ensure(ctx)
	if err := q.conn.EnsureViews(ctx, "cards"); err != nil {
		return nil, err
	}
	if !q.conn.HasView("all_prices_today") {
		return nil, nil
	}
	cfg := &priceFilter{provider: "tcgplayer", finish: "normal", priceType: "retail"}
	for _, opt := range opts {
		opt(cfg)
	}

	sql := "SELECT c.uuid, c.setCode, c.number, p.price, p.date " +
		"FROM cards c " +
		"JOIN all_prices_today p ON c.uuid = p.uuid " +
		"WHERE c.name = $1 AND p.provider = $2 " +
		"AND p.finish = $3 AND p.price_type = $4 " +
		"AND p.date = (SELECT MAX(p2.date) FROM all_prices_today p2 " +
		"WHERE p2.uuid = c.uuid AND p2.provider = $2 " +
		"AND p2.finish = $3 AND p2.price_type = $4) " +
		"ORDER BY p.price ASC " +
		"LIMIT 1"
	rows, err := q.conn.Execute(ctx, sql, name, cfg.provider, cfg.finish, cfg.priceType)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

// CheapestPrintings finds the cheapest available printing of each card.
func (q *PriceQuery) CheapestPrintings(ctx context.Context, opts ...PriceListOption) ([]models.PricePrinting, error) {
	q.ensure(ctx)
	if err := q.conn.EnsureViews(ctx, "cards"); err != nil {
		return nil, err
	}
	if !q.conn.HasView("all_prices_today") {
		return nil, nil
	}
	cfg := &priceListConfig{provider: "tcgplayer", finish: "normal", priceType: "retail", limit: 100}
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
			"JOIN all_prices_today p ON c.uuid = p.uuid "+
			"WHERE p.provider = $1 AND p.finish = $2 AND p.price_type = $3 "+
			"AND p.date = (SELECT MAX(date) FROM all_prices_today) "+
			"GROUP BY c.name "+
			"ORDER BY min_price ASC "+
			"LIMIT %d OFFSET %d", cfg.limit, cfg.offset)

	var result []models.PricePrinting
	if err := q.conn.ExecuteInto(ctx, &result, sql, cfg.provider, cfg.finish, cfg.priceType); err != nil {
		return nil, err
	}
	return result, nil
}

// MostExpensivePrintings finds the most expensive printing of each card.
func (q *PriceQuery) MostExpensivePrintings(ctx context.Context, opts ...PriceListOption) ([]models.ExpensivePrinting, error) {
	q.ensure(ctx)
	if err := q.conn.EnsureViews(ctx, "cards"); err != nil {
		return nil, err
	}
	if !q.conn.HasView("all_prices_today") {
		return nil, nil
	}
	cfg := &priceListConfig{provider: "tcgplayer", finish: "normal", priceType: "retail", limit: 100}
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
			"JOIN all_prices_today p ON c.uuid = p.uuid "+
			"WHERE p.provider = $1 AND p.finish = $2 AND p.price_type = $3 "+
			"AND p.date = (SELECT MAX(date) FROM all_prices_today) "+
			"GROUP BY c.name "+
			"ORDER BY max_price DESC "+
			"LIMIT %d OFFSET %d", cfg.limit, cfg.offset)

	var result []models.ExpensivePrinting
	if err := q.conn.ExecuteInto(ctx, &result, sql, cfg.provider, cfg.finish, cfg.priceType); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Functional option types ---

type priceFilter struct {
	provider  string
	finish    string
	priceType string
}

// PriceFilterOption configures price query filters.
type PriceFilterOption func(*priceFilter)

// WithPriceProvider filters by price provider (e.g. "tcgplayer", "cardmarket").
func WithPriceProvider(provider string) PriceFilterOption {
	return func(c *priceFilter) { c.provider = provider }
}

// WithPriceFinish filters by card finish (e.g. "normal", "foil", "etched").
func WithPriceFinish(finish string) PriceFilterOption {
	return func(c *priceFilter) { c.finish = finish }
}

// WithPriceType filters by price type ("retail" or "buylist").
func WithPriceType(priceType string) PriceFilterOption {
	return func(c *priceFilter) { c.priceType = priceType }
}

type priceHistoryConfig struct {
	provider  string
	finish    string
	priceType string
	dateFrom  string
	dateTo    string
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

// WithHistoryPriceType filters history by price type.
func WithHistoryPriceType(priceType string) PriceHistoryOption {
	return func(c *priceHistoryConfig) { c.priceType = priceType }
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
	provider  string
	finish    string
	priceType string
	limit     int
	offset    int
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

// WithListPriceType sets the price type for list queries.
func WithListPriceType(priceType string) PriceListOption {
	return func(c *priceListConfig) { c.priceType = priceType }
}

// WithListLimit sets the max results for list queries.
func WithListLimit(limit int) PriceListOption {
	return func(c *priceListConfig) { c.limit = limit }
}

// WithListOffset sets the offset for list query pagination.
func WithListOffset(offset int) PriceListOption {
	return func(c *priceListConfig) { c.offset = offset }
}

// --- Helper ---

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
