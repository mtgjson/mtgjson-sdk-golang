package queries

import (
	"context"
	"encoding/json"

	"github.com/mtgjson/mtgjson-sdk-go/db"
)

// SealedQuery provides methods to query sealed product data (booster boxes, bundles, etc.).
// Sealed product data lives inside the sets table as nested structs.
type SealedQuery struct {
	conn *db.Connection
}

func NewSealedQuery(conn *db.Connection) *SealedQuery {
	return &SealedQuery{conn: conn}
}

func (q *SealedQuery) ensure(ctx context.Context) error {
	return q.conn.EnsureViews(ctx, "sets")
}

// ListSealedParams contains optional filters for listing sealed products.
type ListSealedParams struct {
	SetCode  string
	Category string
	Limit    int
}

// List returns sealed products from set data.
// Note: Requires the sealedProduct column (present in AllPrintings or test data,
// but NOT in the flat sets.parquet from CDN).
func (q *SealedQuery) List(ctx context.Context, params ListSealedParams) ([]map[string]any, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}
	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}

	sb := db.NewSQLBuilder("sets")
	sb.Select("code", "name AS setName", "sealedProduct")
	if params.SetCode != "" {
		sb.WhereEq("code", params.SetCode)
	}
	sb.Limit(limit)
	sql, sqlParams := sb.Build()

	rows, err := q.conn.Execute(ctx, sql, sqlParams...)
	if err != nil {
		// sealedProduct column may not exist in flat sets.parquet
		return nil, nil
	}

	var products []map[string]any
	for _, row := range rows {
		setCode, _ := row["code"].(string)
		sealed := extractSealedProducts(row["sealedProduct"])
		for _, sp := range sealed {
			if params.Category != "" {
				if cat, _ := sp["category"].(string); cat != params.Category {
					continue
				}
			}
			sp["setCode"] = setCode
			products = append(products, sp)
		}
	}
	return products, nil
}

// Get returns a sealed product by UUID.
func (q *SealedQuery) Get(ctx context.Context, uuid string) (map[string]any, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}

	sql := "SELECT sub.code AS setCode, sub.sp " +
		"FROM (" +
		"  SELECT code, UNNEST(sealedProduct) AS sp " +
		"  FROM sets WHERE sealedProduct IS NOT NULL" +
		") sub " +
		"WHERE sub.sp.uuid = $1 " +
		"LIMIT 1"
	rows, err := q.conn.Execute(ctx, sql, uuid)
	if err != nil {
		// sealedProduct column may not exist
		return nil, nil
	}
	if len(rows) == 0 {
		return nil, nil
	}
	row := rows[0]
	product := extractMapFromValue(row["sp"])
	if product != nil {
		product["setCode"] = row["setCode"]
	}
	return product, nil
}

// extractSealedProducts extracts sealed products from a column value.
func extractSealedProducts(v any) []map[string]any {
	if v == nil {
		return nil
	}

	// Try as []any (DuckDB might return this directly)
	if arr, ok := v.([]any); ok {
		var result []map[string]any
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				result = append(result, m)
			}
		}
		return result
	}

	// Try as string (JSON)
	if s, ok := v.(string); ok {
		var arr []map[string]any
		if err := json.Unmarshal([]byte(s), &arr); err == nil {
			return arr
		}
	}

	return nil
}

// extractMapFromValue converts a DuckDB struct value to map[string]any.
func extractMapFromValue(v any) map[string]any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	if s, ok := v.(string); ok {
		var m map[string]any
		if err := json.Unmarshal([]byte(s), &m); err == nil {
			return m
		}
	}
	// Try marshaling through JSON as a last resort
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}
