package mtgjson

import (
	"context"
	"testing"
)

var sampleCards = []map[string]any{
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
		"printings": []any{"A25", "2ED", "3ED", "M10", "M11"},
		"purchaseUrls": map[string]any{}, "relatedCards": nil,
		"setCode": "A25", "number": "141", "artist": "Christopher Moeller",
		"artistIds": nil, "borderColor": "black", "frameVersion": "2015",
		"frameEffects": nil, "watermark": nil, "signature": nil, "securityStamp": nil,
		"flavorText": nil, "flavorName": nil, "faceFlavorName": nil,
		"originalText": "Lightning Bolt deals 3 damage to any target.",
		"originalType": "Instant",
		"printedName": nil, "printedText": nil, "printedType": nil, "facePrintedName": nil,
		"availability": []any{"paper", "mtgo"}, "boosterTypes": nil,
		"finishes": []any{"nonfoil", "foil"}, "promoTypes": nil, "attractionLights": nil,
		"isFullArt": nil, "isOnlineOnly": nil, "isOversized": nil,
		"isPromo": nil, "isReprint": true, "isTextless": nil,
		"otherFaceIds": nil, "cardParts": nil,
		"language": "English", "sourceProducts": nil,
		"rarity": "uncommon", "duelDeck": nil,
		"isRebalanced": nil, "originalPrintings": nil, "rebalancedPrintings": nil,
		"originalReleaseDate": nil, "isAlternative": nil, "isStorySpotlight": nil,
		"isTimeshifted": nil, "hasContentWarning": nil, "variations": nil,
		"foreignData": nil,
	},
	{
		"uuid": "card-uuid-002", "name": "Counterspell", "asciiName": "Counterspell",
		"faceName": nil, "type": "Instant", "types": []any{"Instant"},
		"subtypes": []any{}, "supertypes": []any{},
		"colors": []any{"U"}, "colorIdentity": []any{"U"},
		"colorIndicator": nil, "producedMana": nil,
		"manaCost": "{U}{U}", "text": "Counter target spell.",
		"layout": "normal", "side": nil,
		"power": nil, "toughness": nil, "loyalty": nil, "defense": nil, "hand": nil, "life": nil,
		"keywords": nil, "identifiers": map[string]any{},
		"isFunny": nil, "edhrecSaltiness": 1.5, "subsets": nil,
		"convertedManaCost": 2.0, "manaValue": 2.0,
		"faceConvertedManaCost": nil, "faceManaValue": nil,
		"edhrecRank": 10, "legalities": map[string]any{},
		"leadershipSkills": nil, "rulings": nil,
		"hasAlternativeDeckLimit": nil, "isReserved": nil, "isGameChanger": nil,
		"printings": []any{"MH2", "A25"},
		"purchaseUrls": map[string]any{}, "relatedCards": nil,
		"setCode": "MH2", "number": "267", "artist": "Zack Stella",
		"artistIds": nil, "borderColor": "black", "frameVersion": "2015",
		"frameEffects": nil, "watermark": nil, "signature": nil, "securityStamp": nil,
		"flavorText": nil, "flavorName": nil, "faceFlavorName": nil,
		"originalText": "Counter target spell.",
		"originalType": "Instant",
		"printedName": nil, "printedText": nil, "printedType": nil, "facePrintedName": nil,
		"availability": []any{"paper", "mtgo"}, "boosterTypes": nil,
		"finishes": []any{"nonfoil", "foil"}, "promoTypes": nil, "attractionLights": nil,
		"isFullArt": nil, "isOnlineOnly": nil, "isOversized": nil,
		"isPromo": nil, "isReprint": true, "isTextless": nil,
		"otherFaceIds": nil, "cardParts": nil,
		"language": "English", "sourceProducts": nil,
		"rarity": "uncommon", "duelDeck": nil,
		"isRebalanced": nil, "originalPrintings": nil, "rebalancedPrintings": nil,
		"originalReleaseDate": nil, "isAlternative": nil, "isStorySpotlight": nil,
		"isTimeshifted": nil, "hasContentWarning": nil, "variations": nil,
		"foreignData": nil,
	},
	{
		"uuid": "card-uuid-003", "name": "Fire // Ice", "asciiName": "Fire // Ice",
		"faceName": "Fire", "type": "Instant", "types": []any{"Instant"},
		"subtypes": []any{}, "supertypes": []any{},
		"colors": []any{"R"}, "colorIdentity": []any{"R", "U"},
		"colorIndicator": nil, "producedMana": nil,
		"manaCost": "{1}{R}", "text": "Fire deals 2 damage divided as you choose among one or two targets.",
		"layout": "split", "side": "a",
		"power": nil, "toughness": nil, "loyalty": nil, "defense": nil, "hand": nil, "life": nil,
		"keywords": nil, "identifiers": map[string]any{},
		"isFunny": nil, "edhrecSaltiness": nil, "subsets": nil,
		"convertedManaCost": 4.0, "manaValue": 4.0,
		"faceConvertedManaCost": 2.0, "faceManaValue": 2.0,
		"edhrecRank": 100, "legalities": map[string]any{},
		"leadershipSkills": nil, "rulings": nil,
		"hasAlternativeDeckLimit": nil, "isReserved": nil, "isGameChanger": nil,
		"printings": []any{"A25"},
		"purchaseUrls": map[string]any{}, "relatedCards": nil,
		"setCode": "A25", "number": "223a", "artist": "Dan Scott",
		"artistIds": nil, "borderColor": "black", "frameVersion": "2015",
		"frameEffects": nil, "watermark": nil, "signature": nil, "securityStamp": nil,
		"flavorText": nil, "flavorName": nil, "faceFlavorName": nil,
		"originalText": nil, "originalType": nil,
		"printedName": nil, "printedText": nil, "printedType": nil, "facePrintedName": nil,
		"availability": []any{"paper", "mtgo"}, "boosterTypes": nil,
		"finishes": []any{"nonfoil"}, "promoTypes": nil, "attractionLights": nil,
		"isFullArt": nil, "isOnlineOnly": nil, "isOversized": nil,
		"isPromo": nil, "isReprint": true, "isTextless": nil,
		"otherFaceIds": []any{"card-uuid-004"}, "cardParts": nil,
		"language": "English", "sourceProducts": nil,
		"rarity": "uncommon", "duelDeck": nil,
		"isRebalanced": nil, "originalPrintings": nil, "rebalancedPrintings": nil,
		"originalReleaseDate": nil, "isAlternative": nil, "isStorySpotlight": nil,
		"isTimeshifted": nil, "hasContentWarning": nil, "variations": nil,
		"foreignData": nil,
	},
}

