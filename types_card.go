package mtgjson

// CardSet is the primary card model for most queries.
// It represents a card as it appears in a specific set printing.
type CardSet struct {
	// Identity
	UUID      string  `json:"uuid"`
	Name      string  `json:"name"`
	ASCIIName *string `json:"asciiName,omitempty"`
	FaceName  *string `json:"faceName,omitempty"`

	// Type line
	Type       string   `json:"type"`
	Types      []string `json:"types"`
	Subtypes   []string `json:"subtypes"`
	Supertypes []string `json:"supertypes"`

	// Colors
	Colors         []string `json:"colors"`
	ColorIdentity  []string `json:"colorIdentity"`
	ColorIndicator []string `json:"colorIndicator,omitempty"`
	ProducedMana   []string `json:"producedMana,omitempty"`

	// Mana
	ManaCost          *string  `json:"manaCost,omitempty"`
	ManaValue         float64  `json:"manaValue"`
	ConvertedManaCost *float64 `json:"convertedManaCost,omitempty"`
	FaceManaValue     *float64 `json:"faceManaValue,omitempty"`

	// Text
	Text   *string `json:"text,omitempty"`
	Layout string  `json:"layout"`
	Side   *string `json:"side,omitempty"`

	// Stats
	Power     *string `json:"power,omitempty"`
	Toughness *string `json:"toughness,omitempty"`
	Loyalty   *string `json:"loyalty,omitempty"`
	Defense   *string `json:"defense,omitempty"`
	Hand      *string `json:"hand,omitempty"`
	Life      *string `json:"life,omitempty"`

	// Printing
	SetCode       string  `json:"setCode"`
	Number        string  `json:"number"`
	Rarity        string  `json:"rarity"`
	Artist        *string `json:"artist,omitempty"`
	BorderColor   string  `json:"borderColor"`
	FrameVersion  string  `json:"frameVersion"`
	Watermark     *string `json:"watermark,omitempty"`
	Signature     *string `json:"signature,omitempty"`
	SecurityStamp *string `json:"securityStamp,omitempty"`
	Language      string  `json:"language"`
	DuelDeck      *string `json:"duelDeck,omitempty"`

	// Flavor
	FlavorText      *string `json:"flavorText,omitempty"`
	FlavorName      *string `json:"flavorName,omitempty"`
	FaceFlavorName  *string `json:"faceFlavorName,omitempty"`
	OriginalText    *string `json:"originalText,omitempty"`
	OriginalType    *string `json:"originalType,omitempty"`
	PrintedName     *string `json:"printedName,omitempty"`
	PrintedText     *string `json:"printedText,omitempty"`
	PrintedType     *string `json:"printedType,omitempty"`
	FacePrintedName *string `json:"facePrintedName,omitempty"`

	// Lists
	ArtistIDs            []string `json:"artistIds,omitempty"`
	Availability         []string `json:"availability"`
	BoosterTypes         []string `json:"boosterTypes,omitempty"`
	Finishes             []string `json:"finishes"`
	FrameEffects         []string `json:"frameEffects,omitempty"`
	Keywords             []string `json:"keywords,omitempty"`
	Printings            []string `json:"printings,omitempty"`
	PromoTypes           []string `json:"promoTypes,omitempty"`
	Variations           []string `json:"variations,omitempty"`
	OtherFaceIDs         []string `json:"otherFaceIds,omitempty"`
	CardParts            []string `json:"cardParts,omitempty"`
	OriginalPrintings    []string `json:"originalPrintings,omitempty"`
	RebalancedPrintings  []string `json:"rebalancedPrintings,omitempty"`
	Subsets              []string `json:"subsets,omitempty"`
	AttractionLights     []int    `json:"attractionLights,omitempty"`

	// Flags
	IsFullArt               *bool `json:"isFullArt,omitempty"`
	IsOnlineOnly            *bool `json:"isOnlineOnly,omitempty"`
	IsOversized             *bool `json:"isOversized,omitempty"`
	IsPromo                 *bool `json:"isPromo,omitempty"`
	IsReprint               *bool `json:"isReprint,omitempty"`
	IsTextless              *bool `json:"isTextless,omitempty"`
	IsFunny                 *bool `json:"isFunny,omitempty"`
	IsRebalanced            *bool `json:"isRebalanced,omitempty"`
	IsAlternative           *bool `json:"isAlternative,omitempty"`
	IsStorySpotlight        *bool `json:"isStorySpotlight,omitempty"`
	IsTimeshifted           *bool `json:"isTimeshifted,omitempty"`
	HasContentWarning       *bool `json:"hasContentWarning,omitempty"`
	HasAlternativeDeckLimit *bool `json:"hasAlternativeDeckLimit,omitempty"`
	IsReserved              *bool `json:"isReserved,omitempty"`
	IsGameChanger           *bool `json:"isGameChanger,omitempty"`

	// EDHREC
	EDHRECRank      *int     `json:"edhrecRank,omitempty"`
	EDHRECSaltiness *float64 `json:"edhrecSaltiness,omitempty"`

	// Dates
	FirstPrinting       *string `json:"firstPrinting,omitempty"`
	OriginalReleaseDate *string `json:"originalReleaseDate,omitempty"`

	// Nested sub-models
	IdentifiersData  Identifiers       `json:"identifiers"`
	LegalitiesData   Legalities        `json:"legalities"`
	PurchaseUrlsData PurchaseUrls      `json:"purchaseUrls"`
	LeadershipSkills *LeadershipSkills  `json:"leadershipSkills,omitempty"`
	RelatedCards     *RelatedCards      `json:"relatedCards,omitempty"`
	RulingsData      []Rulings          `json:"rulings,omitempty"`
	ForeignDataList  []ForeignData      `json:"foreignData,omitempty"`
	SourceProducts   *SourceProducts    `json:"sourceProducts,omitempty"`
}

