package queries

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mtgjson/mtgjson-sdk-go/db"
)

// Extended sample price data for price query tests
var samplePricesExtended = []map[string]any{
	{
		"uuid": "card-uuid-001", "source": "paper", "provider": "tcgplayer",
		"currency": "USD", "price_type": "retail", "finish": "normal",
		"date": "2024-01-01", "price": 1.50,
	},
	{
		"uuid": "card-uuid-001", "source": "paper", "provider": "tcgplayer",
		"currency": "USD", "price_type": "retail", "finish": "normal",
		"date": "2024-01-02", "price": 1.75,
	},
	{
		"uuid": "card-uuid-001", "source": "paper", "provider": "tcgplayer",
		"currency": "USD", "price_type": "retail", "finish": "normal",
		"date": "2024-01-03", "price": 2.00,
	},
	{
		"uuid": "card-uuid-001", "source": "paper", "provider": "tcgplayer",
		"currency": "USD", "price_type": "retail", "finish": "foil",
		"date": "2024-01-01", "price": 3.50,
	},
	{
		"uuid": "card-uuid-001", "source": "paper", "provider": "tcgplayer",
		"currency": "USD", "price_type": "retail", "finish": "foil",
		"date": "2024-01-03", "price": 4.00,
	},
	{
		"uuid": "card-uuid-001", "source": "paper", "provider": "tcgplayer",
		"currency": "USD", "price_type": "buylist", "finish": "normal",
		"date": "2024-01-03", "price": 0.80,
	},
	{
		"uuid": "card-uuid-002", "source": "paper", "provider": "tcgplayer",
		"currency": "USD", "price_type": "retail", "finish": "normal",
		"date": "2024-01-03", "price": 5.00,
	},
}

// dateStr extracts a date string from a DuckDB value (may be time.Time or string).
func dateStr(v any) string {
	switch d := v.(type) {
	case time.Time:
		return d.Format("2006-01-02")
	case string:
		return d
	default:
		return fmt.Sprint(v)
	}
}

func setupPriceQuery(t *testing.T) *PriceQuery {
	t.Helper()
	conn := setupSampleDB(t)
	ctx := context.Background()
	if err := conn.RegisterTableFromData(ctx, "all_prices_today", samplePricesExtended); err != nil {
		t.Fatal(err)
	}
	if err := conn.RegisterTableFromData(ctx, "all_prices", samplePricesExtended); err != nil {
		t.Fatal(err)
	}
	pq := &PriceQuery{conn: conn}
	return pq
}

func TestTodayReturnsLatestDate(t *testing.T) {
	pq := setupPriceQuery(t)
	ctx := context.Background()

	rows, err := pq.Today(ctx, "card-uuid-001")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if dateStr(r["date"]) != "2024-01-03" {
			t.Fatalf("expected date 2024-01-03, got %v", r["date"])
		}
	}
}

func TestTodayWithProviderFilter(t *testing.T) {
	pq := setupPriceQuery(t)
	ctx := context.Background()

	rows, err := pq.Today(ctx, "card-uuid-001", WithPriceProvider("tcgplayer"))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if r["provider"] != "tcgplayer" {
			t.Fatalf("expected provider tcgplayer, got %v", r["provider"])
		}
		if dateStr(r["date"]) != "2024-01-03" {
			t.Fatalf("expected date 2024-01-03, got %v", r["date"])
		}
	}
}

