package queries

import (
	"context"
	"fmt"

	"github.com/mtgjson/mtgjson-sdk-go/db"
	"github.com/mtgjson/mtgjson-sdk-go/models"
)

// SearchCardsParams contains all optional filters for card search.
// Zero values are ignored. Use pointer types for fields where zero is a valid filter.
type SearchCardsParams struct {
	Name          string
	FuzzyName     string
	LocalizedName string
	SetCode       string
	Colors        []string
	ColorIdentity []string
	Types         string
	Rarity        string
	LegalIn       string
	ManaValue     *float64
	ManaValueLTE  *float64
	ManaValueGTE  *float64
	Text          string
	TextRegex     string
	Power         string
	Toughness     string
	Artist        string
	Keyword       string
	IsPromo       *bool
	Availability  string
	Language      string
	Layout        string
	SetType       string
	Limit         int // 0 means default (100)
	Offset        int
}

// CardQuery provides methods to search, filter, and retrieve card data.
type CardQuery struct {
	conn *db.Connection
}

func NewCardQuery(conn *db.Connection) *CardQuery {
	return &CardQuery{conn: conn}
}

// GetByUUID returns a single card by its MTGJSON UUID, or nil if not found.
func (q *CardQuery) GetByUUID(ctx context.Context, uuid string) (*models.CardSet, error) {
	if err := q.conn.EnsureViews(ctx, "cards"); err != nil {
		return nil, err
	}
	var cards []models.CardSet
	if err := q.conn.ExecuteInto(ctx, &cards, "SELECT * FROM cards WHERE uuid = $1", uuid); err != nil {
		return nil, err
	}
	if len(cards) == 0 {
		return nil, nil
	}
	return &cards[0], nil
}

// GetByUUIDs fetches multiple cards by UUID in a single query.
func (q *CardQuery) GetByUUIDs(ctx context.Context, uuids []string) ([]models.CardSet, error) {
	if len(uuids) == 0 {
		return []models.CardSet{}, nil
	}
	if err := q.conn.EnsureViews(ctx, "cards"); err != nil {
		return nil, err
	}
	b := db.NewSQLBuilder("cards")
	vals := make([]any, len(uuids))
	for i, u := range uuids {
		vals[i] = u
	}
	b.WhereIn("uuid", vals)
	sql, params := b.Build()
	var cards []models.CardSet
	if err := q.conn.ExecuteInto(ctx, &cards, sql, params...); err != nil {
		return nil, err
	}
	return cards, nil
}

// GetByName returns all printings of a card by exact name.
func (q *CardQuery) GetByName(ctx context.Context, name string, setCode ...string) ([]models.CardSet, error) {
	if err := q.conn.EnsureViews(ctx, "cards"); err != nil {
		return nil, err
	}
	b := db.NewSQLBuilder("cards").WhereEq("name", name)
	if len(setCode) > 0 && setCode[0] != "" {
		b.WhereEq("setCode", setCode[0])
	}
	b.OrderBy("setCode DESC", "number ASC")
	sql, params := b.Build()
	var cards []models.CardSet
	if err := q.conn.ExecuteInto(ctx, &cards, sql, params...); err != nil {
		return nil, err
	}
	return cards, nil
}

