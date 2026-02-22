package mtgjson

import (
	"context"
	"testing"
)

var sampleSkuData = []map[string]any{
	{
		"uuid": "card-uuid-001", "skuId": 12345, "productId": 100,
		"condition": "Near Mint", "finish": "Normal",
		"language": "English", "printing": "Normal",
	},
	{
		"uuid": "card-uuid-001", "skuId": 12346, "productId": 100,
		"condition": "Near Mint", "finish": "Foil",
		"language": "English", "printing": "Foil",
	},
	{
		"uuid": "card-uuid-002", "skuId": 67890, "productId": 200,
		"condition": "Near Mint", "finish": "Normal",
		"language": "English", "printing": "Normal",
	},
}

func setupSkuQuery(t *testing.T) *SkuQuery {
	t.Helper()
	conn := setupSampleDB(t)
	ctx := context.Background()
	if err := conn.RegisterTableFromData(ctx, "tcgplayer_skus", sampleSkuData); err != nil {
		t.Fatal(err)
	}
	return &SkuQuery{conn: conn, cache: nil, loaded: true}
}

func TestSkuGet(t *testing.T) {
	sq := setupSkuQuery(t)
	ctx := context.Background()

	skus, err := sq.Get(ctx, "card-uuid-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(skus) != 2 {
		t.Fatalf("expected 2 SKUs, got %d", len(skus))
	}
}

func TestSkuGetNotFound(t *testing.T) {
	sq := setupSkuQuery(t)
	ctx := context.Background()

	skus, err := sq.Get(ctx, "nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if len(skus) != 0 {
		t.Fatalf("expected 0 SKUs, got %d", len(skus))
	}
}

func TestSkuFindBySkuID(t *testing.T) {
	sq := setupSkuQuery(t)
	ctx := context.Background()

	sku, err := sq.FindBySkuID(ctx, 12345)
	if err != nil {
		t.Fatal(err)
	}
	if sku == nil {
		t.Fatal("expected SKU, got nil")
	}
	if sku["uuid"] != "card-uuid-001" {
		t.Fatalf("expected card-uuid-001, got %v", sku["uuid"])
	}
}

func TestSkuFindBySkuIDNotFound(t *testing.T) {
	sq := setupSkuQuery(t)
	ctx := context.Background()

	sku, err := sq.FindBySkuID(ctx, 99999)
	if err != nil {
		t.Fatal(err)
	}
	if sku != nil {
		t.Fatalf("expected nil, got %v", sku)
	}
}

func TestSkuFindByProductID(t *testing.T) {
	sq := setupSkuQuery(t)
	ctx := context.Background()

	rows, err := sq.FindByProductID(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}

func TestSkuFindByProductIDNotFound(t *testing.T) {
	sq := setupSkuQuery(t)
	ctx := context.Background()

	rows, err := sq.FindByProductID(ctx, 99999)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(rows))
	}
}

func TestSkuNoTable(t *testing.T) {
	conn := setupSampleDB(t)
	sq := &SkuQuery{conn: conn, cache: nil, loaded: true}
	ctx := context.Background()

	skus, err := sq.Get(ctx, "card-uuid-001")
	if err != nil {
		t.Fatal(err)
	}
	if skus != nil {
		t.Fatalf("expected nil, got %v", skus)
	}
}
