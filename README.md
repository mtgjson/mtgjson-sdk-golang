# mtgjson-sdk-go

A DuckDB-backed Go query client for [MTGJSON](https://mtgjson.com) card data. Auto-downloads Parquet data from the MTGJSON CDN and exposes the full Magic: The Gathering dataset through a typed Go API with context support and functional options.

## Install

```bash
go get github.com/mtgjson/mtgjson-sdk-go
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"

	mtgjson "github.com/mtgjson/mtgjson-sdk-go"
	"github.com/mtgjson/mtgjson-sdk-go/queries"
)

func main() {
	ctx := context.Background()

	sdk, err := mtgjson.New()
	if err != nil {
		panic(err)
	}
	defer sdk.Close()

	// Search for cards
	bolts, _ := sdk.Cards().GetByName(ctx, "Lightning Bolt")
	fmt.Printf("Found %d printings of Lightning Bolt\n", len(bolts))

	// Get a specific set
	mh3, _ := sdk.Sets().Get(ctx, "MH3")
	if mh3 != nil {
		fmt.Printf("%s -- %d cards\n", mh3.Name, mh3.TotalSetSize)
	}

	// Check format legality
	isLegal, _ := sdk.Legalities().IsLegal(ctx, bolts[0].UUID, "modern")
	fmt.Printf("Modern legal: %v\n", isLegal)

	// Find the cheapest printing
	cheapest, _ := sdk.Prices().CheapestPrinting(ctx, "Lightning Bolt")
	if cheapest != nil {
		fmt.Printf("Cheapest: $%v (%v)\n", cheapest["price"], cheapest["setCode"])
	}

	// Raw SQL for anything else
	rows, _ := sdk.SQL(ctx, "SELECT name, manaValue FROM cards WHERE manaValue = $1 LIMIT 5", 0)
	_ = rows
}
```

## Use Cases

### Price Tracking

```go
ctx := context.Background()
sdk, _ := mtgjson.New()
defer sdk.Close()

// Find the cheapest printing of any card
cheapest, _ := sdk.Prices().CheapestPrinting(ctx, "Ragavan, Nimble Pilferer")

// Price trend over time
uuid := cheapest["uuid"].(string)
trend, _ := sdk.Prices().PriceTrend(ctx, uuid,
	queries.WithPriceProvider("tcgplayer"),
	queries.WithPriceFinish("normal"),
)
fmt.Printf("Range: $%.2f - $%.2f\n", trend.MinPrice, trend.MaxPrice)
fmt.Printf("Average: $%.2f over %d data points\n", trend.AvgPrice, trend.DataPoints)

// Full price history with date range
history, _ := sdk.Prices().History(ctx, uuid,
	queries.WithHistoryProvider("tcgplayer"),
	queries.WithDateFrom("2024-01-01"),
	queries.WithDateTo("2024-12-31"),
)

// Most expensive printings across the entire dataset
priciest, _ := sdk.Prices().MostExpensivePrintings(ctx,
	queries.WithListLimit(10),
)
```

### Deck Building Helper

```go
ctx := context.Background()
sdk, _ := mtgjson.New()
defer sdk.Close()

// Find modern-legal red creatures with CMC <= 2
manaValueLte := 2.0
aggro, _ := sdk.Cards().Search(ctx, queries.SearchCardsParams{
	Colors:      []string{"R"},
	Types:       "Creature",
	ManaValueLTE: &manaValueLte,
	LegalIn:     "modern",
	Limit:       50,
})

// Check what's banned
banned, _ := sdk.Legalities().BannedIn(ctx, "modern")
fmt.Printf("%d cards banned in Modern\n", len(banned))

// Search by keyword ability
flyers, _ := sdk.Cards().Search(ctx, queries.SearchCardsParams{
	Keyword: "Flying",
	Colors:  []string{"W", "U"},
	LegalIn: "standard",
})

// Fuzzy search -- handles typos
results, _ := sdk.Cards().Search(ctx, queries.SearchCardsParams{
	FuzzyName: "Ligtning Bolt",  // still finds it!
})

// Find cards by foreign-language name
blitz, _ := sdk.Cards().Search(ctx, queries.SearchCardsParams{
	LocalizedName: "Blitzschlag",  // German for Lightning Bolt
})
```

### Collection Management

```go
ctx := context.Background()
sdk, _ := mtgjson.New()
defer sdk.Close()

// Cross-reference by Scryfall ID
cards, _ := sdk.Identifiers().FindByScryfallID(ctx, "f7a21fe4-...")

// Look up by TCGPlayer product ID
cards, _ = sdk.Identifiers().FindByTCGPlayerID(ctx, "12345")

// Get all identifiers for a card (Scryfall, TCGPlayer, MTGO, Arena, etc.)
allIDs, _ := sdk.Identifiers().GetIdentifiers(ctx, "card-uuid-here")

// Export to a standalone DuckDB file for offline analysis
sdk.ExportDB(ctx, "my_collection.duckdb")
// Now query with: duckdb my_collection.duckdb "SELECT * FROM cards LIMIT 5"
```

### Booster Pack Simulation

```go
ctx := context.Background()
sdk, _ := mtgjson.New()
defer sdk.Close()

// See what booster types are available
types, _ := sdk.Booster().AvailableTypes(ctx, "MH3")  // ["draft", "collector", ...]

// Open a single draft pack
pack, _ := sdk.Booster().OpenPack(ctx, "MH3", "draft")
for _, card := range pack {
	fmt.Printf("  %s (%s)\n", card.Name, card.Rarity)
}

// Open an entire box
box, _ := sdk.Booster().OpenBox(ctx, "MH3", "draft", 36)
totalCards := 0
for _, p := range box {
	totalCards += len(p)
}
fmt.Printf("Opened %d packs, %d total cards\n", len(box), totalCards)
```

## API Reference

### Cards

```go
sdk.Cards().GetByUUID(ctx, "uuid")                    // -> (*CardSet, error)
sdk.Cards().GetByUUIDs(ctx, []string{"uuid1", "uuid2"})  // -> ([]CardSet, error)
sdk.Cards().GetByName(ctx, "Lightning Bolt")           // -> ([]CardSet, error)
sdk.Cards().GetByName(ctx, "Lightning Bolt", "A25")    // with set filter
sdk.Cards().Search(ctx, SearchCardsParams{
    Name:          "Lightning%",       // name pattern (% = wildcard)
    FuzzyName:     "Ligtning Bolt",    // typo-tolerant (Jaro-Winkler)
    LocalizedName: "Blitzschlag",      // foreign-language name search
    Colors:        []string{"R"},      // cards containing these colors
    ColorIdentity: []string{"R", "U"}, // filter by color identity
    LegalIn:       "modern",           // format legality
    Rarity:        "rare",             // rarity filter
    ManaValue:     &manaVal,           // exact mana value (*float64)
    ManaValueLTE:  &manaMax,           // mana value range
    ManaValueGTE:  &manaMin,
    Text:          "damage",           // rules text search
    TextRegex:     `deals? \d+ damage`,// regex rules text search
    Types:         "Creature",         // type line search
    Artist:        "Christopher Moeller",
    Keyword:       "Flying",           // keyword ability
    IsPromo:       &isPromo,           // promo status (*bool)
    Availability:  "paper",            // paper, mtgo
    Language:      "English",          // language filter
    Layout:        "normal",           // card layout
    SetCode:       "MH3",             // filter by set
    SetType:       "expansion",        // set type (joins sets table)
    Power:         "3",                // P/T filter
    Toughness:     "3",
    Limit:         100,                // pagination
    Offset:        0,
})                                                     // -> ([]CardSet, error)
sdk.Cards().GetPrintings(ctx, "Lightning Bolt")        // all printings across sets
sdk.Cards().GetAtomic(ctx, "Lightning Bolt")           // oracle data (no printing info)
sdk.Cards().FindByScryfallID(ctx, "...")               // cross-reference
sdk.Cards().Random(ctx, 5)                             // random cards
sdk.Cards().Count(ctx)                                 // total count
sdk.Cards().Count(ctx, Filter{"setCode", "MH3"})      // filtered count
```

### Tokens

```go
sdk.Tokens().GetByUUID(ctx, "uuid")                    // -> (*CardToken, error)
sdk.Tokens().GetByName(ctx, "Soldier")                 // -> ([]CardToken, error)
sdk.Tokens().Search(ctx, SearchTokensParams{
    Name: "%Token", SetCode: "MH3", Colors: []string{"W"},
})
sdk.Tokens().ForSet(ctx, "MH3")                        // all tokens for a set
sdk.Tokens().Count(ctx)
```

### Sets

```go
sdk.Sets().Get(ctx, "MH3")                             // -> (*SetList, error)
sdk.Sets().List(ctx, ListSetsParams{SetType: "expansion"})
sdk.Sets().Search(ctx, SearchSetsParams{
    Name: "Horizons", ReleaseYear: &year,
})
sdk.Sets().GetFinancialSummary(ctx, "MH3",             // -> (*FinancialSummary, error)
    WithProvider("tcgplayer"),
    WithCurrency("USD"),
    WithFinish("normal"),
)
sdk.Sets().Count(ctx)
```

### Identifiers

```go
sdk.Identifiers().FindByScryfallID(ctx, "...")
sdk.Identifiers().FindByTCGPlayerID(ctx, "...")
sdk.Identifiers().FindByMTGOID(ctx, "...")
sdk.Identifiers().FindByMTGOFoilID(ctx, "...")
sdk.Identifiers().FindByMTGArenaID(ctx, "...")
sdk.Identifiers().FindByMultiverseID(ctx, "...")
sdk.Identifiers().FindByMCMID(ctx, "...")
sdk.Identifiers().FindByCardKingdomID(ctx, "...")
sdk.Identifiers().FindByCardKingdomFoilID(ctx, "...")
sdk.Identifiers().FindByCardKingdomEtchedID(ctx, "...")
sdk.Identifiers().FindByCardsphereID(ctx, "...")
sdk.Identifiers().FindByCardsphereFoilID(ctx, "...")
sdk.Identifiers().FindByScryfallOracleID(ctx, "...")
sdk.Identifiers().FindByScryfallIllustrationID(ctx, "...")
sdk.Identifiers().FindByTCGPlayerEtchedID(ctx, "...")
sdk.Identifiers().FindBy(ctx, "scryfallId", "...")     // generic lookup
sdk.Identifiers().GetIdentifiers(ctx, "uuid")          // all IDs for a card
```

### Legalities

```go
sdk.Legalities().FormatsForCard(ctx, "uuid")           // -> (map[string]string, error)
sdk.Legalities().LegalIn(ctx, "modern")                // all modern-legal cards
sdk.Legalities().IsLegal(ctx, "uuid", "modern")        // -> (bool, error)
sdk.Legalities().BannedIn(ctx, "modern")               // banned cards
sdk.Legalities().RestrictedIn(ctx, "vintage")           // restricted cards
sdk.Legalities().SuspendedIn(ctx, "historic")           // suspended cards
sdk.Legalities().NotLegalIn(ctx, "standard")            // not-legal cards
```

### Prices

```go
sdk.Prices().Get(ctx, "uuid")                          // full nested price data
sdk.Prices().Today(ctx, "uuid",                        // latest prices
    WithPriceProvider("tcgplayer"),
    WithPriceFinish("foil"),
)
sdk.Prices().History(ctx, "uuid",                      // historical prices
    WithHistoryProvider("tcgplayer"),
    WithDateFrom("2024-01-01"),
)
sdk.Prices().PriceTrend(ctx, "uuid")                   // min/max/avg statistics
sdk.Prices().CheapestPrinting(ctx, "Lightning Bolt")   // cheapest printing by name
sdk.Prices().CheapestPrintings(ctx,                    // N cheapest cards overall
    WithListLimit(10),
)
sdk.Prices().MostExpensivePrintings(ctx,               // most expensive cards
    WithListLimit(10),
)
```

### Decks

```go
sdk.Decks().List(ctx, ListDecksParams{SetCode: "MH3"})
sdk.Decks().Search(ctx, SearchDecksParams{Name: "Eldrazi"})
sdk.Decks().Count(ctx)
```

### Sealed Products

```go
sdk.Sealed().List(ctx, ListSealedParams{SetCode: "MH3"})
sdk.Sealed().Get(ctx, "uuid")
```

### SKUs

```go
sdk.Skus().Get(ctx, "uuid")                            // TCGPlayer SKUs for a card
sdk.Skus().FindBySkuID(ctx, 123456)
sdk.Skus().FindByProductID(ctx, 789)
```

### Booster Simulation

```go
sdk.Booster().AvailableTypes(ctx, "MH3")               // -> ([]string, error)
sdk.Booster().OpenPack(ctx, "MH3", "draft")            // -> ([]CardSet, error)
sdk.Booster().OpenBox(ctx, "MH3", "draft", 36)         // -> ([][]CardSet, error)
sdk.Booster().SheetContents(ctx, "MH3", "draft", "common")  // card weights
```

### Enums

```go
sdk.Enums().Keywords(ctx)                              // -> (map[string]any, error)
sdk.Enums().CardTypes(ctx)                             // -> (map[string]any, error)
sdk.Enums().EnumValues(ctx)                            // all enum values
```

### Metadata & Utilities

```go
sdk.Meta(ctx)                                          // -> (Meta, error)
sdk.Views()                                            // -> []string
sdk.Refresh(ctx)                                       // check for new data -> (bool, error)
sdk.SQL(ctx, "SELECT ...", params...)                   // raw parameterized SQL ($1, $2, ...)
sdk.ExportDB(ctx, "output.duckdb")                     // export to persistent DuckDB file
sdk.EnsureViews(ctx, "cards", "sets")                  // pre-download specific tables
sdk.Connection()                                       // *db.Connection for advanced usage
sdk.Close()                                            // release resources
```

## Advanced Usage

### Functional Options

```go
import "time"

sdk, err := mtgjson.New(
    mtgjson.WithCacheDir("/data/mtgjson-cache"),
    mtgjson.WithOffline(false),
    mtgjson.WithTimeout(5 * time.Minute),
    mtgjson.WithProgress(func(filename string, downloaded, total int64) {
        pct := float64(downloaded) / float64(total) * 100
        fmt.Printf("\r%s: %.1f%%", filename, pct)
    }),
)
```

### Context Support

All query methods accept `context.Context` for cancellation and timeouts:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

cards, err := sdk.Cards().Search(ctx, queries.SearchCardsParams{
    Name: "Lightning%",
})
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Println("Query timed out")
    }
}
```

### Database Export

Export all loaded data to a standalone DuckDB file that can be queried without the SDK:

```go
ctx := context.Background()
sdk, _ := mtgjson.New()
defer sdk.Close()