// Search searches cards with flexible filters.
func (q *CardQuery) Search(ctx context.Context, p SearchCardsParams) ([]models.CardSet, error) {
	if err := q.conn.EnsureViews(ctx, "cards"); err != nil {
		return nil, err
	}
	b := db.NewSQLBuilder("cards")

	if p.Name != "" {
		if containsWildcard(p.Name) {
			b.WhereLike("name", p.Name)
		} else {
			b.WhereEq("name", p.Name)
		}
	}
	if p.FuzzyName != "" {
		b.WhereFuzzy("cards.name", p.FuzzyName, 0.8)
	}
	if p.SetCode != "" {
		b.WhereEq("setCode", p.SetCode)
	}
	if p.Rarity != "" {
		b.WhereEq("rarity", p.Rarity)
	}
	if p.ManaValue != nil {
		b.WhereEq("manaValue", *p.ManaValue)
	}
	if p.ManaValueLTE != nil {
		b.WhereLTE("manaValue", *p.ManaValueLTE)
	}
	if p.ManaValueGTE != nil {
		b.WhereGTE("manaValue", *p.ManaValueGTE)
	}
	if p.Text != "" {
		b.WhereLike("text", "%"+p.Text+"%")
	}
	if p.TextRegex != "" {
		b.WhereRegex("text", p.TextRegex)
	}
	if p.Types != "" {
		b.WhereLike("type", "%"+p.Types+"%")
	}
	if p.Power != "" {
		b.WhereEq("power", p.Power)
	}
	if p.Toughness != "" {
		b.WhereEq("toughness", p.Toughness)
	}
	if p.Artist != "" {
		b.WhereLike("artist", "%"+p.Artist+"%")
	}
	if p.Language != "" {
		b.WhereEq("language", p.Language)
	}
	if p.Layout != "" {
		b.WhereEq("layout", p.Layout)
	}
	if p.IsPromo != nil {
		b.WhereEq("isPromo", *p.IsPromo)
	}
	if len(p.Colors) > 0 {
		for _, color := range p.Colors {
			idx := b.AddParam(color)
			b.AddWhere(fmt.Sprintf("list_contains(colors, $%d)", idx))
		}
	}
	if len(p.ColorIdentity) > 0 {
		for _, color := range p.ColorIdentity {
			idx := b.AddParam(color)
			b.AddWhere(fmt.Sprintf("list_contains(colorIdentity, $%d)", idx))
		}
	}
	if p.Keyword != "" {
		idx := b.AddParam(p.Keyword)
		b.AddWhere(fmt.Sprintf("list_contains(keywords, $%d)", idx))
	}
	if p.Availability != "" {
		idx := b.AddParam(p.Availability)
		b.AddWhere(fmt.Sprintf("list_contains(availability, $%d)", idx))
	}
	if p.LocalizedName != "" {
		if err := q.conn.EnsureViews(ctx, "card_foreign_data"); err != nil {
			return nil, err
		}
		b.Select("cards.*")
		b.Join("JOIN card_foreign_data cfd ON cards.uuid = cfd.uuid")
		if containsWildcard(p.LocalizedName) {
			b.WhereLike("cfd.name", p.LocalizedName)
		} else {
			b.WhereEq("cfd.name", p.LocalizedName)
		}
	}
	if p.LegalIn != "" {
		if err := q.conn.EnsureViews(ctx, "card_legalities"); err != nil {
			return nil, err
		}
		b.Join("JOIN card_legalities cl ON cards.uuid = cl.uuid")
		b.WhereEq("cl.format", p.LegalIn)
		b.WhereEq("cl.status", "Legal")
	}
	if p.SetType != "" {
		if err := q.conn.EnsureViews(ctx, "sets"); err != nil {
			return nil, err
		}
		b.Select("cards.*")
		b.Join("JOIN sets s ON cards.setCode = s.code")
		b.WhereEq("s.type", p.SetType)
	}

	if p.FuzzyName != "" {
		idx := b.AddParam(p.FuzzyName)
		b.OrderBy(
			fmt.Sprintf("jaro_winkler_similarity(cards.name, $%d) DESC", idx),
			"cards.number ASC",
		)
	} else {
		b.OrderBy("cards.name ASC", "cards.number ASC")
	}

	limit := p.Limit
	if limit <= 0 {
		limit = 100
	}
	b.Limit(limit).Offset(p.Offset)

	sql, params := b.Build()
	var cards []models.CardSet
	if err := q.conn.ExecuteInto(ctx, &cards, sql, params...); err != nil {
		return nil, err
	}
	return cards, nil
}

// GetPrintings returns all printings of a card across all sets.
func (q *CardQuery) GetPrintings(ctx context.Context, name string) ([]models.CardSet, error) {
	return q.GetByName(ctx, name)
}