var sampleSets = []map[string]any{
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
	},
}

var sampleTokens = []map[string]any{
	{
		"uuid": "token-uuid-001", "name": "Soldier Token", "asciiName": "Soldier Token",
		"faceName": nil, "type": "Token Creature — Soldier",
		"types": []any{"Token", "Creature"}, "subtypes": []any{"Soldier"},
		"supertypes": []any{},
		"colors": []any{"W"}, "colorIdentity": []any{"W"},
		"colorIndicator": nil, "producedMana": nil,
		"power": "1", "toughness": "1", "text": nil,
		"layout": "token", "artist": "Zack Stella",
		"artistIds": nil, "borderColor": "black", "frameVersion": "2015",
		"frameEffects": nil, "watermark": nil,
		"availability": []any{"paper"}, "finishes": []any{"nonfoil"},
		"promoTypes": nil, "keywords": nil, "otherFaceIds": nil,
		"boosterTypes": nil, "reverseRelated": nil,
		"isFullArt": nil, "isOnlineOnly": nil, "isPromo": nil,
		"isReprint": nil, "isTextless": nil,
		"identifiers": map[string]any{}, "relatedCards": nil, "sourceProducts": nil,
		"setCode": "A25", "number": "T1", "language": "English",
	},
	{
		"uuid": "token-uuid-002", "name": "Beast Token", "asciiName": "Beast Token",
		"faceName": nil, "type": "Token Creature — Beast",
		"types": []any{"Token", "Creature"}, "subtypes": []any{"Beast"},
		"supertypes": []any{},
		"colors": []any{"G"}, "colorIdentity": []any{"G"},
		"colorIndicator": nil, "producedMana": nil,
		"power": "3", "toughness": "3", "text": nil,
		"layout": "token", "artist": "Jason Rainville",
		"artistIds": nil, "borderColor": "black", "frameVersion": "2015",
		"frameEffects": nil, "watermark": nil,
		"availability": []any{"paper"}, "finishes": []any{"nonfoil"},
		"promoTypes": nil, "keywords": nil, "otherFaceIds": nil,
		"boosterTypes": nil, "reverseRelated": nil,
		"isFullArt": nil, "isOnlineOnly": nil, "isPromo": nil,
		"isReprint": nil, "isTextless": nil,
		"identifiers": map[string]any{}, "relatedCards": nil, "sourceProducts": nil,
		"setCode": "MH2", "number": "T2", "language": "English",
	},
}

