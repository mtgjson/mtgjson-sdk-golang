package models

// DeckList is a summary deck entry without card details.
type DeckList struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	FileName    string  `json:"fileName"`
	Type        string  `json:"type"`
	ReleaseDate *string `json:"releaseDate,omitempty"`
}

// Deck is a full deck with expanded card references.
type Deck struct {
	Code               string      `json:"code"`
	Name               string      `json:"name"`
	Type               string      `json:"type"`
	ReleaseDate        *string     `json:"releaseDate,omitempty"`
	SealedProductUUIDs []string    `json:"sealedProductUuids,omitempty"`
	MainBoard          []CardDeck  `json:"mainBoard"`
	SideBoard          []CardDeck  `json:"sideBoard"`
	Commander          []CardDeck  `json:"commander,omitempty"`
	DisplayCommander   []CardDeck  `json:"displayCommander,omitempty"`
	Planes             []CardDeck  `json:"planes,omitempty"`
	Schemes            []CardDeck  `json:"schemes,omitempty"`
	Tokens             []CardToken `json:"tokens,omitempty"`
	SourceSetCodes     []string    `json:"sourceSetCodes,omitempty"`
}
