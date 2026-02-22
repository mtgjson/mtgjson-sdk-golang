package mtgjson

import (
	"os"
	"path/filepath"
	"runtime"
)

// CDNBase is the base URL for the MTGJSON v5 API / CDN.
const CDNBase = "https://mtgjson.com/api/v5"

// MetaURL is the URL for the MTGJSON version metadata endpoint.
const MetaURL = CDNBase + "/Meta.json"

// ParquetFiles maps logical view names to CDN parquet file paths.
var ParquetFiles = map[string]string{
	// Flat normalized tables
	"cards":             "parquet/cards.parquet",
	"tokens":            "parquet/tokens.parquet",
	"sets":              "parquet/sets.parquet",
	"card_identifiers":  "parquet/cardIdentifiers.parquet",
	"card_legalities":   "parquet/cardLegalities.parquet",
	"card_foreign_data": "parquet/cardForeignData.parquet",
	"card_rulings":      "parquet/cardRulings.parquet",
	"card_purchase_urls": "parquet/cardPurchaseUrls.parquet",
	"set_translations":  "parquet/setTranslations.parquet",
	"token_identifiers": "parquet/tokenIdentifiers.parquet",
	// Booster tables
	"set_booster_content_weights": "parquet/setBoosterContentWeights.parquet",
	"set_booster_contents":        "parquet/setBoosterContents.parquet",
	"set_booster_sheet_cards":     "parquet/setBoosterSheetCards.parquet",
	"set_booster_sheets":          "parquet/setBoosterSheets.parquet",
	// Full nested
	"all_printings": "parquet/AllPrintings.parquet",
}

// JSONFiles maps logical data names to CDN JSON file paths.
var JSONFiles = map[string]string{
	"all_prices_today": "AllPricesToday.json.gz",
	"tcgplayer_skus":   "TcgplayerSkus.json.gz",
	"keywords":         "Keywords.json",
	"card_types":       "CardTypes.json",
	"deck_list":        "DeckList.json",
	"enum_values":      "EnumValues.json",
	"meta":             "Meta.json",
}

// defaultCacheDir returns the platform-appropriate cache directory.
func defaultCacheDir() string {
	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
		return filepath.Join(base, "mtgjson-sdk")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Caches", "mtgjson-sdk")
	default:
		base := os.Getenv("XDG_CACHE_HOME")
		if base == "" {
			base = filepath.Join(os.Getenv("HOME"), ".cache")
		}
		return filepath.Join(base, "mtgjson-sdk")
	}
}
