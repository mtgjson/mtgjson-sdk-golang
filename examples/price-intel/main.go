// price-intel is a CLI tool for MTG card price intelligence and cross-platform
// identifier lookups. It demonstrates the MTGJSON Go SDK's price analytics,
// identifier cross-referencing, set financial summaries, and sealed product
// querying capabilities.
//
// Usage:
//
//	price-intel <command> [arguments]
//
// Commands:
//
//	identify <id-type> <id>           Resolve an external platform ID to MTGJSON cards
//	price <card-name>                 Show price trend and cheapest/priciest printings
//	history <uuid> [--from] [--to]    Show price history for a card
//	set-value <set-code>              Show financial summary for a set
//	sealed [--set <code>]             Browse sealed products
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	mtgjson "github.com/mtgjson/mtgjson-sdk-go"
	"github.com/mtgjson/mtgjson-sdk-go/models"
	"github.com/mtgjson/mtgjson-sdk-go/queries"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	ctx := context.Background()

	sdk, err := mtgjson.New(mtgjson.WithProgress(func(name string, current, total int64) {
		if total > 0 {
			pct := float64(current) / float64(total) * 100
			fmt.Fprintf(os.Stderr, "\r  downloading %s... %.0f%%", name, pct)
			if current == total {
				fmt.Fprintln(os.Stderr)
			}
		}
	}))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to initialize SDK: %v\n", err)
		os.Exit(1)
	}
	defer sdk.Close()

	switch cmd {
	case "identify":
		err = runIdentify(ctx, sdk, args)
	case "price":
		err = runPrice(ctx, sdk, args)
	case "history":
		err = runHistory(ctx, sdk, args)
	case "set-value":
		err = runSetValue(ctx, sdk, args)
	case "sealed":
		err = runSealed(ctx, sdk, args)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command %q\n\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `price-intel - MTG price intelligence & identifier lookup tool

Usage:
  price-intel <command> [arguments]

Commands:
  identify <id-type> <id>           Resolve an external platform ID to MTGJSON cards
                                    id-types: scryfall, tcgplayer, mtgo, arena,
                                              multiverse, mcm, cardkingdom
  price <card-name>                 Show price trend and cheapest/priciest printings
  history <uuid> [--from] [--to]    Show price history for a card UUID
  set-value <set-code>              Show financial summary for a set
  sealed [--set <code>]             Browse sealed products

Examples:
  price-intel identify scryfall "f7a99750-1a12-4bb8-b279-5c37a2d9dc2e"
  price-intel price "Lightning Bolt"
  price-intel history "abc-123-uuid" --from 2026-01-01 --to 2026-03-01
  price-intel set-value MH3
  price-intel sealed --set MH3`)
}

// --- identify ---

func runIdentify(ctx context.Context, sdk *mtgjson.SDK, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: price-intel identify <id-type> <id>\n  id-types: scryfall, tcgplayer, mtgo, arena, multiverse, mcm, cardkingdom")
	}

	idType := strings.ToLower(args[0])
	id := args[1]

	iq := sdk.Identifiers()

	type findFunc func(context.Context, string) ([]models.CardSet, error)

	var fn findFunc
	switch idType {
	case "scryfall":
		fn = iq.FindByScryfallID
	case "tcgplayer":
		fn = iq.FindByTCGPlayerID
	case "mtgo":
		fn = iq.FindByMTGOID
	case "arena":
		fn = iq.FindByMTGArenaID
	case "multiverse":
		fn = iq.FindByMultiverseID
	case "mcm":
		fn = iq.FindByMCMID
	case "cardkingdom":
		fn = iq.FindByCardKingdomID
	default:
		return fmt.Errorf("unknown id-type %q; supported: scryfall, tcgplayer, mtgo, arena, multiverse, mcm, cardkingdom", idType)
	}

	cards, err := fn(ctx, id)
	if err != nil {
		return fmt.Errorf("identifier lookup failed: %w", err)
	}

	if len(cards) == 0 {
		fmt.Printf("No cards found for %s ID %q\n", idType, id)
		return nil
	}

	fmt.Printf("Found %d card(s) for %s ID %q:\n\n", len(cards), idType, id)
	fmt.Printf("  %-40s %-8s %-8s %s\n", "NAME", "SET", "NUMBER", "UUID")
	fmt.Printf("  %-40s %-8s %-8s %s\n", strings.Repeat("-", 40), "---", "------", strings.Repeat("-", 36))
	for _, c := range cards {
		fmt.Printf("  %-40s %-8s %-8s %s\n", truncate(c.Name, 40), c.SetCode, c.Number, c.UUID)
	}

	return nil
}

