package queries

import (
	"context"

	"github.com/mtgjson/mtgjson-sdk-go/db"
	"github.com/mtgjson/mtgjson-sdk-go/models"
)

// SkuQuery provides methods to query TCGPlayer SKU data.
// SKUs represent individual purchasable variants of a card.
type SkuQuery struct {
	conn *db.Connection
}

func NewSkuQuery(conn *db.Connection) *SkuQuery {
	return &SkuQuery{conn: conn}
}

func (q *SkuQuery) ensure(ctx context.Context) {
	_ = q.conn.EnsureViews(ctx, "tcgplayer_skus")
}

// Get returns all TCGPlayer SKUs for a card UUID.
func (q *SkuQuery) Get(ctx context.Context, uuid string) ([]models.TcgplayerSkus, error) {
	q.ensure(ctx)
	if !q.conn.HasView("tcgplayer_skus") {
		return nil, nil
	}
	var skus []models.TcgplayerSkus
	if err := q.conn.ExecuteInto(ctx, &skus, "SELECT * FROM tcgplayer_skus WHERE uuid = $1", uuid); err != nil {
		return nil, err
	}
	return skus, nil
}

// FindBySkuID finds a SKU by its TCGPlayer SKU ID.
func (q *SkuQuery) FindBySkuID(ctx context.Context, skuID int) (map[string]any, error) {
	q.ensure(ctx)
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
	q.ensure(ctx)
	if !q.conn.HasView("tcgplayer_skus") {
		return nil, nil
	}
	return q.conn.Execute(ctx, "SELECT * FROM tcgplayer_skus WHERE productId = $1", productID)
}
