package mtgjson

// PriceRow is a single flattened price data point from the prices table.
type PriceRow struct {
	UUID     string  `json:"uuid"`
	Source   string  `json:"source"`
	Provider string  `json:"provider"`
	Currency string  `json:"currency"`
	Category string  `json:"category"`
	Finish   string  `json:"finish"`
	Date     string  `json:"date"`
	Price    float64 `json:"price"`
}

// PriceTrend contains aggregate price statistics over time.
type PriceTrend struct {
	MinPrice   float64 `json:"min_price"`
	MaxPrice   float64 `json:"max_price"`
	AvgPrice   float64 `json:"avg_price"`
	FirstDate  string  `json:"first_date"`
	LastDate   string  `json:"last_date"`
	DataPoints int64   `json:"data_points"`
}

// FinancialSummary contains aggregate price data for a set.
type FinancialSummary struct {
	TotalValue float64 `json:"total_value"`
	AvgValue   float64 `json:"avg_value"`
	MinValue   float64 `json:"min_value"`
	MaxValue   float64 `json:"max_value"`
	CardCount  int64   `json:"card_count"`
	Date       string  `json:"date"`
}

// PricePrinting represents a card printing with its price info.
type PricePrinting struct {
	Name     string  `json:"name"`
	SetCode  string  `json:"cheapest_set"`
	Number   string  `json:"cheapest_number"`
	UUID     string  `json:"cheapest_uuid"`
	MinPrice float64 `json:"min_price"`
}

// ExpensivePrinting represents an expensive card printing.
type ExpensivePrinting struct {
	Name     string  `json:"name"`
	SetCode  string  `json:"priciest_set"`
	Number   string  `json:"priciest_number"`
	UUID     string  `json:"priciest_uuid"`
	MaxPrice float64 `json:"max_price"`
}