// CardAtomic is oracle-like card data without printing-specific fields.
type CardAtomic struct {
	Name           string   `json:"name"`
	ASCIIName      *string  `json:"asciiName,omitempty"`
	FaceName       *string  `json:"faceName,omitempty"`
	Type           string   `json:"type"`
	Types          []string `json:"types"`
	Subtypes       []string `json:"subtypes"`
	Supertypes     []string `json:"supertypes"`
	Colors         []string `json:"colors"`
	ColorIdentity  []string `json:"colorIdentity"`
	ColorIndicator []string `json:"colorIndicator,omitempty"`
	ProducedMana   []string `json:"producedMana,omitempty"`
	ManaCost       *string  `json:"manaCost,omitempty"`
	ManaValue      float64  `json:"manaValue"`
	Text           *string  `json:"text,omitempty"`
	Layout         string   `json:"layout"`
	Side           *string  `json:"side,omitempty"`
	Power          *string  `json:"power,omitempty"`
	Toughness      *string  `json:"toughness,omitempty"`
	Loyalty        *string  `json:"loyalty,omitempty"`
	Defense        *string  `json:"defense,omitempty"`
	Hand           *string  `json:"hand,omitempty"`
	Life           *string  `json:"life,omitempty"`
	Keywords       []string `json:"keywords,omitempty"`
	Printings      []string `json:"printings,omitempty"`
	Subsets        []string `json:"subsets,omitempty"`

	EDHRECRank      *int     `json:"edhrecRank,omitempty"`
	EDHRECSaltiness *float64 `json:"edhrecSaltiness,omitempty"`
	IsFunny         *bool    `json:"isFunny,omitempty"`
	IsReserved      *bool    `json:"isReserved,omitempty"`
	IsGameChanger   *bool    `json:"isGameChanger,omitempty"`
	FirstPrinting   *string  `json:"firstPrinting,omitempty"`

	HasAlternativeDeckLimit *bool `json:"hasAlternativeDeckLimit,omitempty"`

	IdentifiersData  Identifiers       `json:"identifiers"`
	LegalitiesData   Legalities        `json:"legalities"`
	LeadershipSkills *LeadershipSkills  `json:"leadershipSkills,omitempty"`
	PurchaseUrlsData *PurchaseUrls      `json:"purchaseUrls,omitempty"`
	RelatedCards     *RelatedCards      `json:"relatedCards,omitempty"`
	RulingsData      []Rulings          `json:"rulings,omitempty"`
	ForeignDataList  []ForeignData      `json:"foreignData,omitempty"`
}

// CardToken represents a token card.
type CardToken struct {
	UUID           string   `json:"uuid"`
	Name           string   `json:"name"`
	ASCIIName      *string  `json:"asciiName,omitempty"`
	FaceName       *string  `json:"faceName,omitempty"`
	SetCode        string   `json:"setCode"`
	Number         string   `json:"number"`
	Type           string   `json:"type"`
	Types          []string `json:"types"`
	Subtypes       []string `json:"subtypes"`
	Supertypes     []string `json:"supertypes"`
	Colors         []string `json:"colors"`
	ColorIdentity  []string `json:"colorIdentity"`
	ColorIndicator []string `json:"colorIndicator,omitempty"`
	ProducedMana   []string `json:"producedMana,omitempty"`
	Power          *string  `json:"power,omitempty"`
	Toughness      *string  `json:"toughness,omitempty"`
	Text           *string  `json:"text,omitempty"`
	Layout         string   `json:"layout"`
	Artist         *string  `json:"artist,omitempty"`
	ArtistIDs      []string `json:"artistIds,omitempty"`
	BorderColor    string   `json:"borderColor"`
	FrameVersion   string   `json:"frameVersion"`
	FrameEffects   []string `json:"frameEffects,omitempty"`
	Availability   []string `json:"availability"`
	BoosterTypes   []string `json:"boosterTypes,omitempty"`
	Finishes       []string `json:"finishes"`
	Keywords       []string `json:"keywords,omitempty"`
	OtherFaceIDs   []string `json:"otherFaceIds,omitempty"`
	PromoTypes     []string `json:"promoTypes,omitempty"`
	ReverseRelated []string `json:"reverseRelated,omitempty"`
	Watermark      *string  `json:"watermark,omitempty"`
	Language       string   `json:"language"`
	Orientation    *string  `json:"orientation,omitempty"`

	IsFullArt    *bool `json:"isFullArt,omitempty"`
	IsOnlineOnly *bool `json:"isOnlineOnly,omitempty"`
	IsPromo      *bool `json:"isPromo,omitempty"`
	IsReprint    *bool `json:"isReprint,omitempty"`
	IsTextless   *bool `json:"isTextless,omitempty"`

	IdentifiersData Identifiers   `json:"identifiers"`
	RelatedCards    *RelatedCards  `json:"relatedCards,omitempty"`
	SourceProducts  *SourceProducts `json:"sourceProducts,omitempty"`
}

// CardDeck is a card in a preconstructed deck with count and foil flags.
type CardDeck struct {
	CardSet
	Count    int   `json:"count"`
	IsFoil   *bool `json:"isFoil,omitempty"`
	IsEtched *bool `json:"isEtched,omitempty"`
}

// CardSetDeck is a minimal card reference in a deck.
type CardSetDeck struct {
	Count  int    `json:"count"`
	IsFoil *bool  `json:"isFoil,omitempty"`
	UUID   string `json:"uuid"`
}

// CardLegality is a lightweight result for banned/restricted queries.
type CardLegality struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}
