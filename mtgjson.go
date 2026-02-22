package mtgjson

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SDK is the main entry point for querying MTGJSON card data.
// It auto-downloads Parquet data from the MTGJSON CDN and provides
// a typed, queryable Go API for the full dataset.
type SDK struct {
	conn  *Connection
	cache *CacheManager

	cards       *CardQuery
	sets        *SetQuery
	tokens      *TokenQuery
	legalities  *LegalityQuery
	identifiers *IdentifierQuery
	prices      *PriceQuery
	decks       *DeckQuery
	enums       *EnumQuery
	skus        *SkuQuery
	sealed      *SealedQuery
	booster     *BoosterSimulator
}

// New creates a new SDK instance with the given options.
func New(opts ...Option) (*SDK, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	cache, err := newCacheManager(cfg)
	if err != nil {
		return nil, err
	}
	conn, err := NewConnection(cache)
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
func (s *SDK) Cards() *CardQuery {
	if s.cards == nil {
		s.cards = newCardQuery(s.conn)
	}
	return s.cards
}

// Sets returns the set query interface.
func (s *SDK) Sets() *SetQuery {
	if s.sets == nil {
		s.sets = newSetQuery(s.conn)
	}
	return s.sets
}

// Tokens returns the token query interface.
func (s *SDK) Tokens() *TokenQuery {
	if s.tokens == nil {
		s.tokens = newTokenQuery(s.conn)
	}
	return s.tokens
}

// Legalities returns the legality query interface.
func (s *SDK) Legalities() *LegalityQuery {
	if s.legalities == nil {
		s.legalities = newLegalityQuery(s.conn)
	}
	return s.legalities
}

// Identifiers returns the identifier cross-reference query interface.
func (s *SDK) Identifiers() *IdentifierQuery {
	if s.identifiers == nil {
		s.identifiers = newIdentifierQuery(s.conn)
	}
	return s.identifiers
}

// Prices returns the price query interface.
func (s *SDK) Prices() *PriceQuery {
	if s.prices == nil {
		s.prices = newPriceQuery(s.conn, s.cache)
	}
	return s.prices
}

// Decks returns the deck query interface.
func (s *SDK) Decks() *DeckQuery {
	if s.decks == nil {
		s.decks = newDeckQuery(s.cache)
	}
	return s.decks
}

// Enums returns the enum query interface.
func (s *SDK) Enums() *EnumQuery {
	if s.enums == nil {
		s.enums = newEnumQuery(s.cache)
	}
	return s.enums
}

// Skus returns the TCGPlayer SKU query interface.
func (s *SDK) Skus() *SkuQuery {
	if s.skus == nil {
		s.skus = newSkuQuery(s.conn, s.cache)
	}
	return s.skus
}

// Sealed returns the sealed product query interface.
func (s *SDK) Sealed() *SealedQuery {
	if s.sealed == nil {
		s.sealed = newSealedQuery(s.conn)
	}
	return s.sealed
}

// Booster returns the booster simulator interface.
func (s *SDK) Booster() *BoosterSimulator {
	if s.booster == nil {
		s.booster = newBoosterSimulator(s.conn)
	}
	return s.booster
}

// Meta returns MTGJSON build metadata (version and date).
func (s *SDK) Meta(ctx context.Context) (Meta, error) {
	data, err := s.cache.LoadJSON(ctx, "meta")
	if err != nil {
		return Meta{}, err
	}
	var meta Meta
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
	// Remove existing file
	os.Remove(path)

	_, err := s.conn.db.ExecContext(ctx, fmt.Sprintf("ATTACH '%s' AS export_db", pathStr))
	if err != nil {
		return fmt.Errorf("mtgjson: attach export db: %w", err)
	}
	defer func() {
		s.conn.db.ExecContext(ctx, "DETACH export_db")
	}()

	for _, viewName := range s.Views() {
		_, err := s.conn.db.ExecContext(ctx, fmt.Sprintf(
			"CREATE TABLE export_db.%s AS SELECT * FROM %s", viewName, viewName,
		))
		if err != nil {
			return fmt.Errorf("mtgjson: export table %s: %w", viewName, err)
		}
	}
	return nil
}

// Connection returns the underlying Connection for advanced usage.
func (s *SDK) Connection() *Connection {
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

// containsStr checks if a string contains a substring (used internally).
func containsStr(s, substr string) bool {
	return strings.Contains(s, substr)
}