// GetAtomic returns de-duplicated oracle card data by name.
// Falls back to searching by faceName for split/adventure/MDFC cards.
func (q *CardQuery) GetAtomic(ctx context.Context, name string) ([]models.CardAtomic, error) {
	if err := q.conn.EnsureViews(ctx, "cards"); err != nil {
		return nil, err
	}
	atomicCols := []string{
		"name", "asciiName", "faceName", "type", "types", "subtypes", "supertypes",
		"colors", "colorIdentity", "colorIndicator", "producedMana",
		"manaCost", "text", "layout", "side", "power", "toughness", "loyalty",
		"keywords", "isFunny", "edhrecSaltiness", "subsets",
		"manaValue", "faceManaValue", "defense", "hand", "life",
		"edhrecRank", "hasAlternativeDeckLimit", "isReserved", "isGameChanger",
		"printings", "leadershipSkills", "relatedCards",
	}
	b := db.NewSQLBuilder("cards")
	b.Select(atomicCols...)
	b.WhereEq("name", name)
	b.OrderBy("isFunny ASC NULLS FIRST", "isOnlineOnly ASC NULLS FIRST", "side ASC NULLS FIRST")
	sql, params := b.Build()

	var results []models.CardAtomic
	if err := q.conn.ExecuteInto(ctx, &results, sql, params...); err != nil {
		return nil, err
	}

	// Fallback: search by faceName for split/adventure/MDFC cards
	if len(results) == 0 {
		b2 := db.NewSQLBuilder("cards")
		b2.Select(atomicCols...)
		b2.Where("CAST(faceName AS VARCHAR) = $1", name)
		b2.OrderBy("isFunny ASC NULLS FIRST", "isOnlineOnly ASC NULLS FIRST", "side ASC NULLS FIRST")
		sql2, params2 := b2.Build()
		if err := q.conn.ExecuteInto(ctx, &results, sql2, params2...); err != nil {
			return nil, err
		}
	}

	if len(results) == 0 {
		return []models.CardAtomic{}, nil
	}

	// De-duplicate by name+faceName
	type key struct {
		name     string
		faceName string
	}
	seen := make(map[key]bool)
	var unique []models.CardAtomic
	for _, r := range results {
		fn := ""
		if r.FaceName != nil {
			fn = *r.FaceName
		}
		k := key{r.Name, fn}
		if !seen[k] {
			seen[k] = true
			unique = append(unique, r)
		}
	}
	return unique, nil
}

// FindByScryfallID finds cards by their Scryfall ID.
func (q *CardQuery) FindByScryfallID(ctx context.Context, scryfallID string) ([]models.CardSet, error) {
	if err := q.conn.EnsureViews(ctx, "cards", "card_identifiers"); err != nil {
		return nil, err
	}
	sql := "SELECT c.* FROM cards c JOIN card_identifiers ci ON c.uuid = ci.uuid WHERE ci.scryfallId = $1"
	var cards []models.CardSet
	if err := q.conn.ExecuteInto(ctx, &cards, sql, scryfallID); err != nil {
		return nil, err
	}
	return cards, nil
}

// Random returns randomly sampled cards.
func (q *CardQuery) Random(ctx context.Context, count int) ([]models.CardSet, error) {
	if err := q.conn.EnsureViews(ctx, "cards"); err != nil {
		return nil, err
	}
	sql := fmt.Sprintf("SELECT * FROM cards USING SAMPLE %d", count)
	var cards []models.CardSet
	if err := q.conn.ExecuteInto(ctx, &cards, sql); err != nil {
		return nil, err
	}
	return cards, nil
}

// Count returns the number of cards matching optional column filters.
func (q *CardQuery) Count(ctx context.Context, filters ...Filter) (int, error) {
	if err := q.conn.EnsureViews(ctx, "cards"); err != nil {
		return 0, err
	}
	if len(filters) == 0 {
		val, err := q.conn.ExecuteScalar(ctx, "SELECT COUNT(*) FROM cards")
		if err != nil {
			return 0, err
		}
		return db.ScalarToInt(val), nil
	}
	b := db.NewSQLBuilder("cards").Select("COUNT(*)")
	for _, f := range filters {
		b.WhereEq(f.Column, f.Value)
	}
	sql, params := b.Build()
	val, err := q.conn.ExecuteScalar(ctx, sql, params...)
	if err != nil {
		return 0, err
	}
	return db.ScalarToInt(val), nil
}

// Filter is a simple column=value filter for Count methods.
type Filter struct {
	Column string
	Value  any
}

func containsWildcard(s string) bool {
	return len(s) > 0 && (s[0] == '%' || s[len(s)-1] == '%' || contains(s, "%"))
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
