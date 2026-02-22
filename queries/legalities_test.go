package queries

import (
	"context"
	"testing"
)

func TestFormatsForCard(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewLegalityQuery(conn)
	ctx := context.Background()

	formats, err := q.FormatsForCard(ctx, "card-uuid-001")
	if err != nil {
		t.Fatal(err)
	}
	if formats["modern"] != "Legal" {
		t.Fatalf("expected modern=Legal, got %s", formats["modern"])
	}
	if formats["vintage"] != "Restricted" {
		t.Fatalf("expected vintage=Restricted, got %s", formats["vintage"])
	}
}

func TestLegalIn(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewLegalityQuery(conn)
	ctx := context.Background()

	cards, err := q.LegalIn(ctx, "modern")
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

func TestIsLegal(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewLegalityQuery(conn)
	ctx := context.Background()

	legal, err := q.IsLegal(ctx, "card-uuid-001", "modern")
	if err != nil {
		t.Fatal(err)
	}
	if !legal {
		t.Fatal("expected true")
	}

	legal, err = q.IsLegal(ctx, "card-uuid-001", "standard")
	if err != nil {
		t.Fatal(err)
	}
	if legal {
		t.Fatal("expected false")
	}
}

func TestBannedIn(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewLegalityQuery(conn)
	ctx := context.Background()

	cards, err := q.BannedIn(ctx, "modern")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 0 {
		t.Fatalf("expected 0, got %d", len(cards))
	}
}

func TestRestrictedIn(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewLegalityQuery(conn)
	ctx := context.Background()

	cards, err := q.RestrictedIn(ctx, "vintage")
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

func TestSuspendedIn(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewLegalityQuery(conn)
	ctx := context.Background()

	cards, err := q.SuspendedIn(ctx, "historic")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1, got %d", len(cards))
	}
	if cards[0].Name != "Counterspell" {
		t.Fatalf("expected Counterspell, got %s", cards[0].Name)
	}
}

func TestNotLegalIn(t *testing.T) {
	conn := setupSampleDB(t)
	q := NewLegalityQuery(conn)
	ctx := context.Background()

	cards, err := q.NotLegalIn(ctx, "standard")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2, got %d", len(cards))
	}
}
