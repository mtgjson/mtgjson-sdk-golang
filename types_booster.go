package mtgjson

// BoosterConfig is the complete booster configuration for a set.
type BoosterConfig struct {
	Boosters           []BoosterPack            `json:"boosters"`
	BoostersTotalWeight int                     `json:"boostersTotalWeight"`
	Name               *string                  `json:"name,omitempty"`
	Sheets             map[string]BoosterSheet  `json:"sheets"`
	SourceSetCodes     []string                 `json:"sourceSetCodes"`
}

// BoosterPack is a possible pack configuration with its selection weight.
type BoosterPack struct {
	Contents map[string]int `json:"contents"`
	Weight   int            `json:"weight"`
}

// BoosterSheet defines a sheet from which cards are drawn.
type BoosterSheet struct {
	AllowDuplicates *bool          `json:"allowDuplicates,omitempty"`
	BalanceColors   *bool          `json:"balanceColors,omitempty"`
	Cards           map[string]int `json:"cards"`
	Foil            bool           `json:"foil"`
	Fixed           *bool          `json:"fixed,omitempty"`
	TotalWeight     int            `json:"totalWeight"`
}
