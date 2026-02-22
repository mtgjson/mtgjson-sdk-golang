package mtgjson

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupEnumCache(t *testing.T) *CacheManager {
	t.Helper()
	tmpDir := t.TempDir()

	// Create sample Keywords.json
	keywords := map[string]any{
		"data": map[string]any{
			"abilityWords":  []any{"Addendum", "Alliance", "Batallion"},
			"keywordActions": []any{"Activate", "Attach", "Cast"},
		},
	}
	writeJSON(t, filepath.Join(tmpDir, "Keywords.json"), keywords)

	// Create sample CardTypes.json
	cardTypes := map[string]any{
		"data": map[string]any{
			"creature": map[string]any{
				"subTypes":   []any{"Human", "Elf", "Goblin"},
				"superTypes": []any{"Legendary"},
			},
			"instant": map[string]any{
				"subTypes":   []any{"Arcane", "Trap"},
				"superTypes": []any{},
			},
		},
	}
	writeJSON(t, filepath.Join(tmpDir, "CardTypes.json"), cardTypes)

	// Create sample EnumValues.json
	enumValues := map[string]any{
		"data": map[string]any{
			"colors":     []any{"B", "G", "R", "U", "W"},
			"rarities":   []any{"common", "uncommon", "rare", "mythic"},
			"frameEffects": []any{"colorshifted", "extendedart", "inverted"},
		},
	}
	writeJSON(t, filepath.Join(tmpDir, "EnumValues.json"), enumValues)

	cfg := defaultConfig()
	cfg.cacheDir = tmpDir
	cfg.offline = true
	cache, err := newCacheManager(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cache.Close() })
	return cache
}

func writeJSON(t *testing.T, path string, data any) {
	t.Helper()
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestKeywords(t *testing.T) {
	cache := setupEnumCache(t)
	eq := newEnumQuery(cache)
	ctx := context.Background()

	kw, err := eq.Keywords(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if kw == nil {
		t.Fatal("expected keywords, got nil")
	}
	aw, ok := kw["abilityWords"].([]any)
	if !ok {
		t.Fatal("expected abilityWords array")
	}
	if len(aw) != 3 {
		t.Fatalf("expected 3 abilityWords, got %d", len(aw))
	}
}

func TestCardTypes(t *testing.T) {
	cache := setupEnumCache(t)
	eq := newEnumQuery(cache)
	ctx := context.Background()

	types, err := eq.CardTypes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if types == nil {
		t.Fatal("expected card types, got nil")
	}
	creature, ok := types["creature"].(map[string]any)
	if !ok {
		t.Fatal("expected creature map")
	}
	subs, ok := creature["subTypes"].([]any)
	if !ok {
		t.Fatal("expected subTypes array")
	}
	if len(subs) != 3 {
		t.Fatalf("expected 3 creature subtypes, got %d", len(subs))
	}
}

func TestEnumValues(t *testing.T) {
	cache := setupEnumCache(t)
	eq := newEnumQuery(cache)
	ctx := context.Background()

	vals, err := eq.EnumValues(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if vals == nil {
		t.Fatal("expected enum values, got nil")
	}
	colors, ok := vals["colors"].([]any)
	if !ok {
		t.Fatal("expected colors array")
	}
	if len(colors) != 5 {
		t.Fatalf("expected 5 colors, got %d", len(colors))
	}
}
