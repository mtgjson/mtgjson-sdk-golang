package mtgjson

// Meta contains metadata about the MTGJSON data set.
type Meta struct {
	Date    string `json:"date"`
	Version string `json:"version"`
}

// Identifiers contains all external identifier mappings for a card.
type Identifiers struct {
	CardKingdomEtchedId      *string `json:"cardKingdomEtchedId,omitempty"`
	CardKingdomFoilId        *string `json:"cardKingdomFoilId,omitempty"`
	CardKingdomId            *string `json:"cardKingdomId,omitempty"`
	CardsphereId             *string `json:"cardsphereId,omitempty"`
	CarpetFoilId             *string `json:"carpetFoilId,omitempty"`
	CarpetId                 *string `json:"carpetId,omitempty"`
	McmId                    *string `json:"mcmId,omitempty"`
	McmMetaId                *string `json:"mcmMetaId,omitempty"`
	MtgArenaId               *string `json:"mtgArenaId,omitempty"`
	MtgoFoilId               *string `json:"mtgoFoilId,omitempty"`
	MtgoId                   *string `json:"mtgoId,omitempty"`
	MultiverseId             *string `json:"multiverseId,omitempty"`
	ScryfallId               *string `json:"scryfallId,omitempty"`
	ScryfallOracleId         *string `json:"scryfallOracleId,omitempty"`
	ScryfallIllustrationId   *string `json:"scryfallIllustrationId,omitempty"`
	ScryfallCardBackId       *string `json:"scryfallCardBackId,omitempty"`
	TcgplayerProductId       *string `json:"tcgplayerProductId,omitempty"`
	TcgplayerEtchedProductId *string `json:"tcgplayerEtchedProductId,omitempty"`
	AsfId                    *string `json:"asfId,omitempty"`
	CanonicalId              *string `json:"canonicalId,omitempty"`
	EdhrecId                 *string `json:"edhrecId,omitempty"`
	GathererOracleId         *string `json:"gathererOracleId,omitempty"`
	ManaBoxFoilId            *string `json:"manaBoxFoilId,omitempty"`
	ManaBoxId                *string `json:"manaBoxId,omitempty"`
	ManaBoxEtchedId          *string `json:"manaBoxEtchedId,omitempty"`
	SpellTableId             *string `json:"spellTableId,omitempty"`
	TokenId                  *string `json:"tokenId,omitempty"`
	TcgplayerSkuId           *string `json:"tcgplayerSkuId,omitempty"`
}

// Legalities contains format legality statuses for a card.
type Legalities struct {
	Alchemy         *string `json:"alchemy,omitempty"`
	Brawl           *string `json:"brawl,omitempty"`
	Commander       *string `json:"commander,omitempty"`
	Duel            *string `json:"duel,omitempty"`
	Explorer        *string `json:"explorer,omitempty"`
	Future          *string `json:"future,omitempty"`
	Gladiator       *string `json:"gladiator,omitempty"`
	Historic        *string `json:"historic,omitempty"`
	HistoricBrawl   *string `json:"historicBrawl,omitempty"`
	Legacy          *string `json:"legacy,omitempty"`
	Modern          *string `json:"modern,omitempty"`
	Oathbreaker     *string `json:"oathbreaker,omitempty"`
	Oldschool       *string `json:"oldschool,omitempty"`
	Pauper          *string `json:"pauper,omitempty"`
	PauperCommander *string `json:"pauperCommander,omitempty"`
	Penny           *string `json:"penny,omitempty"`
	Pioneer         *string `json:"pioneer,omitempty"`
	Predh           *string `json:"predh,omitempty"`
	Premodern       *string `json:"premodern,omitempty"`
	Standard        *string `json:"standard,omitempty"`
	StandardBrawl   *string `json:"standardBrawl,omitempty"`
	Timeless        *string `json:"timeless,omitempty"`
	Vintage         *string `json:"vintage,omitempty"`
}

// LeadershipSkills indicates whether a card can be a commander in various formats.
type LeadershipSkills struct {
	Brawl       bool `json:"brawl"`
	Commander   bool `json:"commander"`
	Oathbreaker bool `json:"oathbreaker"`
}

// PurchaseUrls contains URLs for purchasing a card from various vendors.
type PurchaseUrls struct {
	CardKingdom       *string `json:"cardKingdom,omitempty"`
	CardKingdomEtched *string `json:"cardKingdomEtched,omitempty"`
	CardKingdomFoil   *string `json:"cardKingdomFoil,omitempty"`
	Cardmarket        *string `json:"cardmarket,omitempty"`
	Tcgplayer         *string `json:"tcgplayer,omitempty"`
	TcgplayerEtched   *string `json:"tcgplayerEtched,omitempty"`
}

// RelatedCards contains references to other related cards.
type RelatedCards struct {
	ReverseRelated []string `json:"reverseRelated,omitempty"`
	Spellbook      []string `json:"spellbook,omitempty"`
	Tokens         []string `json:"tokens,omitempty"`
}

// Rulings represents an official ruling for a card.
type Rulings struct {
	Date string `json:"date"`
	Text string `json:"text"`
}

// ForeignData contains localized card information in a non-English language.
type ForeignData struct {
	FaceName   *string `json:"faceName,omitempty"`
	FlavorText *string `json:"flavorText,omitempty"`
	Language   string  `json:"language"`
	Name       string  `json:"name"`
	Text       *string `json:"text,omitempty"`
	Type       *string `json:"type,omitempty"`
}

// SourceProducts contains product UUIDs grouped by finish type.
type SourceProducts struct {
	Etched  []string `json:"etched,omitempty"`
	Foil    []string `json:"foil,omitempty"`
	Nonfoil []string `json:"nonfoil,omitempty"`
}

// Translations is a map of language names to translated set names.
type Translations map[string]*string

// TcgplayerSkus represents a TCGplayer SKU entry for a card.
type TcgplayerSkus struct {
	Condition string  `json:"condition"`
	Finish    *string `json:"finish,omitempty"`
	Language  string  `json:"language"`
	Printing  string  `json:"printing"`
	ProductId int     `json:"productId"`
	SkuId     int     `json:"skuId"`
}
