package mtgjson

import (
	"context"
	"testing"
)

// sampleSetsWithSealed adds sealedProduct data to set fixtures.
var sampleSetsWithSealed = []map[string]any{
	{
		"code": "A25", "name": "Masters 25", "type": "masters",
		"releaseDate": "2018-03-16", "baseSetSize": 249, "totalSetSize": 249,
		"keyruneCode": "A25", "translations": map[string]any{},
		"block": nil, "parentCode": nil, "mtgoCode": "A25", "tokenSetCode": nil,
		"mcmId": nil, "mcmIdExtras": nil, "mcmName": nil,
		"tcgplayerGroupId": nil, "cardsphereSetId": nil,
		"isFoilOnly": false, "isNonFoilOnly": nil, "isOnlineOnly": false,
		"isPaperOnly": nil, "isForeignOnly": nil, "isPartialPreview": nil,
		"languages": []any{"English"},
		"sealedProduct": []any{
			map[string]any{
				"uuid": "sealed-uuid-001", "name": "Masters 25 Booster Box",
				"category": "booster_box", "subtype": nil,
				"purchaseUrls": map[string]any{}, "releaseDate": "2018-03-16",
			},
			map[string]any{
				"uuid": "sealed-uuid-002", "name": "Masters 25 Booster Pack",
				"category": "booster_pack", "subtype": nil,
				"purchaseUrls": map[string]any{}, "releaseDate": "2018-03-16",
			},
		},
		"booster": nil,
	},
	{
		"code": "MH2", "name": "Modern Horizons 2", "type": "draft_innovation",
		"releaseDate": "2021-06-18", "baseSetSize": 303, "totalSetSize": 531,
		"keyruneCode": "MH2", "translations": map[string]any{},
		"block": nil, "parentCode": nil, "mtgoCode": "MH2", "tokenSetCode": nil,
		"mcmId": nil, "mcmIdExtras": nil, "mcmName": nil,
		"tcgplayerGroupId": nil, "cardsphereSetId": nil,
		"isFoilOnly": false, "isNonFoilOnly": nil, "isOnlineOnly": false,
		"isPaperOnly": nil, "isForeignOnly": nil, "isPartialPreview": nil,
		"languages": []any{"English"},
		"sealedProduct": []any{
			map[string]any{
				"uuid": "sealed-uuid-003", "name": "MH2 Set Booster Box",
				"category": "booster_box", "subtype": nil,
				"purchaseUrls": map[string]any{}, "releaseDate": "2021-06-18",
			},
		},
		"booster": nil,
	},
}

func setupSealedDB(t *testing.T) *Connection {
	t.Helper()
	cfg := defaultConfig()
	cfg.cacheDir = t.TempDir()
	cfg.offline = true
	cache, err := newCacheManager(cfg)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := NewConnection(cache)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	ctx := context.Background()
	if err := conn.RegisterTableFromData(ctx, "sets", sampleSetsWithSealed); err != nil {
		t.Fatalf("register sets: %v", err)
	}
	return conn
}

func TestSealedList(t *testing.T) {
	conn := setupSealedDB(t)
	sq := newSealedQuery(conn)
	ctx := context.Background()

	products, err := sq.List(ctx, ListSealedParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 3 {
		t.Fatalf("expected 3 products, got %d", len(products))
	}
}

func TestSealedListBySet(t *testing.T) {
	conn := setupSealedDB(t)
	sq := newSealedQuery(conn)
	ctx := context.Background()

	products, err := sq.List(ctx, ListSealedParams{SetCode: "A25"})
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}
	for _, p := range products {
		if p["setCode"] != "A25" {
			t.Fatalf("expected setCode A25, got %v", p["setCode"])
		}
	}
}

func TestSealedListByCategory(t *testing.T) {
	conn := setupSealedDB(t)
	sq := newSealedQuery(conn)
	ctx := context.Background()

	products, err := sq.List(ctx, ListSealedParams{Category: "booster_box"})
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 2 {
		t.Fatalf("expected 2 booster_box products, got %d", len(products))
	}
}

func TestSealedGet(t *testing.T) {
	conn := setupSealedDB(t)
	sq := newSealedQuery(conn)
	ctx := context.Background()

	product, err := sq.Get(ctx, "sealed-uuid-001")
	if err != nil {
		t.Fatal(err)
	}
	if product == nil {
		t.Fatal("expected product, got nil")
	}
	if product["name"] != "Masters 25 Booster Box" {
		t.Fatalf("expected 'Masters 25 Booster Box', got %v", product["name"])
	}
	if product["setCode"] != "A25" {
		t.Fatalf("expected setCode A25, got %v", product["setCode"])
	}
}

func TestSealedGetNotFound(t *testing.T) {
	conn := setupSealedDB(t)
	sq := newSealedQuery(conn)
	ctx := context.Background()

	product, err := sq.Get(ctx, "nonexistent-uuid")
	if err != nil {
		t.Fatal(err)
	}
	if product != nil {
		t.Fatalf("expected nil, got %v", product)
	}
}
