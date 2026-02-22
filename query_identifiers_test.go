package mtgjson

import (
	"context"
	"strings"
	"testing"
)

func TestIdentFindByScryfallID(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	cards, err := sdk.Identifiers().FindByScryfallID(ctx, "scryfall-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
	if cards[0].Name != "Lightning Bolt" {
		t.Fatalf("expected Lightning Bolt, got %s", cards[0].Name)
	}
}

func TestIdentFindByScryfallOracleID(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	cards, err := sdk.Identifiers().FindByScryfallOracleID(ctx, "oracle-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
	if cards[0].Name != "Lightning Bolt" {
		t.Fatalf("expected Lightning Bolt, got %s", cards[0].Name)
	}
}

func TestIdentFindByScryfallIllustrationID(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	cards, err := sdk.Identifiers().FindByScryfallIllustrationID(ctx, "illust-002")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
	if cards[0].Name != "Counterspell" {
		t.Fatalf("expected Counterspell, got %s", cards[0].Name)
	}
}

func TestIdentFindByTCGPlayerID(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	cards, err := sdk.Identifiers().FindByTCGPlayerID(ctx, "12345")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
	if cards[0].Name != "Lightning Bolt" {
		t.Fatalf("expected Lightning Bolt, got %s", cards[0].Name)
	}
}

func TestIdentFindByMTGOID(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	cards, err := sdk.Identifiers().FindByMTGOID(ctx, "mtgo-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
}

func TestIdentFindByMTGOFoilID(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	cards, err := sdk.Identifiers().FindByMTGOFoilID(ctx, "mtgo-foil-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
	if cards[0].Name != "Lightning Bolt" {
		t.Fatalf("expected Lightning Bolt, got %s", cards[0].Name)
	}
}

func TestIdentFindByMTGArenaID(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	cards, err := sdk.Identifiers().FindByMTGArenaID(ctx, "arena-002")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
	if cards[0].Name != "Counterspell" {
		t.Fatalf("expected Counterspell, got %s", cards[0].Name)
	}
}

func TestIdentFindByMultiverseID(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	cards, err := sdk.Identifiers().FindByMultiverseID(ctx, "442130")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
}

func TestIdentFindByMCMID(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	cards, err := sdk.Identifiers().FindByMCMID(ctx, "mcm-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
}

func TestIdentFindByMCMMetaID(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	cards, err := sdk.Identifiers().FindByMCMMetaID(ctx, "mcm-meta-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
	if cards[0].Name != "Lightning Bolt" {
		t.Fatalf("expected Lightning Bolt, got %s", cards[0].Name)
	}
}

func TestIdentFindByCardKingdomID(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	cards, err := sdk.Identifiers().FindByCardKingdomID(ctx, "ck-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
}

func TestIdentFindByCardKingdomFoilID(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	cards, err := sdk.Identifiers().FindByCardKingdomFoilID(ctx, "ck-foil-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
	if cards[0].Name != "Lightning Bolt" {
		t.Fatalf("expected Lightning Bolt, got %s", cards[0].Name)
	}
}

func TestIdentFindByCardsphereID(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	cards, err := sdk.Identifiers().FindByCardsphereID(ctx, "cs-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
}

func TestIdentFindByGeneric(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	cards, err := sdk.Identifiers().FindBy(ctx, "scryfallId", "scryfall-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
	if cards[0].Name != "Lightning Bolt" {
		t.Fatalf("expected Lightning Bolt, got %s", cards[0].Name)
	}
}

func TestIdentFindByGenericInvalidType(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	_, err := sdk.Identifiers().FindBy(ctx, "invalidColumn", "123")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unknown identifier type") {
		t.Fatalf("expected 'unknown identifier type' error, got: %v", err)
	}
}

func TestIdentGetIdentifiers(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	ids, err := sdk.Identifiers().GetIdentifiers(ctx, "card-uuid-001")
	if err != nil {
		t.Fatal(err)
	}
	if ids == nil {
		t.Fatal("expected identifiers, got nil")
	}
	if ids["scryfallId"] != "scryfall-001" {
		t.Fatalf("expected scryfall-001, got %v", ids["scryfallId"])
	}
	if ids["mtgArenaId"] != "arena-001" {
		t.Fatalf("expected arena-001, got %v", ids["mtgArenaId"])
	}
}

func TestIdentFindByNotFound(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	cards, err := sdk.Identifiers().FindByScryfallID(ctx, "nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 0 {
		t.Fatalf("expected 0, got %d", len(cards))
	}
}