func TestTodayWithFinishFilter(t *testing.T) {
	pq := setupPriceQuery(t)
	ctx := context.Background()

	rows, err := pq.Today(ctx, "card-uuid-001", WithPriceFinish("foil"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	price := db.ToFloat64(rows[0]["price"])
	if price != 4.00 {
		t.Fatalf("expected price 4.00, got %v", price)
	}
}

func TestTodayWithPriceTypeFilter(t *testing.T) {
	pq := setupPriceQuery(t)
	ctx := context.Background()

	rows, err := pq.Today(ctx, "card-uuid-001", WithPriceType("buylist"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["price_type"] != "buylist" {
		t.Fatalf("expected price_type buylist, got %v", rows[0]["price_type"])
	}
}

func TestHistoryAllDates(t *testing.T) {
	pq := setupPriceQuery(t)
	ctx := context.Background()

	rows, err := pq.History(ctx, "card-uuid-001",
		WithHistoryFinish("normal"), WithHistoryPriceType("retail"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	expected := []string{"2024-01-01", "2024-01-02", "2024-01-03"}
	for i, r := range rows {
		if dateStr(r["date"]) != expected[i] {
			t.Fatalf("expected date %s, got %v", expected[i], r["date"])
		}
	}
}

func TestHistoryDateRange(t *testing.T) {
	pq := setupPriceQuery(t)
	ctx := context.Background()

	rows, err := pq.History(ctx, "card-uuid-001",
		WithHistoryFinish("normal"), WithHistoryPriceType("retail"),
		WithDateFrom("2024-01-02"), WithDateTo("2024-01-03"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}

func TestHistoryDateFromOnly(t *testing.T) {
	pq := setupPriceQuery(t)
	ctx := context.Background()

	rows, err := pq.History(ctx, "card-uuid-001",
		WithHistoryFinish("normal"), WithHistoryPriceType("retail"),
		WithDateFrom("2024-01-03"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if db.ToFloat64(rows[0]["price"]) != 2.00 {
		t.Fatalf("expected price 2.00, got %v", rows[0]["price"])
	}
}

func TestPriceTrend(t *testing.T) {
	pq := setupPriceQuery(t)
	ctx := context.Background()

	trend, err := pq.PriceTrend(ctx, "card-uuid-001",
		WithPriceProvider("tcgplayer"), WithPriceFinish("normal"))
	if err != nil {
		t.Fatal(err)
	}
	if trend == nil {
		t.Fatal("expected trend, got nil")
	}
	if trend.MinPrice != 1.50 {
		t.Fatalf("expected min_price 1.50, got %v", trend.MinPrice)
	}
	if trend.MaxPrice != 2.00 {
		t.Fatalf("expected max_price 2.00, got %v", trend.MaxPrice)
	}
	if trend.FirstDate != "2024-01-01" {
		t.Fatalf("expected first_date 2024-01-01, got %s", trend.FirstDate)
	}
	if trend.LastDate != "2024-01-03" {
		t.Fatalf("expected last_date 2024-01-03, got %s", trend.LastDate)
	}
	if trend.DataPoints != 3 {
		t.Fatalf("expected 3 data_points, got %d", trend.DataPoints)
	}
}

func TestPriceTrendNoData(t *testing.T) {
	pq := setupPriceQuery(t)
	ctx := context.Background()

	trend, err := pq.PriceTrend(ctx, "nonexistent-uuid")
	if err != nil {
		t.Fatal(err)
	}
	if trend != nil {
		t.Fatal("expected nil trend")
	}
}

func TestCheapestPrintings(t *testing.T) {
	pq := setupPriceQuery(t)
	ctx := context.Background()

	rows, err := pq.CheapestPrintings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) < 1 {
		t.Fatal("expected at least 1 row")
	}
	for _, r := range rows {
		if r.Name == "" {
			t.Fatal("expected name")
		}
		if r.SetCode == "" {
			t.Fatal("expected cheapest_set")
		}
		if r.UUID == "" {
			t.Fatal("expected cheapest_uuid")
		}
	}
	// Find the two cards by name
	priceMap := make(map[string]float64)
	for _, r := range rows {
		priceMap[r.Name] = r.MinPrice
	}
	if priceMap["Lightning Bolt"] >= priceMap["Counterspell"] {
		t.Fatalf("Lightning Bolt ($%.2f) should be cheaper than Counterspell ($%.2f)",
			priceMap["Lightning Bolt"], priceMap["Counterspell"])
	}
}

func TestMostExpensivePrintings(t *testing.T) {
	pq := setupPriceQuery(t)
	ctx := context.Background()

	rows, err := pq.MostExpensivePrintings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) < 1 {
		t.Fatal("expected at least 1 row")
	}
	for _, r := range rows {
		if r.Name == "" {
			t.Fatal("expected name")
		}
		if r.SetCode == "" {
			t.Fatal("expected priciest_set")
		}
	}
	// Should be ordered DESC
	if len(rows) >= 2 && rows[0].MaxPrice < rows[len(rows)-1].MaxPrice {
		t.Fatal("expected DESC order")
	}
}

func TestCheapestPrintingsNoPrices(t *testing.T) {
	conn := setupSampleDB(t)
	pq := &PriceQuery{conn: conn}
	ctx := context.Background()

	rows, err := pq.CheapestPrintings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if rows != nil {
		t.Fatalf("expected nil, got %v", rows)
	}
}

func TestGetReturnsNestedStructure(t *testing.T) {
	pq := setupPriceQuery(t)
	ctx := context.Background()

	result, err := pq.Get(ctx, "card-uuid-001")
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	// Should have paper->tcgplayer->...
	paper, ok := result["paper"].(map[string]any)
	if !ok {
		t.Fatal("expected paper key")
	}
	tcg, ok := paper["tcgplayer"].(map[string]any)
	if !ok {
		t.Fatal("expected tcgplayer key")
	}
	if tcg["currency"] != "USD" {
		t.Fatalf("expected USD currency, got %v", tcg["currency"])
	}
}
