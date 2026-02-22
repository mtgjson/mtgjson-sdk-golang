package queries

import (
	"context"
	"fmt"
	"strings"

	"github.com/mtgjson/mtgjson-sdk-go/db"
	"github.com/mtgjson/mtgjson-sdk-go/models"
)

// ListSetsParams contains filters for listing sets.
type ListSetsParams struct {
	SetType string
	Name    string
	Limit   int // 0 means default (1000)
	Offset  int
}

// SearchSetsParams contains filters for searching sets.
type SearchSetsParams struct {
	Name        string
	SetType     string
	Block       string
	ReleaseYear *int
	Limit       int // 0 means default (100)
}

// SetQuery provides methods to search and retrieve set metadata.
type SetQuery struct {
	conn *db.Connection
}

func NewSetQuery(conn *db.Connection) *SetQuery {
	return &SetQuery{conn: conn}
}

// Get returns a set by its code (case-insensitive), or nil if not found.
func (q *SetQuery) Get(ctx context.Context, code string) (*models.SetList, error) {
	if err := q.conn.EnsureViews(ctx, "sets"); err != nil {
		return nil, err
	}
	var sets []models.SetList
	if err := q.conn.ExecuteInto(ctx, &sets, "SELECT * FROM sets WHERE code = $1", strings.ToUpper(code)); err != nil {
		return nil, err
	}
	if len(sets) == 0 {
		return nil, nil
	}
	return &sets[0], nil
}

// List returns sets with optional filters, ordered by release date descending.
func (q *SetQuery) List(ctx context.Context, p ListSetsParams) ([]models.SetList, error) {
	if err := q.conn.EnsureViews(ctx, "sets"); err != nil {
		return nil, err
	}
	b := db.NewSQLBuilder("sets")
	if p.SetType != "" {
		b.WhereEq("type", p.SetType)
	}
	if p.Name != "" {
		if containsWildcard(p.Name) {
			b.WhereLike("name", p.Name)
		} else {
			b.WhereEq("name", p.Name)
		}
	}
	b.OrderBy("releaseDate DESC")
	limit := p.Limit
	if limit <= 0 {
		limit = 1000
	}
	b.Limit(limit).Offset(p.Offset)
	sql, params := b.Build()
	var sets []models.SetList
	if err := q.conn.ExecuteInto(ctx, &sets, sql, params...); err != nil {
		return nil, err
	}
	return sets, nil
}

// Search searches sets with flexible filters.
func (q *SetQuery) Search(ctx context.Context, p SearchSetsParams) ([]models.SetList, error) {
	if err := q.conn.EnsureViews(ctx, "sets"); err != nil {
		return nil, err
	}
	b := db.NewSQLBuilder("sets")
	if p.Name != "" {
		b.WhereLike("name", "%"+p.Name+"%")
	}
	if p.SetType != "" {
		b.WhereEq("type", p.SetType)
	}
	if p.Block != "" {
		b.WhereLike("block", "%"+p.Block+"%")
	}
	if p.ReleaseYear != nil {
		idx := b.AddParam(*p.ReleaseYear)
		b.AddWhere(fmt.Sprintf("EXTRACT(YEAR FROM CAST(releaseDate AS DATE)) = $%d", idx))
	}
	b.OrderBy("releaseDate DESC")
	limit := p.Limit
	if limit <= 0 {
		limit = 100
	}
	b.Limit(limit)
	sql, params := b.Build()
	var sets []models.SetList
	if err := q.conn.ExecuteInto(ctx, &sets, sql, params...); err != nil {
		return nil, err
	}
	return sets, nil
}

// GetFinancialSummary returns aggregate price statistics for a set.
// Requires price data to be loaded (via PriceQuery). Returns nil if unavailable.
func (q *SetQuery) GetFinancialSummary(ctx context.Context, setCode string, opts ...FinancialSummaryOption) (*models.FinancialSummary, error) {
	if err := q.conn.EnsureViews(ctx, "cards"); err != nil {
		return nil, err
	}
	if !q.conn.HasView("prices_today") {
		return nil, nil
	}
	cfg := financialSummaryDefaults()
	for _, opt := range opts {
		opt(&cfg)
	}
	sql := `SELECT
		COUNT(DISTINCT c.uuid) AS card_count,
		ROUND(SUM(p.price), 2) AS total_value,
		ROUND(AVG(p.price), 2) AS avg_value,
		MIN(p.price) AS min_value,
		MAX(p.price) AS max_value,
		MAX(p.date) AS date
	FROM cards c
	JOIN prices_today p ON c.uuid = p.uuid
	WHERE c.setCode = $1
	  AND p.provider = $2
	  AND p.currency = $3
	  AND p.finish = $4
	  AND p.category = $5
	  AND p.date = (SELECT MAX(p2.date) FROM prices_today p2)`

	var results []models.FinancialSummary
	if err := q.conn.ExecuteInto(ctx, &results, sql,
		strings.ToUpper(setCode), cfg.provider, cfg.currency, cfg.finish, cfg.category,
	); err != nil {
		return nil, err
	}
	if len(results) == 0 || results[0].CardCount == 0 {
		return nil, nil
	}
	return &results[0], nil
}

// Count returns the total number of sets.
func (q *SetQuery) Count(ctx context.Context) (int, error) {
	if err := q.conn.EnsureViews(ctx, "sets"); err != nil {
		return 0, err
	}
	val, err := q.conn.ExecuteScalar(ctx, "SELECT COUNT(*) FROM sets")
	if err != nil {
		return 0, err
	}
	return db.ScalarToInt(val), nil
}

// FinancialSummaryOption configures GetFinancialSummary.
type FinancialSummaryOption func(*financialSummaryCfg)

type financialSummaryCfg struct {
	provider string
	currency string
	finish   string
	category string
}

func financialSummaryDefaults() financialSummaryCfg {
	return financialSummaryCfg{
		provider: "tcgplayer",
		currency: "USD",
		finish:   "normal",
		category: "retail",
	}
}

// WithProvider sets the price provider for financial summary.
func WithProvider(provider string) FinancialSummaryOption {
	return func(c *financialSummaryCfg) { c.provider = provider }
}

// WithCurrency sets the currency for financial summary.
func WithCurrency(currency string) FinancialSummaryOption {
	return func(c *financialSummaryCfg) { c.currency = currency }
}

// WithFinish sets the card finish for financial summary.
func WithFinish(finish string) FinancialSummaryOption {
	return func(c *financialSummaryCfg) { c.finish = finish }
}

// WithCategory sets the price category for financial summary.
func WithCategory(category string) FinancialSummaryOption {
	return func(c *financialSummaryCfg) { c.category = category }
}
