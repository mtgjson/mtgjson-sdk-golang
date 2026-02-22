package mtgjson

import (
	"context"
	"testing"
)

func TestSetGet(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	s, err := sdk.Sets().Get(ctx, "A25")
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("expected set, got nil")
	}
	if s.Name != "Masters 25" {
		t.Fatalf("expected Masters 25, got %s", s.Name)
	}
	if s.Code != "A25" {
		t.Fatalf("expected A25, got %s", s.Code)
	}
}

func TestSetGetCaseInsensitive(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	s, err := sdk.Sets().Get(ctx, "a25")
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("expected set, got nil")
	}
	if s.Code != "A25" {
		t.Fatalf("expected A25, got %s", s.Code)
	}
}

func TestSetGetNotFound(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	s, err := sdk.Sets().Get(ctx, "XXXXX")
	if err != nil {
		t.Fatal(err)
	}
	if s != nil {
		t.Fatalf("expected nil, got %v", s)
	}
}

func TestSetList(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	sets, err := sdk.Sets().List(ctx, ListSetsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sets) != 2 {
		t.Fatalf("expected 2 sets, got %d", len(sets))
	}
}

func TestSetListByType(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	sets, err := sdk.Sets().List(ctx, ListSetsParams{SetType: "masters"})
	if err != nil {
		t.Fatal(err)
	}
	if len(sets) != 1 {
		t.Fatalf("expected 1 set, got %d", len(sets))
	}
	if sets[0].Code != "A25" {
		t.Fatalf("expected A25, got %s", sets[0].Code)
	}
}

func TestSetSearch(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	sets, err := sdk.Sets().Search(ctx, SearchSetsParams{Name: "Horizons"})
	if err != nil {
		t.Fatal(err)
	}
	if len(sets) < 1 {
		t.Fatal("expected at least 1 set")
	}
	found := false
	for _, s := range sets {
		if containsStr(s.Name, "Horizons") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected a set with Horizons in name")
	}
}

func TestSetCount(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	count, err := sdk.Sets().Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

func TestSetFinancialSummary(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	// Load price data
	if err := sdk.conn.RegisterTableFromData(ctx, "prices_today", samplePrices); err != nil {
		t.Fatal(err)
	}

	summary, err := sdk.Sets().GetFinancialSummary(ctx, "A25")
	if err != nil {
		t.Fatal(err)
	}
	if summary == nil {
		t.Fatal("expected summary, got nil")
	}
	if summary.CardCount != 2 {
		t.Fatalf("expected card_count=2, got %d", summary.CardCount)
	}
	if summary.TotalValue != 5.00 {
		t.Fatalf("expected total_value=5.00, got %f", summary.TotalValue)
	}
	if summary.MinValue != 2.00 {
		t.Fatalf("expected min_value=2.00, got %f", summary.MinValue)
	}
	if summary.MaxValue != 3.00 {
		t.Fatalf("expected max_value=3.00, got %f", summary.MaxValue)
	}
}

func TestSetFinancialSummarySingleCard(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	if err := sdk.conn.RegisterTableFromData(ctx, "prices_today", samplePrices); err != nil {
		t.Fatal(err)
	}

	summary, err := sdk.Sets().GetFinancialSummary(ctx, "MH2")
	if err != nil {
		t.Fatal(err)
	}
	if summary == nil {
		t.Fatal("expected summary, got nil")
	}
	if summary.CardCount != 1 {
		t.Fatalf("expected card_count=1, got %d", summary.CardCount)
	}
	if summary.TotalValue != 5.00 {
		t.Fatalf("expected total_value=5.00, got %f", summary.TotalValue)
	}
}

func TestSetFinancialSummaryNoPrices(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	summary, err := sdk.Sets().GetFinancialSummary(ctx, "A25")
	if err != nil {
		t.Fatal(err)
	}
	if summary != nil {
		t.Fatalf("expected nil (no prices loaded), got %v", summary)
	}
}

func TestSetFinancialSummaryNoDataForSet(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	if err := sdk.conn.RegisterTableFromData(ctx, "prices_today", samplePrices); err != nil {
		t.Fatal(err)
	}

	summary, err := sdk.Sets().GetFinancialSummary(ctx, "XXXXX")
	if err != nil {
		t.Fatal(err)
	}
	if summary != nil {
		t.Fatalf("expected nil, got %v", summary)
	}
}
