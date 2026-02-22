package mtgjson

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	_ "github.com/marcboeker/go-duckdb" // DuckDB driver registration
)

// staticListColumns are known list columns that don't follow the plural naming convention.
var staticListColumns = map[string]map[string]bool{
	"cards": {
		"artistIds": true, "attractionLights": true, "availability": true,
		"boosterTypes": true, "cardParts": true, "colorIdentity": true,
		"colorIndicator": true, "colors": true, "finishes": true,
		"frameEffects": true, "keywords": true, "originalPrintings": true,
		"otherFaceIds": true, "printings": true, "producedMana": true,
		"promoTypes": true, "rebalancedPrintings": true, "subsets": true,
		"subtypes": true, "supertypes": true, "types": true, "variations": true,
	},
	"tokens": {
		"artistIds": true, "availability": true, "boosterTypes": true,
		"colorIdentity": true, "colorIndicator": true, "colors": true,
		"finishes": true, "frameEffects": true, "keywords": true,
		"otherFaceIds": true, "producedMana": true, "promoTypes": true,
		"reverseRelated": true, "subtypes": true, "supertypes": true,
		"types": true,
	},
}

// ignoredColumns are VARCHAR columns that are NOT lists, even if they match the plural heuristic.
var ignoredColumns = map[string]bool{
	"text": true, "originalText": true, "flavorText": true, "printedText": true,
	"identifiers": true, "legalities": true, "leadershipSkills": true,
	"purchaseUrls": true, "relatedCards": true, "rulings": true,
	"sourceProducts": true, "foreignData": true, "translations": true,
	"toughness": true, "status": true, "format": true, "uris": true,
	"scryfallUri": true,
}

// jsonCastColumns are VARCHAR columns containing JSON strings to cast to DuckDB JSON type.
var jsonCastColumns = map[string]bool{
	"identifiers": true, "legalities": true, "leadershipSkills": true,
	"purchaseUrls": true, "relatedCards": true, "rulings": true,
	"sourceProducts": true, "foreignData": true, "translations": true,
}

// Connection wraps a DuckDB database/sql connection and registers parquet files as views.
type Connection struct {
	db              *sql.DB
	cache           *CacheManager
	registeredViews map[string]bool
	mu              sync.RWMutex
}

// NewConnection creates a new in-memory DuckDB connection backed by the given cache.
func NewConnection(cache *CacheManager) (*Connection, error) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("mtgjson: open DuckDB: %w", err)
	}
	// Prevent connection caching issues with temp objects
	db.SetMaxIdleConns(0)
	return &Connection{
		db:              db,
		cache:           cache,
		registeredViews: make(map[string]bool),
	}, nil
}

// Close closes the underlying DuckDB connection.
func (c *Connection) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// EnsureViews ensures one or more views are registered, downloading data if needed.
func (c *Connection) EnsureViews(ctx context.Context, names ...string) error {
	for _, name := range names {
		if err := c.ensureView(ctx, name); err != nil {
			return err
		}
	}
	return nil
}

func (c *Connection) ensureView(ctx context.Context, name string) error {
	c.mu.RLock()
	if c.registeredViews[name] {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.registeredViews[name] {
		return nil
	}

	path, err := c.cache.EnsureParquet(ctx, name)
	if err != nil {
		return err
	}
	pathStr := filepath.ToSlash(path)

	if name == "card_legalities" {
		return c.registerLegalitiesView(ctx, pathStr)
	}

	replaceClause, err := c.buildCSVReplace(ctx, pathStr, name)
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(ctx, fmt.Sprintf(
		"CREATE OR REPLACE VIEW %s AS SELECT *%s FROM read_parquet('%s')",
		name, replaceClause, pathStr,
	))
	if err != nil {
		return fmt.Errorf("mtgjson: register view %s: %w", name, err)
	}
	c.registeredViews[name] = true
	slog.Debug("Registered view", "name", name, "path", pathStr)
	return nil
}

func (c *Connection) buildCSVReplace(ctx context.Context, pathStr, viewName string) (string, error) {
	rows, err := c.db.QueryContext(ctx, fmt.Sprintf(
		"SELECT column_name, column_type FROM (DESCRIBE SELECT * FROM read_parquet('%s'))", pathStr,
	))
	if err != nil {
		return "", err
	}
	defer rows.Close()

	schema := make(map[string]string)
	for rows.Next() {
		var colName, colType string
		if err := rows.Scan(&colName, &colType); err != nil {
			return "", err
		}
		schema[colName] = colType
	}

	// Build candidate set
	candidates := make(map[string]bool)

	// Layer 1: Static baseline
	if static, ok := staticListColumns[viewName]; ok {
		for col := range static {
			candidates[col] = true
		}
	}

	// Layer 2: Dynamic heuristic
	for col, dtype := range schema {
		if dtype != "VARCHAR" {
			continue
		}
		if ignoredColumns[col] {
			continue
		}
		if strings.HasSuffix(col, "s") {
			candidates[col] = true
		}
	}

	// Filter to columns that actually exist as VARCHAR
	var finalCols []string
	for col := range candidates {
		if schema[col] == "VARCHAR" {
			finalCols = append(finalCols, col)
		}
	}
	sort.Strings(finalCols)

	var exprs []string
	for _, col := range finalCols {
		exprs = append(exprs, fmt.Sprintf(
			`CASE WHEN "%s" IS NULL OR TRIM("%s") = '' THEN []::VARCHAR[] ELSE string_split("%s", ', ') END AS "%s"`,
			col, col, col, col,
		))
	}

	// Layer 4: JSON casting
	var jsonCols []string
	for col := range jsonCastColumns {
		jsonCols = append(jsonCols, col)
	}
	sort.Strings(jsonCols)
	for _, col := range jsonCols {
		if schema[col] == "VARCHAR" {
			exprs = append(exprs, fmt.Sprintf(`TRY_CAST("%s" AS JSON) AS "%s"`, col, col))
		}
	}

	if len(exprs) == 0 {
		return "", nil
	}
	return " REPLACE (" + strings.Join(exprs, ", ") + ")", nil
}

func (c *Connection) registerLegalitiesView(ctx context.Context, pathStr string) error {
	rows, err := c.db.QueryContext(ctx, fmt.Sprintf(
		"SELECT column_name FROM (DESCRIBE SELECT * FROM read_parquet('%s'))", pathStr,
	))
	if err != nil {
		return err
	}
	defer rows.Close()

	var allCols []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return err
		}
		allCols = append(allCols, col)
	}

	staticCols := map[string]bool{"uuid": true}
	var formatCols []string
	for _, c := range allCols {
		if !staticCols[c] {
			formatCols = append(formatCols, c)
		}
	}

	if len(formatCols) == 0 {
		_, err = c.db.ExecContext(ctx, fmt.Sprintf(
			"CREATE OR REPLACE VIEW card_legalities AS SELECT * FROM read_parquet('%s')", pathStr,
		))
	} else {
		colsSQL := make([]string, len(formatCols))
		for i, col := range formatCols {
			colsSQL[i] = `"` + col + `"`
		}
		_, err = c.db.ExecContext(ctx, fmt.Sprintf(
			"CREATE OR REPLACE VIEW card_legalities AS "+
				"SELECT uuid, format, status FROM ("+
				"  UNPIVOT (SELECT * FROM read_parquet('%s'))"+
				"  ON %s"+
				"  INTO NAME format VALUE status"+
				") WHERE status IS NOT NULL",
			pathStr, strings.Join(colsSQL, ", "),
		))
	}
	if err != nil {
		return fmt.Errorf("mtgjson: register legalities view: %w", err)
	}
	c.registeredViews["card_legalities"] = true
	slog.Debug("Registered legalities view", "formats", len(formatCols), "path", pathStr)
	return nil
}

