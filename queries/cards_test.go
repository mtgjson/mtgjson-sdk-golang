package queries

import (
	"context"
	"testing"
)

func TestCardGetByUUID(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	card, err := q.GetByUUID(ctx, "card-uuid-001")
	if err != nil {
		t.Fatal(err)
	}
	if card == nil {
		t.Fatal("expected card, got nil")
	}
	if card.Name != "Lightning Bolt" {
		t.Fatalf("expected Lightning Bolt, got %s", card.Name)
	}
	if card.UUID != "card-uuid-001" {
		t.Fatalf("expected card-uuid-001, got %s", card.UUID)
	}
}

func TestCardGetByUUIDNotFound(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	card, err := q.GetByUUID(ctx, "nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if card != nil {
		t.Fatalf("expected nil, got %v", card)
	}
}

func TestCardGetByName(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.GetByName(ctx, "Lightning Bolt")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].Name != "Lightning Bolt" {
		t.Fatalf("expected Lightning Bolt, got %s", cards[0].Name)
	}
}

func TestCardSearchByNameLike(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{Name: "Lightning%"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
	for _, c := range cards {
		if len(c.Name) < 9 || c.Name[:9] != "Lightning" {
			t.Fatalf("expected name starting with Lightning, got %s", c.Name)
		}
	}
}

func TestCardSearchByRarity(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{Rarity: "uncommon"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 3 {
		t.Fatalf("expected 3 cards, got %d", len(cards))
	}
}

func TestCardSearchByManaValue(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	mv := 1.0
	cards, err := q.Search(ctx, SearchCardsParams{ManaValue: &mv})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].Name != "Lightning Bolt" {
		t.Fatalf("expected Lightning Bolt, got %s", cards[0].Name)
	}
}

func TestCardSearchByColors(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{Colors: []string{"U"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
	for _, c := range cards {
		found := false
		for _, color := range c.Colors {
			if color == "U" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected U in colors, got %v", c.Colors)
		}
	}
}

func TestCardSearchByText(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{Text: "Counter target spell"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
}

func TestCardSearchWithLimit(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
}

func TestCardSearchLegalIn(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{LegalIn: "modern"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	names := make(map[string]bool)
	for _, c := range cards {
		names[c.Name] = true
	}
	if !names["Lightning Bolt"] {
		t.Fatal("expected Lightning Bolt")
	}
	if !names["Counterspell"] {
		t.Fatal("expected Counterspell")
	}
}

func TestCardGetPrintings(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.GetPrintings(ctx, "Counterspell")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 printing")
	}
}

func TestCardRandom(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Random(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
}

func TestCardCount(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	count, err := q.Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("expected 3, got %d", count)
	}
}

func TestCardFindByScryfallID(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.FindByScryfallID(ctx, "scryfall-001")
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

func TestCardSearchByArtist(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{Artist: "Christopher Moeller"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].Name != "Lightning Bolt" {
		t.Fatalf("expected Lightning Bolt, got %s", cards[0].Name)
	}
}

func TestCardSearchByArtistLike(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{Artist: "Zack"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].Name != "Counterspell" {
		t.Fatalf("expected Counterspell, got %s", cards[0].Name)
	}
}

func TestCardSearchByColorIdentity(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{ColorIdentity: []string{"R"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
	for _, c := range cards {
		found := false
		for _, ci := range c.ColorIdentity {
			if ci == "R" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected R in color identity, got %v", c.ColorIdentity)
		}
	}
}

func TestCardSearchByAvailability(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{Availability: "paper"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 3 {
		t.Fatalf("expected 3 cards, got %d", len(cards))
	}
}

func TestCardSearchByLanguage(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{Language: "English"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 3 {
		t.Fatalf("expected 3 cards, got %d", len(cards))
	}
}

func TestCardSearchByLayout(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{Layout: "normal"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
}

func TestCardSearchByLayoutSplit(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{Layout: "split"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].Name != "Fire // Ice" {
		t.Fatalf("expected Fire // Ice, got %s", cards[0].Name)
	}
}

func TestCardSearchBySetType(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{SetType: "masters"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards (A25), got %d", len(cards))
	}
}

func TestCardSearchByManaValueRange(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	gte := 1.0
	lte := 1.5
	cards, err := q.Search(ctx, SearchCardsParams{ManaValueGTE: &gte, ManaValueLTE: &lte})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].Name != "Lightning Bolt" {
		t.Fatalf("expected Lightning Bolt, got %s", cards[0].Name)
	}
}

func TestCardGetAtomicByFaceName(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	result, err := q.GetAtomic(ctx, "Fire")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) < 1 {
		t.Fatal("expected at least 1 result")
	}
	if result[0].FaceName == nil || *result[0].FaceName != "Fire" {
		t.Fatalf("expected faceName=Fire, got %v", result[0].FaceName)
	}
	if result[0].Layout != "split" {
		t.Fatalf("expected split layout, got %s", result[0].Layout)
	}
}

func TestCardGetAtomicNotFound(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	result, err := q.GetAtomic(ctx, "Nonexistent Card Face")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty, got %d", len(result))
	}
}

func TestCardGetByUUIDs(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.GetByUUIDs(ctx, []string{"card-uuid-001", "card-uuid-002"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	names := make(map[string]bool)
	for _, c := range cards {
		names[c.Name] = true
	}
	if !names["Lightning Bolt"] || !names["Counterspell"] {
		t.Fatalf("expected Lightning Bolt and Counterspell, got %v", names)
	}
}

func TestCardGetByUUIDsEmpty(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.GetByUUIDs(ctx, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 0 {
		t.Fatalf("expected 0, got %d", len(cards))
	}
}

func TestCardGetByUUIDsNonexistent(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.GetByUUIDs(ctx, []string{"no-such-uuid"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 0 {
		t.Fatalf("expected 0, got %d", len(cards))
	}
}

func TestCardGetByUUIDsPartialMatch(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.GetByUUIDs(ctx, []string{"card-uuid-001", "no-such-uuid"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1, got %d", len(cards))
	}
	if cards[0].Name != "Lightning Bolt" {
		t.Fatalf("expected Lightning Bolt, got %s", cards[0].Name)
	}
}

func TestCardSearchLocalizedNameExact(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{LocalizedName: "Blitzschlag"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].Name != "Lightning Bolt" {
		t.Fatalf("expected Lightning Bolt, got %s", cards[0].Name)
	}
}

func TestCardSearchLocalizedNameLike(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{LocalizedName: "Blitz%"})
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

func TestCardSearchLocalizedNameNotFound(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{LocalizedName: "Nonexistent"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 0 {
		t.Fatalf("expected 0, got %d", len(cards))
	}
}

func TestCardSearchLocalizedNameFrench(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{LocalizedName: "Foudre"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].Name != "Lightning Bolt" {
		t.Fatalf("expected Lightning Bolt, got %s", cards[0].Name)
	}
}

func TestCardSearchTextRegex(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{TextRegex: "deals \\d+ damage"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
	found := false
	for _, c := range cards {
		if c.Name == "Lightning Bolt" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected Lightning Bolt in results")
	}
}

func TestCardSearchTextRegexNoMatch(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{TextRegex: "^This card does nothing$"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 0 {
		t.Fatalf("expected 0, got %d", len(cards))
	}
}

func TestCardSearchFuzzyName(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{FuzzyName: "Ligtning Bolt"})
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

func TestCardSearchFuzzyNameExact(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{FuzzyName: "Lightning Bolt"})
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

func TestCardSearchFuzzyNameNoMatch(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{FuzzyName: "zzzzzzzzzzzzz"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 0 {
		t.Fatalf("expected 0, got %d", len(cards))
	}
}

func TestCardSearchFuzzyNameOrderedBySimilarity(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewCardQuery(conn)
	ctx := context.Background()

	cards, err := q.Search(ctx, SearchCardsParams{FuzzyName: "Countrsepll"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 1 {
		t.Fatal("expected at least 1 card")
	}
	if cards[0].Name != "Counterspell" {
		t.Fatalf("expected Counterspell first, got %s", cards[0].Name)
	}
}
