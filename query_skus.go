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
)

// SkuQuery provides methods to query TCGPlayer SKU data.
// SKUs represent individual purchasable variants of a card.
type SkuQuery struct {
	conn   *Connection
	cache  *CacheManager
	loaded bool
}

func newSkuQuery(conn *Connection, cache *CacheManager) *SkuQuery {
	return &SkuQuery{conn: conn, cache: cache}
}

func (q *SkuQuery) ensure(ctx context.Context) error {
	if q.loaded {
		return nil
	}
	if q.conn.HasView("tcgplayer_skus") {
		q.loaded = true
		return nil
	}
	path, err := q.cache.EnsureJSON(ctx, "tcgplayer_skus")
	if err != nil {
		slog.Warn("SKU data not available", "error", err)
		q.loaded = true
		return nil
	}
	if err := loadSkusToDuckDB(ctx, path, q.conn); err != nil {
		return err
	}
	q.loaded = true
	return nil
}

// Get returns all TCGPlayer SKUs for a card UUID.
func (q *SkuQuery) Get(ctx context.Context, uuid string) ([]TcgplayerSkus, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}
	if !q.conn.HasView("tcgplayer_skus") {
		return nil, nil
	}
	var skus []TcgplayerSkus
	if err := q.conn.ExecuteInto(ctx, &skus, "SELECT * FROM tcgplayer_skus WHERE uuid = $1", uuid); err != nil {
		return nil, err
	}
	return skus, nil
}

// FindBySkuID finds a SKU by its TCGPlayer SKU ID.
func (q *SkuQuery) FindBySkuID(ctx context.Context, skuID int) (map[string]any, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}
	if !q.conn.HasView("tcgplayer_skus") {
		return nil, nil
	}
	rows, err := q.conn.Execute(ctx, "SELECT * FROM tcgplayer_skus WHERE skuId = $1", skuID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

// FindByProductID finds all SKUs for a TCGPlayer product ID.
func (q *SkuQuery) FindByProductID(ctx context.Context, productID int) ([]map[string]any, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}
	if !q.conn.HasView("tcgplayer_skus") {
		return nil, nil
	}
	return q.conn.Execute(ctx, "SELECT * FROM tcgplayer_skus WHERE productId = $1", productID)
}

// loadSkusToDuckDB parses TcgplayerSkus JSON, stream-flattens to NDJSON, and loads into DuckDB.
func loadSkusToDuckDB(ctx context.Context, path string, conn *Connection) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("mtgjson: open sku file: %w", err)
	}
	defer f.Close()

	var reader io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		gr, err := gzip.NewReader(f)
		if err != nil {
			return fmt.Errorf("mtgjson: decompress sku file: %w", err)
		}
		defer gr.Close()
		reader = gr
	}

	var raw map[string]any
	if err := json.NewDecoder(reader).Decode(&raw); err != nil {
		return fmt.Errorf("mtgjson: parse sku JSON: %w", err)
	}

	data, _ := raw["data"].(map[string]any)
	if data == nil {
		return nil
	}

	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("mtgjson_skus_%d.ndjson", os.Getpid()))
	ndjson, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	count := 0
	enc := json.NewEncoder(ndjson)
	for uuid, skusRaw := range data {
		skus, ok := skusRaw.([]any)
		if !ok {
			continue
		}
		for _, skuRaw := range skus {
			sku, ok := skuRaw.(map[string]any)
			if !ok {
				continue
			}
			row := make(map[string]any, len(sku)+1)
			for k, v := range sku {
				row[k] = v
			}
			row["uuid"] = uuid
			if err := enc.Encode(row); err != nil {
				ndjson.Close()
				return err
			}
			count++
		}
	}
	ndjson.Close()

	if count > 0 {
		return conn.RegisterTableFromNdjson(ctx, "tcgplayer_skus", tmpPath)
	}
	return nil
}
