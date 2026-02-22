//go:build smoke

package mtgjsonsdk

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mtgjson/mtgjson-sdk-go/queries"
)

func TestSmoke(t *testing.T) {
	sdk, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer sdk.Close()

	ctx := context.Background()

	// ══════════════════════════════════════════════════════════
	//  CLIENT LIFECYCLE
	// ══════════════════════════════════════════════════════════
	t.Run("Meta", func(t *testing.T) {
		meta, err := sdk.Meta(ctx)
		if err != nil {
			t.Fatalf("Meta() error: %v", err)
		}
		if meta.Version == "" {
			t.Fatal("expected non-empty version")
		}
		t.Logf("version=%s date=%s", meta.Version, meta.Date)
	})

	t.Run("String", func(t *testing.T) {
		s := sdk.String()
		if s == "" {
			t.Fatal("expected non-empty string")
		}
		t.Logf("String()=%s", s)
	})

	viewsBefore := sdk.Views()
	t.Run("ViewsInitial", func(t *testing.T) {
		t.Logf("initial views: %v", viewsBefore)
	})

	// ══════════════════════════════════════════════════════════
	//  CARDS
	// ══════════════════════════════════════════════════════════
	var boltUUID string

	t.Run("Cards", func(t *testing.T) {
		t.Run("GetByName", func(t *testing.T) {
			cards, err := sdk.Cards().GetByName(ctx, "Lightning Bolt")
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected at least 1 printing")
			}
			boltUUID = cards[0].UUID
			t.Logf("found %d printings, uuid=%s", len(cards), boltUUID)
		})

		t.Run("GetByNameWithSet", func(t *testing.T) {
			cards, err := sdk.Cards().GetByName(ctx, "Lightning Bolt", "A25")
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("A25 printings: %d", len(cards))
		})

		t.Run("GetByUUID", func(t *testing.T) {
			if boltUUID == "" {
				t.Skip("no UUID from GetByName")
			}
			card, err := sdk.Cards().GetByUUID(ctx, boltUUID)
			if err != nil {
				t.Fatal(err)
			}
			if card == nil {
				t.Fatal("expected card, got nil")
			}
			if card.Name != "Lightning Bolt" {
				t.Fatalf("expected Lightning Bolt, got %s", card.Name)
			}
		})

		t.Run("GetByUUIDNotFound", func(t *testing.T) {
			card, err := sdk.Cards().GetByUUID(ctx, "00000000-0000-0000-0000-000000000000")
			if err != nil {
				t.Fatal(err)
			}
			if card != nil {
				t.Fatal("expected nil")
			}
		})

		t.Run("GetByUUIDs", func(t *testing.T) {
			cards, err := sdk.Cards().GetByUUIDs(ctx, []string{})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) != 0 {
				t.Fatalf("expected 0, got %d", len(cards))
			}
		})

		t.Run("SearchNameLike", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{Name: "Lightning%", Limit: 10})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
			t.Logf("found %d", len(cards))
		})

		t.Run("SearchExactName", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{Name: "Lightning Bolt", Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchColors", func(t *testing.T) {
			mv := 1.0
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{Colors: []string{"R"}, ManaValue: &mv, Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
			t.Logf("found %d", len(cards))
		})

		t.Run("SearchColorIdentity", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{ColorIdentity: []string{"W", "U"}, Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchTypes", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{Types: "Creature", Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchRarity", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{Rarity: "mythic", Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchText", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{Text: "draw a card", Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchTextRegex", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{TextRegex: "deals \\d+ damage", Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchPowerToughness", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{Power: "4", Toughness: "4", Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchManaValueRange", func(t *testing.T) {
			gte := 1.0
			lte := 2.0
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{ManaValueGTE: &gte, ManaValueLTE: &lte, Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchArtist", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{Artist: "Christopher Moeller", Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchKeyword", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{Keyword: "Flying", Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchLayout", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{Layout: "split", Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchIsPromo", func(t *testing.T) {
			promo := true
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{IsPromo: &promo, Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchAvailability", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{Availability: "mtgo", Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchSetCode", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{SetCode: "MH3", Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchSetType", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{SetType: "expansion", Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchLegalIn", func(t *testing.T) {
			lte := 2.0
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{LegalIn: "modern", ManaValueLTE: &lte, Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchCombinedFilters", func(t *testing.T) {
			lte := 3.0
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{Colors: []string{"R"}, Rarity: "rare", ManaValueLTE: &lte, Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchPagination", func(t *testing.T) {
			p1, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{Name: "Lightning%", Limit: 3, Offset: 0})
			if err != nil {
				t.Fatal(err)
			}
			p2, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{Name: "Lightning%", Limit: 3, Offset: 3})
			if err != nil {
				t.Fatal(err)
			}
			if len(p1) == 0 || len(p2) == 0 {
				t.Fatal("expected results on both pages")
			}
			if p1[0].UUID == p2[0].UUID {
				t.Fatal("pages should have different cards")
			}
		})

		t.Run("SearchLocalizedName", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{LocalizedName: "Blitzschlag", Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
			t.Logf("found %d, name=%s", len(cards), cards[0].Name)
		})

		t.Run("SearchFuzzyName", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{FuzzyName: "Ligtning Bolt"})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
			if cards[0].Name != "Lightning Bolt" {
				t.Fatalf("expected Lightning Bolt first, got %s", cards[0].Name)
			}
		})

		t.Run("Random", func(t *testing.T) {
			cards, err := sdk.Cards().Random(ctx, 3)
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) != 3 {
				t.Fatalf("expected 3, got %d", len(cards))
			}
			t.Logf("random: %v", []string{cards[0].Name, cards[1].Name, cards[2].Name})
		})

		t.Run("Count", func(t *testing.T) {
			count, err := sdk.Cards().Count(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if count < 1000 {
				t.Fatalf("expected >1000 cards, got %d", count)
			}
			t.Logf("total cards: %d", count)
		})

		t.Run("CountFiltered", func(t *testing.T) {
			count, err := sdk.Cards().Count(ctx, queries.Filter{Column: "rarity", Value: "mythic"})
			if err != nil {
				t.Fatal(err)
			}
			if count == 0 {
				t.Fatal("expected >0 mythic cards")
			}
			t.Logf("mythic cards: %d", count)
		})

		t.Run("GetPrintings", func(t *testing.T) {
			cards, err := sdk.Cards().GetPrintings(ctx, "Counterspell")
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) < 5 {
				t.Fatalf("expected >5 printings, got %d", len(cards))
			}
			t.Logf("Counterspell printings: %d", len(cards))
		})

		t.Run("GetAtomic", func(t *testing.T) {
			atomic, err := sdk.Cards().GetAtomic(ctx, "Lightning Bolt")
			if err != nil {
				t.Fatal(err)
			}
			if len(atomic) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("GetAtomicFaceName", func(t *testing.T) {
			atomic, err := sdk.Cards().GetAtomic(ctx, "Fire")
			if err != nil {
				t.Fatal(err)
			}
			if len(atomic) == 0 {
				t.Fatal("expected results for face name 'Fire'")
			}
			t.Logf("layout=%s", atomic[0].Layout)
		})

		t.Run("FindByScryfallID", func(t *testing.T) {
			if boltUUID == "" {
				t.Skip("no UUID")
			}
			cards, err := sdk.Cards().FindByScryfallID(ctx, boltUUID)
			if err != nil {
				t.Fatal(err)
			}
			// May or may not match since UUID != scryfallId
			t.Logf("found %d cards by scryfallId", len(cards))
		})

		t.Run("SearchEmpty", func(t *testing.T) {
			cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{Name: "XYZ_NONEXISTENT_12345"})
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) != 0 {
				t.Fatalf("expected 0, got %d", len(cards))
			}
		})
	})

	// ══════════════════════════════════════════════════════════
	//  TOKENS
	// ══════════════════════════════════════════════════════════
	t.Run("Tokens", func(t *testing.T) {
		t.Run("Count", func(t *testing.T) {
			count, err := sdk.Tokens().Count(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if count == 0 {
				t.Fatal("expected >0 tokens")
			}
			t.Logf("total tokens: %d", count)
		})

		t.Run("SearchByName", func(t *testing.T) {
			tokens, err := sdk.Tokens().Search(ctx, queries.SearchTokensParams{Name: "%Soldier%", Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(tokens) == 0 {
				t.Fatal("expected results")
			}
			t.Logf("found %d", len(tokens))
		})

		t.Run("SearchByColors", func(t *testing.T) {
			tokens, err := sdk.Tokens().Search(ctx, queries.SearchTokensParams{Colors: []string{"W"}, Limit: 5})
			if err != nil {
				t.Fatal(err)
			}
			if len(tokens) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("GetByName", func(t *testing.T) {
			tokens, err := sdk.Tokens().GetByName(ctx, "Soldier")
			if err != nil {
				t.Fatal(err)
			}
			if len(tokens) == 0 {
				t.Fatal("expected results")
			}
			t.Logf("found %d Soldier tokens", len(tokens))
		})

		t.Run("ForSet", func(t *testing.T) {
			// Find a set that has tokens
			rows, err := sdk.SQL(ctx, "SELECT DISTINCT setCode FROM tokens LIMIT 1")
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) == 0 {
				t.Skip("no token data")
			}
			setCode := rows[0]["setCode"].(string)
			tokens, err := sdk.Tokens().ForSet(ctx, setCode)
			if err != nil {
				t.Fatal(err)
			}
			if len(tokens) == 0 {
				t.Fatalf("expected tokens for set %s", setCode)
			}
			t.Logf("set=%s, found %d tokens", setCode, len(tokens))
		})

		t.Run("GetByUUIDs", func(t *testing.T) {
			tokens, err := sdk.Tokens().GetByUUIDs(ctx, []string{})
			if err != nil {
				t.Fatal(err)
			}
			if len(tokens) != 0 {
				t.Fatalf("expected 0, got %d", len(tokens))
			}
		})
	})

	// ══════════════════════════════════════════════════════════
	//  SETS
	// ══════════════════════════════════════════════════════════
	t.Run("Sets", func(t *testing.T) {
		t.Run("Get", func(t *testing.T) {
			s, err := sdk.Sets().Get(ctx, "MH3")
			if err != nil {
				t.Fatal(err)
			}
			if s == nil {
				t.Fatal("expected set")
			}
			if !strings.Contains(s.Name, "Horizons") {
				t.Fatalf("expected Horizons in name, got %s", s.Name)
			}
			t.Logf("name=%s, size=%d", s.Name, s.TotalSetSize)
		})

		t.Run("GetNotFound", func(t *testing.T) {
			s, err := sdk.Sets().Get(ctx, "ZZZZZ")
			if err != nil {
				t.Fatal(err)
			}
			if s != nil {
				t.Fatal("expected nil")
			}
		})

		t.Run("List", func(t *testing.T) {
			sets, err := sdk.Sets().List(ctx, queries.ListSetsParams{Limit: 10})
			if err != nil {
				t.Fatal(err)
			}
			if len(sets) == 0 {
				t.Fatal("expected results")
			}
			t.Logf("found %d sets", len(sets))
		})

		t.Run("ListByType", func(t *testing.T) {
			sets, err := sdk.Sets().List(ctx, queries.ListSetsParams{SetType: "expansion", Limit: 10})
			if err != nil {
				t.Fatal(err)
			}
			if len(sets) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("Search", func(t *testing.T) {
			sets, err := sdk.Sets().Search(ctx, queries.SearchSetsParams{Name: "Horizons"})
			if err != nil {
				t.Fatal(err)
			}
			if len(sets) == 0 {
				t.Fatal("expected results")
			}
			t.Logf("found %d sets matching 'Horizons'", len(sets))
		})

		t.Run("SearchByType", func(t *testing.T) {
			sets, err := sdk.Sets().Search(ctx, queries.SearchSetsParams{SetType: "masters", Limit: 10})
			if err != nil {
				t.Fatal(err)
			}
			if len(sets) == 0 {
				t.Fatal("expected results")
			}
		})

		t.Run("SearchByYear", func(t *testing.T) {
			year := 2024
			sets, err := sdk.Sets().Search(ctx, queries.SearchSetsParams{ReleaseYear: &year, Limit: 10})
			if err != nil {
				t.Fatal(err)
			}
			if len(sets) == 0 {
				t.Fatal("expected results")
			}
			t.Logf("found %d sets from 2024", len(sets))
		})

		t.Run("Count", func(t *testing.T) {
			count, err := sdk.Sets().Count(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if count < 100 {
				t.Fatalf("expected >100 sets, got %d", count)
			}
			t.Logf("total sets: %d", count)
		})
	})

	// ══════════════════════════════════════════════════════════
	//  IDENTIFIERS
	// ══════════════════════════════════════════════════════════
	t.Run("Identifiers", func(t *testing.T) {
		if boltUUID == "" {
			t.Skip("no UUID")
		}

		t.Run("GetIdentifiers", func(t *testing.T) {
			ids, err := sdk.Identifiers().GetIdentifiers(ctx, boltUUID)
			if err != nil {
				t.Fatal(err)
			}
			if ids == nil {
				t.Skip("no identifiers for this card")
			}
			t.Logf("identifier keys: %v", func() []string {
				keys := make([]string, 0)
				for k := range ids {
					keys = append(keys, k)
				}
				return keys
			}())

			// Test specific finders with real IDs
			if scryfallID, ok := ids["scryfallId"].(string); ok && scryfallID != "" {
				cards, err := sdk.Identifiers().FindByScryfallID(ctx, scryfallID)
				if err != nil {
					t.Fatal(err)
				}
				if len(cards) == 0 {
					t.Fatal("expected results for FindByScryfallID")
				}
			}

			if oracleID, ok := ids["scryfallOracleId"].(string); ok && oracleID != "" {
				cards, err := sdk.Identifiers().FindByScryfallOracleID(ctx, oracleID)
				if err != nil {
					t.Fatal(err)
				}
				if len(cards) == 0 {
					t.Fatal("expected results for FindByScryfallOracleID")
				}
			}

			if tcgID, ok := ids["tcgplayerProductId"].(string); ok && tcgID != "" {
				cards, err := sdk.Identifiers().FindByTCGPlayerID(ctx, tcgID)
				if err != nil {
					t.Fatal(err)
				}
				if len(cards) == 0 {
					t.Fatal("expected results for FindByTCGPlayerID")
				}
			}
		})

		t.Run("FindByGeneric", func(t *testing.T) {
			ids, _ := sdk.Identifiers().GetIdentifiers(ctx, boltUUID)
			if ids == nil {
				t.Skip("no identifiers")
			}
			if scryfallID, ok := ids["scryfallId"].(string); ok && scryfallID != "" {
				cards, err := sdk.Identifiers().FindBy(ctx, "scryfallId", scryfallID)
				if err != nil {
					t.Fatal(err)
				}
				if len(cards) == 0 {
					t.Fatal("expected results")
				}
			}
		})

		t.Run("FindByInvalidColumn", func(t *testing.T) {
			_, err := sdk.Identifiers().FindBy(ctx, "invalidColumn", "123")
			if err == nil {
				t.Fatal("expected error for invalid column")
			}
			if !strings.Contains(err.Error(), "unknown identifier type") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	})

	// ══════════════════════════════════════════════════════════
	//  LEGALITIES
	// ══════════════════════════════════════════════════════════
	t.Run("Legalities", func(t *testing.T) {
		if boltUUID == "" {
			t.Skip("no UUID")
		}

		t.Run("FormatsForCard", func(t *testing.T) {
			formats, err := sdk.Legalities().FormatsForCard(ctx, boltUUID)
			if err != nil {
				t.Fatal(err)
			}
			if len(formats) == 0 {
				t.Fatal("expected formats")
			}
			t.Logf("formats: %v", formats)
		})

		t.Run("IsLegal", func(t *testing.T) {
			legal, err := sdk.Legalities().IsLegal(ctx, boltUUID, "modern")
			if err != nil {
				t.Fatal(err)
			}
			if !legal {
				t.Fatal("expected Lightning Bolt to be modern legal")
			}

			legalFake, err := sdk.Legalities().IsLegal(ctx, boltUUID, "nonexistent_format")
			if err != nil {
				t.Fatal(err)
			}
			if legalFake {
				t.Fatal("expected not legal in nonexistent format")
			}
		})

		t.Run("LegalIn", func(t *testing.T) {
			cards, err := sdk.Legalities().LegalIn(ctx, "modern", 5)
			if err != nil {
				t.Fatal(err)
			}
			if len(cards) == 0 {
				t.Fatal("expected results")
			}
			t.Logf("modern legal cards: %d", len(cards))
		})

		t.Run("BannedIn", func(t *testing.T) {
			cards, err := sdk.Legalities().BannedIn(ctx, "modern", 5)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("modern banned: %d", len(cards))
		})

		t.Run("RestrictedIn", func(t *testing.T) {
			cards, err := sdk.Legalities().RestrictedIn(ctx, "vintage", 5)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("vintage restricted: %d", len(cards))
		})

		t.Run("SuspendedIn", func(t *testing.T) {
			cards, err := sdk.Legalities().SuspendedIn(ctx, "historic", 5)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("historic suspended: %d", len(cards))
		})

		t.Run("NotLegalIn", func(t *testing.T) {
			cards, err := sdk.Legalities().NotLegalIn(ctx, "standard", 5)
			if err != nil {
				t.Fatal(err)
			}
			// May be 0 if "Not Legal" status is not tracked in legalities table
			t.Logf("not legal in standard: %d", len(cards))
		})
	})

	// ══════════════════════════════════════════════════════════
	//  PRICES
	// ══════════════════════════════════════════════════════════
	t.Run("Prices", func(t *testing.T) {
		if boltUUID == "" {
			t.Skip("no UUID")
		}

		t.Run("Get", func(t *testing.T) {
			price, err := sdk.Prices().Get(ctx, boltUUID)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("price data: %v", price != nil)
		})

		t.Run("Today", func(t *testing.T) {
			rows, err := sdk.Prices().Today(ctx, boltUUID)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("today rows: %d", len(rows))
		})

		t.Run("TodayWithFilter", func(t *testing.T) {
			rows, err := sdk.Prices().Today(ctx, boltUUID, queries.WithPriceProvider("tcgplayer"))
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("tcgplayer today rows: %d", len(rows))
		})

		t.Run("History", func(t *testing.T) {
			rows, err := sdk.Prices().History(ctx, boltUUID)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("history rows: %d", len(rows))
		})

		t.Run("PriceTrend", func(t *testing.T) {
			trend, err := sdk.Prices().PriceTrend(ctx, boltUUID)
			if err != nil {
				t.Fatal(err)
			}
			if trend != nil {
				t.Logf("trend: min=%.2f max=%.2f avg=%.2f points=%d",
					trend.MinPrice, trend.MaxPrice, trend.AvgPrice, trend.DataPoints)
			}
		})

		t.Run("CheapestPrinting", func(t *testing.T) {
			cheapest, err := sdk.Prices().CheapestPrinting(ctx, "Lightning Bolt")
			if err != nil {
				t.Fatal(err)
			}
			if cheapest != nil {
				t.Logf("cheapest: set=%v price=%v", cheapest["setCode"], cheapest["price"])
			}
		})
	})

	// ══════════════════════════════════════════════════════════
	//  DECKS
	// ══════════════════════════════════════════════════════════
	t.Run("Decks", func(t *testing.T) {
		t.Run("Count", func(t *testing.T) {
			count, err := sdk.Decks().Count(ctx)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("total decks: %d", count)
		})

		t.Run("List", func(t *testing.T) {
			decks, err := sdk.Decks().List(ctx, queries.ListDecksParams{})
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("listed %d decks", len(decks))
		})

		t.Run("Search", func(t *testing.T) {
			decks, err := sdk.Decks().Search(ctx, queries.SearchDecksParams{Name: "Commander"})
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("search 'Commander': %d results", len(decks))
		})
	})

	// ══════════════════════════════════════════════════════════
	//  SKUS
	// ══════════════════════════════════════════════════════════
	t.Run("SKUs", func(t *testing.T) {
		if boltUUID == "" {
			t.Skip("no UUID")
		}

		t.Run("Get", func(t *testing.T) {
			skus, err := sdk.Skus().Get(ctx, boltUUID)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("SKUs for bolt: %d", len(skus))

			if len(skus) > 0 {
				t.Run("FindBySkuID", func(t *testing.T) {
					allSkus, err := sdk.Skus().Get(ctx, boltUUID)
					if err != nil || len(allSkus) == 0 {
						t.Skip("no SKUs available")
					}
					skuID := allSkus[0].SkuId
					result, err := sdk.Skus().FindBySkuID(ctx, skuID)
					if err != nil {
						t.Fatal(err)
					}
					if result == nil {
						t.Fatal("expected result")
					}
					t.Logf("found SKU %d", skuID)
				})

				t.Run("FindByProductID", func(t *testing.T) {
					allSkus, err := sdk.Skus().Get(ctx, boltUUID)
					if err != nil || len(allSkus) == 0 {
						t.Skip("no SKUs available")
					}
					prodID := allSkus[0].ProductId
					results, err := sdk.Skus().FindByProductID(ctx, prodID)
					if err != nil {
						t.Fatal(err)
					}
					if len(results) == 0 {
						t.Fatal("expected results")
					}
					t.Logf("product %d: %d SKUs", prodID, len(results))
				})
			}
		})
	})

	// ══════════════════════════════════════════════════════════
	//  ENUMS
	// ══════════════════════════════════════════════════════════
	t.Run("Enums", func(t *testing.T) {
		t.Run("Keywords", func(t *testing.T) {
			kw, err := sdk.Enums().Keywords(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(kw) == 0 {
				t.Fatal("expected keywords")
			}
			keys := make([]string, 0)
			for k := range kw {
				keys = append(keys, k)
			}
			t.Logf("keyword categories: %v", keys)
		})

		t.Run("CardTypes", func(t *testing.T) {
			ct, err := sdk.Enums().CardTypes(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(ct) == 0 {
				t.Fatal("expected card types")
			}
		})

		t.Run("EnumValues", func(t *testing.T) {
			ev, err := sdk.Enums().EnumValues(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(ev) == 0 {
				t.Fatal("expected enum values")
			}
		})
	})

	// ══════════════════════════════════════════════════════════
	//  SEALED
	// ══════════════════════════════════════════════════════════
	t.Run("Sealed", func(t *testing.T) {
		t.Run("List", func(t *testing.T) {
			products, err := sdk.Sealed().List(ctx, queries.ListSealedParams{})
			if err != nil {
				t.Fatal(err)
			}
			// May be empty with flat parquet
			t.Logf("sealed products: %d", len(products))
		})

		t.Run("ListBySet", func(t *testing.T) {
			products, err := sdk.Sealed().List(ctx, queries.ListSealedParams{SetCode: "MH3"})
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("MH3 sealed products: %d", len(products))
		})

		t.Run("GetNotFound", func(t *testing.T) {
			product, err := sdk.Sealed().Get(ctx, "00000000-0000-0000-0000-000000000000")
			if err != nil {
				t.Fatal(err)
			}
			if product != nil {
				t.Logf("unexpected sealed product (may be valid in full data)")
			}
		})
	})

	// ══════════════════════════════════════════════════════════
	//  BOOSTER
	// ══════════════════════════════════════════════════════════
	t.Run("Booster", func(t *testing.T) {
		t.Run("AvailableTypes", func(t *testing.T) {
			types, err := sdk.Booster().AvailableTypes(ctx, "MH3")
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("booster types for MH3: %v", types)
		})

		t.Run("SheetContents", func(t *testing.T) {
			contents, err := sdk.Booster().SheetContents(ctx, "MH3", "draft", "common")
			if err != nil {
				t.Fatal(err)
			}
			// May be nil with flat parquet
			if contents != nil {
				t.Logf("common sheet has %d cards", len(contents))
			}
		})

		t.Run("OpenPack", func(t *testing.T) {
			pack, err := sdk.Booster().OpenPack(ctx, "MH3", "draft")
			if err != nil {
				// Expected with flat parquet
				t.Logf("open_pack error (expected with flat parquet): %v", err)
				return
			}
			if pack != nil {
				t.Logf("pack has %d cards", len(pack))
			}
		})
	})

	// ══════════════════════════════════════════════════════════
	//  RAW SQL
	// ══════════════════════════════════════════════════════════
	t.Run("SQL", func(t *testing.T) {
		t.Run("Count", func(t *testing.T) {
			rows, err := sdk.SQL(ctx, "SELECT COUNT(*) AS cnt FROM cards")
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("SQL count: %v", rows[0]["cnt"])
		})

		t.Run("Params", func(t *testing.T) {
			rows, err := sdk.SQL(ctx, "SELECT name FROM cards WHERE manaValue = $1 LIMIT $2", 1.0, 5)
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) == 0 {
				t.Fatal("expected results")
			}
			t.Logf("found %d cards with mana value 1", len(rows))
		})

		t.Run("Join", func(t *testing.T) {
			rows, err := sdk.SQL(ctx,
				"SELECT c.name, s.name AS setName "+
					"FROM cards c JOIN sets s ON c.setCode = s.code "+
					"LIMIT 3")
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) == 0 {
				t.Fatal("expected results")
			}
			if _, ok := rows[0]["setName"]; !ok {
				t.Fatal("expected setName column")
			}
		})
	})

	// ══════════════════════════════════════════════════════════
	//  VIEWS
	// ══════════════════════════════════════════════════════════
	t.Run("ViewsGrew", func(t *testing.T) {
		viewsAfter := sdk.Views()
		if len(viewsAfter) <= len(viewsBefore) {
			t.Fatalf("expected views to grow: before=%d after=%d", len(viewsBefore), len(viewsAfter))
		}
		t.Logf("views: before=%d after=%d list=%v", len(viewsBefore), len(viewsAfter), viewsAfter)
	})

	// ══════════════════════════════════════════════════════════
	//  REFRESH
	// ══════════════════════════════════════════════════════════
	t.Run("Refresh", func(t *testing.T) {
		stale, err := sdk.Refresh(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("refresh stale=%v", stale)
	})

	fmt.Println("\nSmoke test completed successfully!")
}
