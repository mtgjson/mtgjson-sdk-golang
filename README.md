# mtgjson-sdk-go

A high-performance, DuckDB-backed Go query client for [MTGJSON](https://mtgjson.com).

Unlike traditional SDKs that rely on rate-limited REST APIs, `mtgjson-sdk-go` implements a local data warehouse architecture. It synchronizes optimized Parquet data from the MTGJSON CDN to your local machine, utilizing DuckDB to execute complex analytics, fuzzy searches, and booster simulations with sub-millisecond latency.

## Key Features

*   **Vectorized Execution**: Powered by DuckDB for high-speed OLAP queries on the full MTG dataset.
*   **Offline-First**: Data is cached locally, allowing for full functionality without an active internet connection.
*   **Fuzzy Search**: Built-in Jaro-Winkler similarity matching to handle typos and approximate name lookups.
*   **Context Support**: All methods accept `context.Context` for cancellation and timeouts.
*   **Functional Options**: Idiomatic Go configuration with composable `With*` option functions.
*   **Booster Simulation**: Accurate pack opening logic using official MTGJSON weights and sheet configurations.

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

## Architecture

By using DuckDB, the SDK leverages columnar storage and vectorized execution, making it significantly faster than SQLite or standard JSON parsing for MTG's relational dataset.

1.  **Synchronization**: On first use, the SDK lazily downloads Parquet and JSON files from the MTGJSON CDN to a platform-specific cache directory (`~/.cache/mtgjson-sdk` on Linux, `~/Library/Caches/mtgjson-sdk` on macOS, `AppData/Local/mtgjson-sdk` on Windows).
2.  **Virtual Schema**: DuckDB views are registered on-demand. Accessing `sdk.Cards()` registers the card view; accessing `sdk.Prices()` registers price data. You only pay the memory cost for the data you query.
3.  **Dynamic Adaptation**: The SDK introspects Parquet metadata to automatically handle schema changes, plural-column array conversion, and format legality unpivoting.
4.  **Materialization**: Queries return typed Go structs for individual record ergonomics, or `map[string]any` for flexible consumption.

## Use Cases

### Price Analytics

```go
ctx := context.Background()
sdk, _ := mtgjson.New()
defer sdk.Close()

// Find the cheapest printing of a card by name
cheapest, _ := sdk.Prices().CheapestPrinting(ctx, "Ragavan, Nimble Pilferer")

// Aggregate statistics (min, max, avg) for a specific card
uuid := cheapest["uuid"].(string)
trend, _ := sdk.Prices().PriceTrend(ctx, uuid,
	queries.WithPriceProvider("tcgplayer"),
	queries.WithPriceFinish("normal"),
)
fmt.Printf("Range: $%.2f - $%.2f\n", trend.MinPrice, trend.MaxPrice)
fmt.Printf("Average: $%.2f over %d data points\n", trend.AvgPrice, trend.DataPoints)

// Historical price lookup with date filtering
history, _ := sdk.Prices().History(ctx, uuid,
	queries.WithHistoryProvider("tcgplayer"),
	queries.WithDateFrom("2024-01-01"),
	queries.WithDateTo("2024-12-31"),
)

// Top 10 most expensive printings across the entire dataset
priciest, _ := sdk.Prices().MostExpensivePrintings(ctx,
	queries.WithListLimit(10),
)
```

### Advanced Card Search

The `Search()` method supports ~20 composable filters that can be combined freely:

```go
ctx := context.Background()
sdk, _ := mtgjson.New()
defer sdk.Close()

// Complex filters: Modern-legal red creatures with CMC <= 2
manaValueLte := 2.0
aggro, _ := sdk.Cards().Search(ctx, queries.SearchCardsParams{
	Colors:       []string{"R"},
	Types:        "Creature",
	ManaValueLTE: &manaValueLte,
	LegalIn:      "modern",
	Limit:        50,
})

// Typo-tolerant fuzzy search (Jaro-Winkler similarity)
results, _ := sdk.Cards().Search(ctx, queries.SearchCardsParams{
	FuzzyName: "Ligtning Bolt",  // still finds it!
})

// Rules text search using regular expressions
burn, _ := sdk.Cards().Search(ctx, queries.SearchCardsParams{
	TextRegex: `deals? \d+ damage to any target`,
})

// Search by keyword ability across formats
flyers, _ := sdk.Cards().Search(ctx, queries.SearchCardsParams{
	Keyword: "Flying",
	Colors:  []string{"W", "U"},
	LegalIn: "standard",
})

// Find cards by foreign-language name
blitz, _ := sdk.Cards().Search(ctx, queries.SearchCardsParams{
	LocalizedName: "Blitzschlag",  // German for Lightning Bolt
})
```

<details>
<summary>All <code>SearchCardsParams</code> fields</summary>

| Field | Type | Description |
|---|---|---|
| `Name` | `string` | Name pattern (`%` = wildcard) |
| `FuzzyName` | `string` | Typo-tolerant Jaro-Winkler match |
| `LocalizedName` | `string` | Foreign-language name search |
| `Colors` | `[]string` | Cards containing these colors |
| `ColorIdentity` | `[]string` | Color identity filter |
| `LegalIn` | `string` | Format legality |
| `Rarity` | `string` | Rarity filter |
| `ManaValue` | `*float64` | Exact mana value |
| `ManaValueLTE` | `*float64` | Mana value upper bound |
| `ManaValueGTE` | `*float64` | Mana value lower bound |
| `Text` | `string` | Rules text substring |
| `TextRegex` | `string` | Rules text regex |
| `Types` | `string` | Type line search |
| `Artist` | `string` | Artist name |
| `Keyword` | `string` | Keyword ability |
| `IsPromo` | `*bool` | Promo status |
| `Availability` | `string` | `"paper"` or `"mtgo"` |
| `Language` | `string` | Language filter |
| `Layout` | `string` | Card layout |
| `SetCode` | `string` | Set code |
| `SetType` | `string` | Set type (joins sets table) |
| `Power` | `string` | Power filter |
| `Toughness` | `string` | Toughness filter |
| `Limit` / `Offset` | `int` | Pagination |

</details>

### Collection & Cross-Reference

```go
ctx := context.Background()
sdk, _ := mtgjson.New()
defer sdk.Close()

// Cross-reference by any external ID system
cards, _ := sdk.Identifiers().FindByScryfallID(ctx, "f7a21fe4-...")
cards, _ = sdk.Identifiers().FindByTCGPlayerID(ctx, "12345")
cards, _ = sdk.Identifiers().FindByMTGOID(ctx, "67890")

// Get all external identifiers for a card
allIDs, _ := sdk.Identifiers().GetIdentifiers(ctx, "card-uuid-here")
// -> Scryfall, TCGPlayer, MTGO, Arena, Cardmarket, Card Kingdom, Cardsphere, ...

// TCGPlayer SKU variants (foil, etched, etc.)
skus, _ := sdk.Skus().Get(ctx, "card-uuid-here")

// Export to a standalone DuckDB file for offline analysis
sdk.ExportDB(ctx, "my_collection.duckdb")
// Now query with: duckdb my_collection.duckdb "SELECT * FROM cards LIMIT 5"
```

### Booster Simulation

```go
ctx := context.Background()
sdk, _ := mtgjson.New()
defer sdk.Close()

// See available booster types for a set
types, _ := sdk.Booster().AvailableTypes(ctx, "MH3")  // ["draft", "collector", ...]

// Open a single draft pack using official set weights
pack, _ := sdk.Booster().OpenPack(ctx, "MH3", "draft")
for _, card := range pack {
	fmt.Printf("  %s (%s)\n", card.Name, card.Rarity)
}

// Simulate opening a full box (36 packs)
box, _ := sdk.Booster().OpenBox(ctx, "MH3", "draft", 36)
totalCards := 0
for _, p := range box {
	totalCards += len(p)
}
fmt.Printf("Opened %d packs, %d total cards\n", len(box), totalCards)
```

## API Reference

### Core Data

```go
// Cards
sdk.Cards().GetByUUID(ctx, "uuid")               // single card lookup
sdk.Cards().GetByUUIDs(ctx, []string{"uuid1"})   // batch lookup
sdk.Cards().GetByName(ctx, "Lightning Bolt")     // all printings of a name
sdk.Cards().Search(ctx, SearchCardsParams{...})  // composable filters (see above)
sdk.Cards().GetPrintings(ctx, "Lightning Bolt")  // all printings across sets
sdk.Cards().GetAtomic(ctx, "Lightning Bolt")     // oracle data (no printing info)
sdk.Cards().FindByScryfallID(ctx, "...")         // cross-reference shortcut
sdk.Cards().Random(ctx, 5)                       // random cards
sdk.Cards().Count(ctx)                           // total (or filtered with kwargs)

// Tokens
sdk.Tokens().GetByUUID(ctx, "uuid")
sdk.Tokens().GetByName(ctx, "Soldier")
sdk.Tokens().Search(ctx, SearchTokensParams{Name: "%Token", SetCode: "MH3"})
sdk.Tokens().ForSet(ctx, "MH3")
sdk.Tokens().Count(ctx)

// Sets
sdk.Sets().Get(ctx, "MH3")
sdk.Sets().List(ctx, ListSetsParams{SetType: "expansion"})
sdk.Sets().Search(ctx, SearchSetsParams{Name: "Horizons"})
sdk.Sets().GetFinancialSummary(ctx, "MH3", WithProvider("tcgplayer"))
sdk.Sets().Count(ctx)
```

### Playability

```go
// Legalities
sdk.Legalities().FormatsForCard(ctx, "uuid")     // -> (map[string]string, error)
sdk.Legalities().LegalIn(ctx, "modern")          // all modern-legal cards
sdk.Legalities().IsLegal(ctx, "uuid", "modern")  // -> (bool, error)
sdk.Legalities().BannedIn(ctx, "modern")         // also: RestrictedIn, SuspendedIn

// Decks & Sealed Products
sdk.Decks().List(ctx, ListDecksParams{SetCode: "MH3"})
sdk.Decks().Search(ctx, SearchDecksParams{Name: "Eldrazi"})
sdk.Decks().Count(ctx)
sdk.Sealed().List(ctx, ListSealedParams{SetCode: "MH3"})
sdk.Sealed().Get(ctx, "uuid")
```

### Market & Identifiers

```go
// Prices
sdk.Prices().Get(ctx, "uuid")                    // full nested price data
sdk.Prices().Today(ctx, "uuid", WithPriceProvider("tcgplayer"))
sdk.Prices().History(ctx, "uuid", WithHistoryProvider("tcgplayer"))
sdk.Prices().PriceTrend(ctx, "uuid")             // min/max/avg statistics
sdk.Prices().CheapestPrinting(ctx, "Lightning Bolt")
sdk.Prices().MostExpensivePrintings(ctx, WithListLimit(10))

// Identifiers (supports all major external ID systems)
sdk.Identifiers().FindByScryfallID(ctx, "...")
sdk.Identifiers().FindByTCGPlayerID(ctx, "...")
sdk.Identifiers().FindByMTGOID(ctx, "...")
sdk.Identifiers().FindByMTGArenaID(ctx, "...")
sdk.Identifiers().FindByMultiverseID(ctx, "...")
sdk.Identifiers().FindByMCMID(ctx, "...")
sdk.Identifiers().FindByCardKingdomID(ctx, "...")
sdk.Identifiers().FindBy(ctx, "scryfallId", "...")  // generic lookup
sdk.Identifiers().GetIdentifiers(ctx, "uuid")       // all IDs for a card

// SKUs
sdk.Skus().Get(ctx, "uuid")
sdk.Skus().FindBySkuID(ctx, 123456)
sdk.Skus().FindByProductID(ctx, 789)
```

### Booster & Enums

```go
sdk.Booster().AvailableTypes(ctx, "MH3")
sdk.Booster().OpenPack(ctx, "MH3", "draft")
sdk.Booster().OpenBox(ctx, "MH3", "draft", 36)
sdk.Booster().SheetContents(ctx, "MH3", "draft", "common")

sdk.Enums().Keywords(ctx)
sdk.Enums().CardTypes(ctx)
sdk.Enums().EnumValues(ctx)
```

### System

```go
sdk.Meta(ctx)                                    // version and build date
sdk.Views()                                      // registered view names
sdk.Refresh(ctx)                                 // check CDN for new data -> (bool, error)
sdk.ExportDB(ctx, "output.duckdb")               // export to persistent DuckDB file
sdk.SQL(ctx, query, params...)                   // raw parameterized SQL
sdk.EnsureViews(ctx, "cards", "sets")            // pre-download specific tables
sdk.Connection()                                 // *db.Connection for advanced usage
sdk.Close()                                      // release resources
```

## Performance and Memory

When querying large datasets (thousands of cards), use `map[string]any` results from raw SQL rather than deserializing into structs. This avoids the reflection overhead of `json.Unmarshal` for bulk analysis.

```go
// Use raw SQL for bulk analysis
rows, _ := sdk.SQL(ctx, `
    SELECT setCode, COUNT(*) as card_count, AVG(manaValue) as avg_cmc
    FROM cards
    GROUP BY setCode
    ORDER BY card_count DESC
    LIMIT 10
`)
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

### Raw SQL

All user input goes through DuckDB parameter binding (`$1`, `$2`, ...):

```go
ctx := context.Background()

// Ensure views are registered before querying
sdk.EnsureViews(ctx, "cards")

// Parameterized queries
rows, _ := sdk.SQL(ctx,
    "SELECT name, setCode, rarity FROM cards WHERE manaValue <= $1 AND rarity = $2",
    2, "mythic",
)
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

## Examples

### Price Intelligence CLI (`examples/price-intel`)

A CLI tool for looking up card prices, tracking price trends, evaluating set value, and resolving cards across platforms like Scryfall, TCGPlayer, and MTGO.

**Subcommands:**

- `identify <id-type> <id>` -- Resolve external platform IDs (Scryfall, TCGPlayer, MTGO, Arena, Multiverse, Cardmarket, Card Kingdom) to MTGJSON cards
- `price <card-name>` -- Price trend statistics, cheapest printing, and most expensive printings
- `history <uuid> [--from] [--to]` -- Price history with date range filtering
- `set-value <set-code>` -- Financial summary for a set (total value, avg/min/max card price)
- `sealed [--set <code>]` -- Browse sealed products; pass a UUID for detail view

**SDK features showcased:** `Identifiers.FindBy*` (7 platform lookups), `Prices.PriceTrend`, `Prices.History`, `Prices.CheapestPrinting`, `Prices.MostExpensivePrintings`, `Sets.GetFinancialSummary`, `Sealed.List`, `Sealed.Get`, `Cards.GetByName`, `Sets.Get`, and `WithProgress` for download feedback.

**Setup:**

```bash
cd examples/price-intel
go build -o price-intel .
./price-intel set-value MH3
./price-intel price "Lightning Bolt"
./price-intel identify tcgplayer "126455"
./price-intel history <uuid> --from 2026-01-01 --to 2026-03-01
```

> First run downloads parquet data from the MTGJSON CDN (~30s cold start). Subsequent runs use the local cache.

## Development

```bash
git clone https://github.com/mtgjson/mtgjson-sdk-go.git
cd mtgjson-sdk-go
go mod download
go test ./...
go vet ./...
```

## Authors

- **Zachary Halpern** — [zach@mtgjson.com](mailto:zach@mtgjson.com)
- **Robert Pratt**

## License

MIT — see [LICENSE](LICENSE) for details.