var sampleIdentifiers = []map[string]any{
	{
		"uuid": "card-uuid-001", "scryfallId": "scryfall-001",
		"scryfallOracleId": "oracle-001", "scryfallIllustrationId": "illust-001",
		"tcgplayerProductId": "12345", "tcgplayerEtchedProductId": nil,
		"mtgoId": "mtgo-001", "mtgoFoilId": "mtgo-foil-001",
		"mtgArenaId": "arena-001", "multiverseId": "442130",
		"mcmId": "mcm-001", "mcmMetaId": "mcm-meta-001",
		"cardKingdomId": "ck-001", "cardKingdomFoilId": "ck-foil-001",
		"cardKingdomEtchedId": nil, "cardsphereId": "cs-001",
	},
	{
		"uuid": "card-uuid-002", "scryfallId": "scryfall-002",
		"scryfallOracleId": "oracle-002", "scryfallIllustrationId": "illust-002",
		"tcgplayerProductId": "67890", "tcgplayerEtchedProductId": nil,
		"mtgoId": "mtgo-002", "mtgoFoilId": nil,
		"mtgArenaId": "arena-002", "multiverseId": "522205",
		"mcmId": "mcm-002", "mcmMetaId": nil,
		"cardKingdomId": "ck-002", "cardKingdomFoilId": nil,
		"cardKingdomEtchedId": nil, "cardsphereId": nil,
	},
}

var sampleForeignData = []map[string]any{
	{
		"uuid": "card-uuid-001", "name": "Blitzschlag", "language": "German",
		"text": "Der Blitzschlag fügt einem Ziel deiner Wahl 3 Schadenspunkte zu.",
		"type": "Spontanzauber", "faceName": nil, "flavorText": nil, "multiverseId": nil,
	},
	{
		"uuid": "card-uuid-001", "name": "Foudre", "language": "French",
		"text": "La Foudre inflige 3 blessures à n'importe quelle cible.",
		"type": "Éphémère", "faceName": nil, "flavorText": nil, "multiverseId": nil,
	},
	{
		"uuid": "card-uuid-002", "name": "Contresort", "language": "French",
		"text": "Contrecarrez le sort ciblé.",
		"type": "Éphémère", "faceName": nil, "flavorText": nil, "multiverseId": nil,
	},
}

var sampleLegalities = []map[string]any{
	{"uuid": "card-uuid-001", "format": "modern", "status": "Legal"},
	{"uuid": "card-uuid-001", "format": "legacy", "status": "Legal"},
	{"uuid": "card-uuid-001", "format": "vintage", "status": "Restricted"},
	{"uuid": "card-uuid-001", "format": "standard", "status": "Not Legal"},
	{"uuid": "card-uuid-002", "format": "modern", "status": "Legal"},
	{"uuid": "card-uuid-002", "format": "legacy", "status": "Legal"},
	{"uuid": "card-uuid-002", "format": "vintage", "status": "Legal"},
	{"uuid": "card-uuid-002", "format": "standard", "status": "Not Legal"},
	{"uuid": "card-uuid-002", "format": "historic", "status": "Suspended"},
}

var samplePrices = []map[string]any{
	{
		"uuid": "card-uuid-001", "source": "paper", "provider": "tcgplayer",
		"currency": "USD", "category": "retail", "finish": "normal",
		"date": "2024-01-03", "price": 2.00,
	},
	{
		"uuid": "card-uuid-002", "source": "paper", "provider": "tcgplayer",
		"currency": "USD", "category": "retail", "finish": "normal",
		"date": "2024-01-03", "price": 5.00,
	},
	{
		"uuid": "card-uuid-003", "source": "paper", "provider": "tcgplayer",
		"currency": "USD", "category": "retail", "finish": "normal",
		"date": "2024-01-03", "price": 3.00,
	},
}

// setupSampleDB creates a DuckDB connection with sample data for testing.
func setupSampleDB(t *testing.T) *Connection {
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
	for _, td := range []struct {
		name string
		data []map[string]any
	}{
		{"cards", sampleCards},
		{"sets", sampleSets},
		{"tokens", sampleTokens},
		{"card_identifiers", sampleIdentifiers},
		{"card_legalities", sampleLegalities},
		{"card_foreign_data", sampleForeignData},
	} {
		if err := conn.RegisterTableFromData(ctx, td.name, td.data); err != nil {
			t.Fatalf("register %s: %v", td.name, err)
		}
	}
	return conn
}

// setupSampleSDK creates an SDK with sample data for testing (no network).
func setupSampleSDK(t *testing.T) *SDK {
	t.Helper()
	sdk, err := New(WithCacheDir(t.TempDir()), WithOffline(true))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sdk.Close() })

	ctx := context.Background()
	for _, td := range []struct {
		name string
		data []map[string]any
	}{
		{"cards", sampleCards},
		{"sets", sampleSets},
		{"tokens", sampleTokens},
		{"card_identifiers", sampleIdentifiers},
		{"card_legalities", sampleLegalities},
		{"card_foreign_data", sampleForeignData},
	} {
		if err := sdk.conn.RegisterTableFromData(ctx, td.name, td.data); err != nil {
			t.Fatalf("register %s: %v", td.name, err)
		}
	}
	return sdk
}
