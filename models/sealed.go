package models

// SealedProduct represents a sealed product (booster box, bundle, etc.).
type SealedProduct struct {
	UUID         string                `json:"uuid"`
	Name         string                `json:"name"`
	Category     *string               `json:"category,omitempty"`
	SetCode      *string               `json:"setCode,omitempty"`
	Subtype      *string               `json:"subtype,omitempty"`
	Language     *string               `json:"language,omitempty"`
	ReleaseDate  *string               `json:"releaseDate,omitempty"`
	CardCount    *int                  `json:"cardCount,omitempty"`
	ProductSize  *int                  `json:"productSize,omitempty"`
	Contents     *SealedProductContents `json:"contents,omitempty"`
	Identifiers  Identifiers           `json:"identifiers"`
	PurchaseUrls PurchaseUrls          `json:"purchaseUrls"`
}

// SealedProductContents contains all possible content types in a sealed product.
type SealedProductContents struct {
	Card     []SealedProductCard          `json:"card,omitempty"`
	Deck     []SealedProductDeck          `json:"deck,omitempty"`
	Other    []SealedProductOther         `json:"other,omitempty"`
	Pack     []SealedProductPack          `json:"pack,omitempty"`
	Sealed   []SealedProductSealed        `json:"sealed,omitempty"`
	Variable []SealedProductVariableEntry `json:"variable,omitempty"`
}

// SealedProductCard is a card within a sealed product.
type SealedProductCard struct {
	Finishes []string `json:"finishes,omitempty"`
	Foil     *bool    `json:"foil,omitempty"`
	Name     string   `json:"name"`
	Number   string   `json:"number"`
	Set      string   `json:"set"`
	UUID     string   `json:"uuid"`
}

// SealedProductDeck is a deck reference in a sealed product.
type SealedProductDeck struct {
	Name string `json:"name"`
	Set  string `json:"set"`
}

// SealedProductOther is a non-card item in a sealed product.
type SealedProductOther struct {
	Name string `json:"name"`
}

// SealedProductPack is a pack reference in a sealed product.
type SealedProductPack struct {
	Code string `json:"code"`
	Set  string `json:"set"`
}

// SealedProductSealed is a nested sealed item in a sealed product.
type SealedProductSealed struct {
	Count int     `json:"count"`
	Name  string  `json:"name"`
	Set   string  `json:"set"`
	UUID  *string `json:"uuid,omitempty"`
}

// SealedProductVariableEntry contains variable content configurations.
type SealedProductVariableEntry struct {
	Configs []SealedProductVariableItem `json:"configs,omitempty"`
}

// SealedProductVariableItem is one possible configuration of variable content.
type SealedProductVariableItem struct {
	Card           []SealedProductCard           `json:"card,omitempty"`
	Deck           []SealedProductDeck           `json:"deck,omitempty"`
	Other          []SealedProductOther          `json:"other,omitempty"`
	Pack           []SealedProductPack           `json:"pack,omitempty"`
	Sealed         []SealedProductSealed         `json:"sealed,omitempty"`
	VariableConfig []SealedProductVariableConfig `json:"variable_config,omitempty"`
}

// SealedProductVariableConfig specifies the probability of a variable item.
type SealedProductVariableConfig struct {
	Chance *int `json:"chance,omitempty"`
	Weight *int `json:"weight,omitempty"`
}
