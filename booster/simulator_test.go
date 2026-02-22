package booster

import (
	"testing"
)

func TestPickPackWeighted(t *testing.T) {
	boosters := []any{
		map[string]any{"contents": map[string]any{"rare": 1.0, "common": 10.0}, "weight": 7.0},
		map[string]any{"contents": map[string]any{"mythic": 1.0, "common": 10.0}, "weight": 1.0},
	}
	pack := pickPack(boosters)
	if pack == nil {
		t.Fatal("expected non-nil pack")
	}
	if _, ok := pack["contents"]; !ok {
		t.Fatal("expected contents key")
	}
	if _, ok := pack["weight"]; !ok {
		t.Fatal("expected weight key")
	}
}

func TestPickFromSheetBasic(t *testing.T) {
	sheet := map[string]any{
		"cards": map[string]any{
			"uuid-a": 10.0, "uuid-b": 5.0, "uuid-c": 1.0,
		},
		"foil":        false,
		"totalWeight": 16.0,
	}
	picked := pickFromSheet(sheet, 2)
	if len(picked) != 2 {
		t.Fatalf("expected 2 picks, got %d", len(picked))
	}
	validUUIDs := map[string]bool{"uuid-a": true, "uuid-b": true, "uuid-c": true}
	for _, u := range picked {
		if !validUUIDs[u] {
			t.Fatalf("unexpected UUID: %s", u)
		}
	}
}

func TestPickFromSheetNoDuplicates(t *testing.T) {
	sheet := map[string]any{
		"cards": map[string]any{
			"uuid-a": 1.0, "uuid-b": 1.0, "uuid-c": 1.0,
		},
		"foil":        false,
		"totalWeight": 3.0,
	}
	picked := pickFromSheet(sheet, 3)
	if len(picked) != 3 {
		t.Fatalf("expected 3 picks, got %d", len(picked))
	}
	seen := make(map[string]bool)
	for _, u := range picked {
		if seen[u] {
			t.Fatalf("duplicate UUID: %s", u)
		}
		seen[u] = true
	}
}

func TestPickFromSheetWithDuplicates(t *testing.T) {
	sheet := map[string]any{
		"cards": map[string]any{
			"uuid-a": 1.0,
		},
		"foil":            false,
		"totalWeight":     1.0,
		"allowDuplicates": true,
	}
	picked := pickFromSheet(sheet, 3)
	if len(picked) != 3 {
		t.Fatalf("expected 3 picks, got %d", len(picked))
	}
	for _, u := range picked {
		if u != "uuid-a" {
			t.Fatalf("expected uuid-a, got %s", u)
		}
	}
}

func TestPickPackEmpty(t *testing.T) {
	pack := pickPack(nil)
	if pack != nil {
		t.Fatalf("expected nil, got %v", pack)
	}
}

func TestPickFromSheetMoreThanAvailable(t *testing.T) {
	sheet := map[string]any{
		"cards": map[string]any{
			"uuid-a": 1.0, "uuid-b": 1.0,
		},
	}
	picked := pickFromSheet(sheet, 5)
	if len(picked) != 2 {
		t.Fatalf("expected 2 picks (all available), got %d", len(picked))
	}
}