// Touch the query modules you want exported
sdk.Cards().Count(ctx)
sdk.Sets().Count(ctx)

// Export to file
sdk.ExportDB(ctx, "mtgjson.duckdb")

// Now use it standalone:
// $ duckdb mtgjson.duckdb "SELECT name, setCode FROM cards LIMIT 10"
```

### Web API Example

```go
package main

import (
    "encoding/json"
    "net/http"

    mtgjson "github.com/mtgjson/mtgjson-sdk-go"
)

var sdk *mtgjson.SDK

func getCard(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("name")
    cards, err := sdk.Cards().GetByName(r.Context(), name)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    json.NewEncoder(w).Encode(cards)
}

func main() {
    var err error
    sdk, err = mtgjson.New()
    if err != nil {
        panic(err)
    }
    defer sdk.Close()

    http.HandleFunc("/card", getCard)
    http.ListenAndServe(":8080", nil)
}
```

### Raw SQL

All user input goes through DuckDB parameter binding (`$1`, `$2`, ...) to prevent SQL injection:

```go
ctx := context.Background()

// Ensure views are registered before querying
sdk.EnsureViews(ctx, "cards")

// Parameterized queries
rows, _ := sdk.SQL(ctx,
    "SELECT name, setCode, rarity FROM cards WHERE manaValue <= $1 AND rarity = $2",
    2, "mythic",
)