// --- price ---

func runPrice(ctx context.Context, sdk *mtgjson.SDK, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: price-intel price <card-name>")
	}

	name := strings.Join(args, " ")
	pq := sdk.Prices()

	// Look up the card by name to get a UUID for trend data
	cards, err := sdk.Cards().GetByName(ctx, name)
	if err != nil {
		return fmt.Errorf("card lookup failed: %w", err)
	}
	if len(cards) == 0 {
		return fmt.Errorf("no card found with name %q", name)
	}

	uuid := cards[0].UUID
	fmt.Printf("Price Intelligence: %s\n", name)
	fmt.Printf("  (using printing %s [%s #%s] for trend data)\n\n", uuid, cards[0].SetCode, cards[0].Number)

	// Price trend
	trend, err := pq.PriceTrend(ctx, uuid)
	if err != nil {
		fmt.Printf("  Price Trend: unavailable (%v)\n\n", err)
	} else if trend != nil {
		fmt.Printf("  Price Trend (TCGPlayer retail, normal):\n")
		fmt.Printf("    Min: $%.2f  Max: $%.2f  Avg: $%.2f\n", trend.MinPrice, trend.MaxPrice, trend.AvgPrice)
		fmt.Printf("    Period: %s to %s (%d data points)\n\n", trend.FirstDate, trend.LastDate, trend.DataPoints)
	}

	// Cheapest printing
	cheapest, err := pq.CheapestPrinting(ctx, name)
	if err != nil {
		fmt.Printf("  Cheapest Printing: unavailable (%v)\n\n", err)
	} else if cheapest != nil {
		fmt.Printf("  Cheapest Printing:\n")
		printMapFields(cheapest, "    ")
		fmt.Println()
	}

	// Most expensive printings (top 5)
	expensive, err := pq.MostExpensivePrintings(ctx, queries.WithListLimit(5))
	if err != nil {
		fmt.Printf("  Most Expensive Printings: unavailable (%v)\n", err)
	} else if len(expensive) > 0 {
		fmt.Printf("  Top 5 Most Expensive Printings (all cards):\n")
		fmt.Printf("    %-35s %-8s %s\n", "NAME", "SET", "PRICE")
		fmt.Printf("    %-35s %-8s %s\n", strings.Repeat("-", 35), "---", "-----")
		for _, e := range expensive {
			fmt.Printf("    %-35s %-8s $%.2f\n", truncate(e.Name, 35), e.SetCode, e.MaxPrice)
		}
	}

	return nil
}

// --- history ---

func runHistory(ctx context.Context, sdk *mtgjson.SDK, args []string) error {
	fs := flag.NewFlagSet("history", flag.ExitOnError)
	from := fs.String("from", "", "Start date (YYYY-MM-DD)")
	to := fs.String("to", "", "End date (YYYY-MM-DD)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: price-intel history <uuid> [--from YYYY-MM-DD] [--to YYYY-MM-DD]")
	}

	uuid := fs.Arg(0)

	var opts []queries.PriceHistoryOption
	if *from != "" {
		opts = append(opts, queries.WithDateFrom(*from))
	}
	if *to != "" {
		opts = append(opts, queries.WithDateTo(*to))
	}

	rows, err := sdk.Prices().History(ctx, uuid, opts...)
	if err != nil {
		return fmt.Errorf("price history lookup failed: %w", err)
	}

	if len(rows) == 0 {
		fmt.Printf("No price history found for UUID %q\n", uuid)
		return nil
	}

	fmt.Printf("Price History for %s", uuid)
	if *from != "" || *to != "" {
		fmt.Printf(" (")
		if *from != "" {
			fmt.Printf("from %s", *from)
		}
		if *from != "" && *to != "" {
			fmt.Printf(" ")
		}
		if *to != "" {
			fmt.Printf("to %s", *to)
		}
		fmt.Printf(")")
	}
	fmt.Printf(":\n\n")

	fmt.Printf("  %-12s %-12s %-8s %-10s %s\n", "DATE", "PROVIDER", "FINISH", "TYPE", "PRICE")
	fmt.Printf("  %-12s %-12s %-8s %-10s %s\n", "----", "--------", "------", "----", "-----")
	for _, row := range rows {
		date := mapStr(row, "date")
		provider := mapStr(row, "provider")
		finish := mapStr(row, "finish")
		priceType := mapStr(row, "price_type")
		price := mapFloat(row, "price")
		fmt.Printf("  %-12s %-12s %-8s %-10s $%.2f\n", date, provider, finish, priceType, price)
	}

	fmt.Printf("\n  Total records: %d\n", len(rows))
	return nil
}

