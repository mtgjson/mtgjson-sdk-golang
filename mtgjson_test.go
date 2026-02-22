package mtgjsonsdk

import (
	"context"
	"testing"
)

var sampleCardsRoot = []map[string]any{
	{
		"uuid": "card-uuid-001", "name": "Lightning Bolt", "asciiName": "Lightning Bolt",
		"faceName": nil, "type": "Instant", "types": []any{"Instant"},
		"subtypes": []any{}, "supertypes": []any{},
		"colors": []any{"R"}, "colorIdentity": []any{"R"},
		"colorIndicator": nil, "producedMana": nil,
		"manaCost": "{R}", "text": "Lightning Bolt deals 3 damage to any target.",
		"layout": "normal", "side": nil,
		"power": nil, "toughness": nil, "loyalty": nil, "defense": nil, "hand": nil, "life": nil,
		"keywords": nil, "identifiers": map[string]any{},
		"isFunny": nil, "edhrecSaltiness": nil, "subsets": nil,
		"convertedManaCost": 1.0, "manaValue": 1.0,
		"faceConvertedManaCost": nil, "faceManaValue": nil,
		"edhrecRank": 5, "legalities": map[string]any{},
		"leadershipSkills": nil, "rulings": nil,
		"hasAlternativeDeckLimit": nil, "isReserved": nil, "isGameChanger": nil,
		"printings": []any{"A25"}, "purchaseUrls": map[string]any{}, "relatedCards": nil,
		"setCode": "A25", "number": "141", "artist": "Christopher Moeller",
		"artistIds": nil, "borderColor": "black", "frameVersion": "2015",
		"frameEffects": nil, "watermark": nil, "signature": nil, "securityStamp": nil,
		"flavorText": nil, "flavorName": nil, "faceFlavorName": nil,
		"originalText": nil, "originalType": nil,
		"printedName": nil, "printedText": nil, "printedType": nil, "facePrintedName": nil,
		"availability": []any{"paper"}, "boosterTypes": nil,
		"finishes": []any{"nonfoil"}, "promoTypes": nil, "attractionLights": nil,
		"isFullArt": nil, "isOnlineOnly": nil, "isOversized": nil,
		"isPromo": nil, "isReprint": nil, "isTextless": nil,
		"otherFaceIds": nil, "cardParts": nil,
		"language": "English", "sourceProducts": nil,
		"rarity": "uncommon", "duelDeck": nil,
		"isRebalanced": nil, "originalPrintings": nil, "rebalancedPrintings": nil,
		"originalReleaseDate": nil, "isAlternative": nil, "isStorySpotlight": nil,
		"isTimeshifted": nil, "hasContentWarning": nil, "variations": nil,
		"foreignData": nil,
	},
}

func setupSampleSDK(t *testing.T) *SDK {
	t.Helper()
	sdk, err := New(WithCacheDir(t.TempDir()), WithOffline(true))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sdk.Close() })

	ctx := context.Background()
	if err := sdk.conn.RegisterTableFromData(ctx, "cards", sampleCardsRoot); err != nil {
		t.Fatalf("register cards: %v", err)
	}
	return sdk
}

func TestSDKNewAndClose(t *testing.T) {
	sdk, err := New(WithCacheDir(t.TempDir()), WithOffline(true))
	if err != nil {
		t.Fatal(err)
	}
	if err := sdk.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSDKViews(t *testing.T) {
	sdk := setupSampleSDK(t)
	views := sdk.Views()
	if len(views) == 0 {
		t.Fatal("expected some views registered")
	}
	found := false
	for _, v := range views {
		if v == "cards" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'cards' in views, got %v", views)
	}
}

func TestSDKSQL(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	rows, err := sdk.SQL(ctx, "SELECT name FROM cards ORDER BY name LIMIT 1")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1, got %d", len(rows))
	}
}

func TestSDKString(t *testing.T) {
	sdk := setupSampleSDK(t)
	s := sdk.String()
	if s == "" {
		t.Fatal("expected non-empty string")
	}
}