// Complex analytics
rows, _ = sdk.SQL(ctx, `
    SELECT setCode, COUNT(*) as card_count, AVG(manaValue) as avg_cmc
    FROM cards
    GROUP BY setCode
    ORDER BY card_count DESC
    LIMIT 10
`)
```

### Auto-Refresh for Long-Running Services

```go
// In a scheduled task or health check:
stale, err := sdk.Refresh(ctx)
if err != nil {
    log.Printf("Refresh check failed: %v", err)
} else if stale {
    log.Println("New MTGJSON data detected -- cache refreshed")
}
```

## Architecture

```
MTGJSON CDN (Parquet + JSON files)
        |
        | auto-download on first access
        v
Local Cache (platform-specific directory)
        |
        | lazy view registration
        v
DuckDB In-Memory Database
        |
        | parameterized SQL queries
        v
Typed Go API (structs / map[string]any)
```

**How it works:**

1. **Auto-download**: On first use, the SDK downloads ~15 Parquet files and ~7 JSON files from the MTGJSON CDN to a platform-specific cache directory (`~/.cache/mtgjson-sdk` on Linux, `~/Library/Caches/mtgjson-sdk` on macOS, `AppData/Local/mtgjson-sdk` on Windows).

2. **Lazy loading**: DuckDB views are registered on-demand -- accessing `sdk.Cards()` triggers the cards view, `sdk.Prices()` triggers price data loading, etc. Only the data you use gets loaded into memory.

3. **Schema adaptation**: The SDK auto-detects array columns in parquet files using a hybrid heuristic (static baseline + dynamic plural detection + blocklist), so it adapts to upstream MTGJSON schema changes without code updates.

4. **Legality UNPIVOT**: Format legality columns are dynamically detected from the parquet schema and UNPIVOTed to `(uuid, format, status)` rows -- automatically scales to new formats.

5. **Price flattening**: Deeply nested JSON price data is streamed to NDJSON and bulk-loaded into DuckDB, minimizing memory overhead.

## Development

### Prerequisites

- Go 1.25+

### Setup

```bash
git clone https://github.com/the-muppet2/mtgjson-sdk-go.git
cd mtgjson-sdk-go
go mod download
```

### Running Tests

```bash
# Unit tests (120+ tests, no network required)
go test ./...

# Smoke test (downloads real data from CDN)
go test -run TestSmoke -tags smoke -v ./...
```

### Linting

```bash
go vet ./...
```

## License

MIT
