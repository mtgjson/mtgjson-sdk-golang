package models

// SetList contains set summary metadata.
type SetList struct {
	Code             string          `json:"code"`
	Name             string          `json:"name"`
	Type             string          `json:"type"`
	ReleaseDate      string          `json:"releaseDate"`
	BaseSetSize      int             `json:"baseSetSize"`
	TotalSetSize     int             `json:"totalSetSize"`
	KeyruneCode      string          `json:"keyruneCode"`
	Block            *string         `json:"block,omitempty"`
	ParentCode       *string         `json:"parentCode,omitempty"`
	MTGOCode         *string         `json:"mtgoCode,omitempty"`
	TokenSetCode     *string         `json:"tokenSetCode,omitempty"`
	MCMID            *int            `json:"mcmId,omitempty"`
	MCMIDExtras      *int            `json:"mcmIdExtras,omitempty"`
	MCMName          *string         `json:"mcmName,omitempty"`
	TCGPlayerGroupID *int            `json:"tcgplayerGroupId,omitempty"`
	CardsphereSetID  *int            `json:"cardsphereSetId,omitempty"`
	IsFoilOnly       bool            `json:"isFoilOnly"`
	IsNonFoilOnly    *bool           `json:"isNonFoilOnly,omitempty"`
	IsOnlineOnly     bool            `json:"isOnlineOnly"`
	IsPaperOnly      *bool           `json:"isPaperOnly,omitempty"`
	IsForeignOnly    *bool           `json:"isForeignOnly,omitempty"`
	IsPartialPreview *bool           `json:"isPartialPreview,omitempty"`
	Translations     Translations    `json:"translations,omitempty"`
	Languages        []string        `json:"languages,omitempty"`
	Decks            []DeckSet       `json:"decks,omitempty"`
	SealedProduct    []SealedProduct `json:"sealedProduct,omitempty"`
}

// MtgSet is a full set with cards, tokens, and booster configuration.
type MtgSet struct {
	SetList
	Cards   []CardSet                 `json:"cards"`
	Tokens  []CardToken               `json:"tokens"`
	Booster map[string]BoosterConfig  `json:"booster,omitempty"`
}

// DeckSet is a deck within a set, with minimal card references.
type DeckSet struct {
	Code               string        `json:"code"`
	Name               string        `json:"name"`
	Type               string        `json:"type"`
	ReleaseDate        *string       `json:"releaseDate,omitempty"`
	SealedProductUUIDs []string      `json:"sealedProductUuids,omitempty"`
	MainBoard          []CardSetDeck `json:"mainBoard"`
	SideBoard          []CardSetDeck `json:"sideBoard"`
	Commander          []CardSetDeck `json:"commander,omitempty"`
	DisplayCommander   []CardSetDeck `json:"displayCommander,omitempty"`
	TokensBoard        []CardSetDeck `json:"tokens,omitempty"`
	Planes             []CardSetDeck `json:"planes,omitempty"`
	Schemes            []CardSetDeck `json:"schemes,omitempty"`
	SourceSetCodes     []string      `json:"sourceSetCodes,omitempty"`
}