// RegisterTableFromData creates a DuckDB table from a slice of maps.
// Primarily used by unit tests with small sample data.
func (c *Connection) RegisterTableFromData(ctx context.Context, tableName string, data []map[string]any) error {
	if len(data) == 0 {
		return nil
	}
	_, err := c.db.ExecContext(ctx, "DROP TABLE IF EXISTS "+tableName)
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("mtgjson_%s_%d.json", tableName, os.Getpid()))
	if err := os.WriteFile(tmpPath, jsonBytes, 0o644); err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	fwd := filepath.ToSlash(tmpPath)
	_, err = c.db.ExecContext(ctx, fmt.Sprintf(
		"CREATE TABLE %s AS SELECT * FROM read_json_auto('%s')", tableName, fwd,
	))
	if err != nil {
		return fmt.Errorf("mtgjson: create table %s: %w", tableName, err)
	}
	c.registeredViews[tableName] = true
	return nil
}

// RegisterTableFromNdjson creates a DuckDB table from a newline-delimited JSON file.
func (c *Connection) RegisterTableFromNdjson(ctx context.Context, tableName, ndjsonPath string) error {
	_, err := c.db.ExecContext(ctx, "DROP TABLE IF EXISTS "+tableName)
	if err != nil {
		return err
	}
	fwd := filepath.ToSlash(ndjsonPath)
	_, err = c.db.ExecContext(ctx, fmt.Sprintf(
		"CREATE TABLE %s AS SELECT * FROM read_json_auto('%s', format='newline_delimited')",
		tableName, fwd,
	))
	if err != nil {
		return fmt.Errorf("mtgjson: create table %s: %w", tableName, err)
	}
	c.registeredViews[tableName] = true
	return nil
}

// Execute runs SQL and returns results as []map[string]any.
func (c *Connection) Execute(ctx context.Context, query string, params ...any) ([]map[string]any, error) {
	rows, err := c.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]any
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = values[i]
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ExecuteJSON runs SQL wrapped in to_json(list(...)) and returns a raw JSON string.
func (c *Connection) ExecuteJSON(ctx context.Context, query string, params ...any) (string, error) {
	wrapped := fmt.Sprintf("SELECT CAST(to_json(list(sub)) AS VARCHAR) FROM (%s) sub", query)
	row := c.db.QueryRowContext(ctx, wrapped, params...)
	var result sql.NullString
	if err := row.Scan(&result); err != nil {
		return "[]", err
	}
	if !result.Valid || result.String == "" {
		return "[]", nil
	}
	return result.String, nil
}

// ExecuteInto runs SQL and JSON-unmarshals results into dst (must be a pointer to a slice).
func (c *Connection) ExecuteInto(ctx context.Context, dst any, query string, params ...any) error {
	jsonStr, err := c.ExecuteJSON(ctx, query, params...)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(jsonStr), dst)
}

// ExecuteScalar runs SQL and returns a single scalar value.
func (c *Connection) ExecuteScalar(ctx context.Context, query string, params ...any) (any, error) {
	row := c.db.QueryRowContext(ctx, query, params...)
	var val any
	if err := row.Scan(&val); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return val, nil
}

// Raw returns the underlying *sql.DB for advanced usage.
func (c *Connection) Raw() *sql.DB {
	return c.db
}

// ClearViews resets the registered views set (used by Refresh).
func (c *Connection) ClearViews() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.registeredViews = make(map[string]bool)
}

// Views returns the names of all registered views.
func (c *Connection) Views() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	names := make([]string, 0, len(c.registeredViews))
	for name := range c.registeredViews {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// HasView checks if a view is registered.
func (c *Connection) HasView(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.registeredViews[name]
}
