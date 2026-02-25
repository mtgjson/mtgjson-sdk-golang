package mtgjsonsdk

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mtgjson/mtgjson-sdk-go/booster"
	"github.com/mtgjson/mtgjson-sdk-go/db"
	"github.com/mtgjson/mtgjson-sdk-go/models"
	"github.com/mtgjson/mtgjson-sdk-go/queries"
)

// SDK is the main entry point for querying MTGJSON card data.
// It auto-downloads Parquet data from the MTGJSON CDN and provides
// a typed, queryable Go API for the full dataset.
type SDK struct {
	conn  *db.Connection
	cache *db.CacheManager

	cards       *queries.CardQuery
	sets        *queries.SetQuery
	tokens      *queries.TokenQuery
	legalities  *queries.LegalityQuery
	identifiers *queries.IdentifierQuery
	prices      *queries.PriceQuery
	decks       *queries.DeckQuery
	enums       *queries.EnumQuery
	skus        *queries.SkuQuery
	sealed      *queries.SealedQuery
	booster     *booster.BoosterSimulator
}

// New creates a new SDK instance with the given options.
func New(opts ...Option) (*SDK, error) {
	cfg := db.DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	cache, err := db.NewCacheManager(cfg)
	if err != nil {
		return nil, err
	}
	conn, err := db.NewConnection(cache)
	if err != nil {
		cache.Close()
		return nil, err
	}
	return &SDK{
		conn:  conn,
		cache: cache,
	}, nil
}

// Close releases all resources (DuckDB connection and HTTP client).
func (s *SDK) Close() error {
	s.cache.Close()
	return s.conn.Close()
}

// Cards returns the card query interface.
func (s *SDK) Cards() *queries.CardQuery {
	if s.cards == nil {
		s.cards = queries.NewCardQuery(s.conn)
	}
	return s.cards
}

// Sets returns the set query interface.
func (s *SDK) Sets() *queries.SetQuery {
	if s.sets == nil {
		s.sets = queries.NewSetQuery(s.conn)
	}
	return s.sets
}

// Tokens returns the token query interface.
func (s *SDK) Tokens() *queries.TokenQuery {
	if s.tokens == nil {
		s.tokens = queries.NewTokenQuery(s.conn)
	}
	return s.tokens
}

// Legalities returns the legality query interface.
func (s *SDK) Legalities() *queries.LegalityQuery {
	if s.legalities == nil {
		s.legalities = queries.NewLegalityQuery(s.conn)
	}
	return s.legalities
}

// Identifiers returns the identifier cross-reference query interface.
func (s *SDK) Identifiers() *queries.IdentifierQuery {
	if s.identifiers == nil {
		s.identifiers = queries.NewIdentifierQuery(s.conn)
	}
	return s.identifiers
}

// Prices returns the price query interface.
func (s *SDK) Prices() *queries.PriceQuery {
	if s.prices == nil {
		s.prices = queries.NewPriceQuery(s.conn)
	}
	return s.prices
}

// Decks returns the deck query interface.
func (s *SDK) Decks() *queries.DeckQuery {
	if s.decks == nil {
		s.decks = queries.NewDeckQuery(s.cache)
	}
	return s.decks
}

// Enums returns the enum query interface.
func (s *SDK) Enums() *queries.EnumQuery {
	if s.enums == nil {
		s.enums = queries.NewEnumQuery(s.cache)
	}
	return s.enums
}

// Skus returns the TCGPlayer SKU query interface.
func (s *SDK) Skus() *queries.SkuQuery {
	if s.skus == nil {
		s.skus = queries.NewSkuQuery(s.conn)
	}
	return s.skus
}

// Sealed returns the sealed product query interface.
func (s *SDK) Sealed() *queries.SealedQuery {
	if s.sealed == nil {
		s.sealed = queries.NewSealedQuery(s.conn)
	}
	return s.sealed
}

// Booster returns the booster simulator interface.
func (s *SDK) Booster() *booster.BoosterSimulator {
	if s.booster == nil {
		s.booster = booster.NewBoosterSimulator(s.conn)
	}
	return s.booster
}

// Meta returns MTGJSON build metadata (version and date).
func (s *SDK) Meta(ctx context.Context) (models.Meta, error) {
	data, err := s.cache.LoadJSON(ctx, "meta")
	if err != nil {
		return models.Meta{}, err
	}
	var meta models.Meta
	if d, ok := data["data"].(map[string]any); ok {
		if v, ok := d["version"].(string); ok {
			meta.Version = v
		}
		if v, ok := d["date"].(string); ok {
			meta.Date = v
		}
	}
	return meta, nil
}

// Views returns the names of all currently registered DuckDB views/tables.
func (s *SDK) Views() []string {
	return s.conn.Views()
}

// SQL executes raw SQL against the DuckDB database.
func (s *SDK) SQL(ctx context.Context, query string, params ...any) ([]map[string]any, error) {
	return s.conn.Execute(ctx, query, params...)
}

// Refresh checks for new MTGJSON data and resets internal state if stale.
// Returns true if data was stale and state was reset.
func (s *SDK) Refresh(ctx context.Context) (bool, error) {
	if !s.cache.IsStale(ctx) {
		return false, nil
	}
	s.conn.ClearViews()
	s.cache.ResetRemoteVersion()
	s.cards = nil
	s.sets = nil
	s.tokens = nil
	s.legalities = nil
	s.identifiers = nil
	s.prices = nil
	s.decks = nil
	s.enums = nil
	s.skus = nil
	s.sealed = nil
	s.booster = nil
	return true, nil
}

// ExportDB exports all loaded data to a persistent DuckDB file.
func (s *SDK) ExportDB(ctx context.Context, path string) error {
	pathStr := filepath.ToSlash(path)
	os.Remove(path)

	_, err := s.conn.Raw().ExecContext(ctx, fmt.Sprintf("ATTACH '%s' AS export_db", pathStr))
	if err != nil {
		return fmt.Errorf("mtgjson: attach export db: %w", err)
	}
	defer func() {
		s.conn.Raw().ExecContext(ctx, "DETACH export_db")
	}()

	for _, viewName := range s.Views() {
		_, err := s.conn.Raw().ExecContext(ctx, fmt.Sprintf(
			"CREATE TABLE export_db.%s AS SELECT * FROM %s", viewName, viewName,
		))
		if err != nil {
			return fmt.Errorf("mtgjson: export table %s: %w", viewName, err)
		}
	}
	return nil
}

// Connection returns the underlying Connection for advanced usage.
func (s *SDK) Connection() *db.Connection {
	return s.conn
}

// EnsureViews registers one or more views, downloading data if needed.
// This is useful before calling SQL() to ensure the required tables exist.
func (s *SDK) EnsureViews(ctx context.Context, names ...string) error {
	return s.conn.EnsureViews(ctx, names...)
}

// String returns a human-readable representation.
func (s *SDK) String() string {
	return fmt.Sprintf("SDK(cache_dir=%s)", s.cache.CacheDir)
}