// --- set-value ---

func runSetValue(ctx context.Context, sdk *mtgjson.SDK, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: price-intel set-value <set-code>")
	}

	setCode := strings.ToUpper(args[0])

	setInfo, err := sdk.Sets().Get(ctx, setCode)
	if err != nil {
		return fmt.Errorf("set lookup failed: %w", err)
	}

	fmt.Printf("Financial Summary: %s (%s)\n", setInfo.Name, setInfo.Code)
	fmt.Printf("  Type: %s  Released: %s  Cards: %d\n\n", setInfo.Type, setInfo.ReleaseDate, setInfo.TotalSetSize)

	summary, err := sdk.Sets().GetFinancialSummary(ctx, setCode)
	if err != nil {
		return fmt.Errorf("financial summary failed: %w", err)
	}

	if summary == nil {
		fmt.Println("  No price data available for this set.")
		return nil
	}

	fmt.Printf("  TCGPlayer Retail (Normal):\n")
	fmt.Printf("    Total Set Value:   $%.2f\n", summary.TotalValue)
	fmt.Printf("    Avg Card Value:    $%.2f\n", summary.AvgValue)
	fmt.Printf("    Min Card Price:    $%.2f\n", summary.MinValue)
	fmt.Printf("    Max Card Price:    $%.2f\n", summary.MaxValue)
	fmt.Printf("    Cards with Prices: %d\n", summary.CardCount)
	fmt.Printf("    Price Date:        %s\n", summary.Date)

	return nil
}

// --- sealed ---

func runSealed(ctx context.Context, sdk *mtgjson.SDK, args []string) error {
	fs := flag.NewFlagSet("sealed", flag.ExitOnError)
	setCode := fs.String("set", "", "Filter by set code")
	if err := fs.Parse(args); err != nil {
		return err
	}

	// If a positional arg is given, treat it as a UUID for detail view
	if fs.NArg() > 0 {
		uuid := fs.Arg(0)
		return runSealedDetail(ctx, sdk, uuid)
	}

	params := queries.ListSealedParams{
		SetCode: strings.ToUpper(*setCode),
		Limit:   50,
	}

	products, err := sdk.Sealed().List(ctx, params)
	if err != nil {
		return fmt.Errorf("sealed product listing failed: %w", err)
	}

	if len(products) == 0 {
		msg := "No sealed products found"
		if *setCode != "" {
			msg += fmt.Sprintf(" for set %q", *setCode)
		}
		fmt.Println(msg)
		return nil
	}

	header := "Sealed Products"
	if *setCode != "" {
		header += fmt.Sprintf(" (set: %s)", strings.ToUpper(*setCode))
	}
	fmt.Printf("%s:\n\n", header)

	fmt.Printf("  %-45s %-12s %-10s %s\n", "NAME", "CATEGORY", "SET", "UUID")
	fmt.Printf("  %-45s %-12s %-10s %s\n", strings.Repeat("-", 45), "--------", "---", strings.Repeat("-", 36))
	for _, p := range products {
		name := mapStr(p, "name")
		category := mapStr(p, "category")
		set := mapStr(p, "setCode")
		uuid := mapStr(p, "uuid")
		fmt.Printf("  %-45s %-12s %-10s %s\n", truncate(name, 45), category, set, uuid)
	}

	fmt.Printf("\n  Showing %d product(s). Use 'price-intel sealed <uuid>' for details.\n", len(products))
	return nil
}

func runSealedDetail(ctx context.Context, sdk *mtgjson.SDK, uuid string) error {
	product, err := sdk.Sealed().Get(ctx, uuid)
	if err != nil {
		return fmt.Errorf("sealed product lookup failed: %w", err)
	}

	if product == nil {
		return fmt.Errorf("no sealed product found with UUID %q", uuid)
	}

	fmt.Printf("Sealed Product Detail:\n\n")
	printMapFields(product, "  ")
	return nil
}

// --- helpers ---

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func mapStr(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func mapFloat(m map[string]any, key string) float64 {
	if v, ok := m[key]; ok {
		switch f := v.(type) {
		case float64:
			return f
		case float32:
			return float64(f)
		case int64:
			return float64(f)
		case int:
			return float64(f)
		}
	}
	return 0
}

func printMapFields(m map[string]any, prefix string) {
	for k, v := range m {
		fmt.Printf("%s%s: %v\n", prefix, k, v)
	}
}
