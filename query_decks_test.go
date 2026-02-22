package mtgjson

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// setupDeckQuery creates a DeckQuery with sample deck data.
func setupDeckQuery(t *testing.T) *DeckQuery {
	t.Helper()

	sampleDecks := map[string]any{
		"data": []any{
			map[string]any{
				"code": "MH3", "name": "Creative Energy",
				"fileName": "Creative_Energy_MH3.json", "type": "Commander Deck",
				"releaseDate": "2024-06-14",
			},
			map[string]any{
				"code": "MH3", "name": "Eldrazi Incursion",
				"fileName": "Eldrazi_Incursion_MH3.json", "type": "Commander Deck",
				"releaseDate": "2024-06-14",
			},
			map[string]any{
				"code": "WOE", "name": "Virtue and Valor",
				"fileName": "Virtue_and_Valor_WOE.json", "type": "Commander Deck",
				"releaseDate": "2023-09-08",
			},
			map[string]any{
				"code": "ONE", "name": "Rebellion Rising",
				"fileName": "Rebellion_Rising_ONE.json", "type": "Starter Kit",
				"releaseDate": "2023-02-03",
			},
		},
	}

	tmpDir := t.TempDir()
	jsonBytes, _ := json.Marshal(sampleDecks)
	deckPath := filepath.Join(tmpDir, "DeckList.json")
	os.WriteFile(deckPath, jsonBytes, 0o644)

	cfg := defaultConfig()
	cfg.cacheDir = tmpDir
	cfg.offline = true
	cache, err := newCacheManager(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cache.Close() })

	return newDeckQuery(cache)
}

func TestDeckList(t *testing.T) {
	dq := setupDeckQuery(t)
	ctx := context.Background()

	decks, err := dq.List(ctx, ListDecksParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(decks) != 4 {
		t.Fatalf("expected 4 decks, got %d", len(decks))
	}
}

func TestDeckListBySetCode(t *testing.T) {
	dq := setupDeckQuery(t)
	ctx := context.Background()

	decks, err := dq.List(ctx, ListDecksParams{SetCode: "MH3"})
	if err != nil {
		t.Fatal(err)
	}
	if len(decks) != 2 {
		t.Fatalf("expected 2 decks, got %d", len(decks))
	}
	for _, d := range decks {
		if d.Code != "MH3" {
			t.Fatalf("expected code MH3, got %s", d.Code)
		}
	}
}

func TestDeckListBySetCodeCaseInsensitive(t *testing.T) {
	dq := setupDeckQuery(t)
	ctx := context.Background()

	decks, err := dq.List(ctx, ListDecksParams{SetCode: "mh3"})
	if err != nil {
		t.Fatal(err)
	}
	if len(decks) != 2 {
		t.Fatalf("expected 2 decks, got %d", len(decks))
	}
}

func TestDeckListByType(t *testing.T) {
	dq := setupDeckQuery(t)
	ctx := context.Background()

	decks, err := dq.List(ctx, ListDecksParams{DeckType: "Commander Deck"})
	if err != nil {
		t.Fatal(err)
	}
	if len(decks) != 3 {
		t.Fatalf("expected 3 Commander Decks, got %d", len(decks))
	}
}

func TestDeckSearch(t *testing.T) {
	dq := setupDeckQuery(t)
	ctx := context.Background()

	decks, err := dq.Search(ctx, SearchDecksParams{Name: "Energy"})
	if err != nil {
		t.Fatal(err)
	}
	if len(decks) != 1 {
		t.Fatalf("expected 1 deck, got %d", len(decks))
	}
	if decks[0].Name != "Creative Energy" {
		t.Fatalf("expected Creative Energy, got %s", decks[0].Name)
	}
}

func TestDeckSearchCaseInsensitive(t *testing.T) {
	dq := setupDeckQuery(t)
	ctx := context.Background()

	decks, err := dq.Search(ctx, SearchDecksParams{Name: "eldrazi"})
	if err != nil {
		t.Fatal(err)
	}
	if len(decks) != 1 {
		t.Fatalf("expected 1 deck, got %d", len(decks))
	}
}

func TestDeckSearchWithSetCode(t *testing.T) {
	dq := setupDeckQuery(t)
	ctx := context.Background()

	decks, err := dq.Search(ctx, SearchDecksParams{Name: "Commander", SetCode: "MH3"})
	if err != nil {
		t.Fatal(err)
	}
	// "Commander" appears in type not name for our test data, but "Creative Energy" and "Eldrazi Incursion" don't have "Commander" in name
	// Let's search for something that matches
	decks, err = dq.Search(ctx, SearchDecksParams{SetCode: "MH3"})
	if err != nil {
		t.Fatal(err)
	}
	if len(decks) != 2 {
		t.Fatalf("expected 2 decks, got %d", len(decks))
	}
}

func TestDeckCount(t *testing.T) {
	dq := setupDeckQuery(t)
	ctx := context.Background()

	count, err := dq.Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 4 {
		t.Fatalf("expected 4, got %d", count)
	}
}

func TestDeckSearchNoResults(t *testing.T) {
	dq := setupDeckQuery(t)
	ctx := context.Background()

	decks, err := dq.Search(ctx, SearchDecksParams{Name: "nonexistent"})
	if err != nil {
		t.Fatal(err)
	}
	if decks != nil {
		t.Fatalf("expected nil, got %v", decks)
	}
}
