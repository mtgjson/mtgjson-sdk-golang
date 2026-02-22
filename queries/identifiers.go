package queries

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/mtgjson/mtgjson-sdk-go/db"
	"github.com/mtgjson/mtgjson-sdk-go/models"
)

// KnownIDColumns contains all valid identifier column names.
var KnownIDColumns = map[string]bool{
	"cardKingdomEtchedId":      true,
	"cardKingdomFoilId":        true,
	"cardKingdomId":            true,
	"cardsphereId":             true,
	"cardsphereFoilId":         true,
	"mcmId":                    true,
	"mcmMetaId":                true,
	"mtgArenaId":               true,
	"mtgoFoilId":               true,
	"mtgoId":                   true,
	"multiverseId":             true,
	"scryfallId":               true,
	"scryfallIllustrationId":   true,
	"scryfallOracleId":         true,
	"tcgplayerEtchedProductId": true,
	"tcgplayerProductId":       true,
}

// IdentifierQuery provides cross-reference lookups by external IDs.
type IdentifierQuery struct {
	conn *db.Connection
}

func NewIdentifierQuery(conn *db.Connection) *IdentifierQuery {
	return &IdentifierQuery{conn: conn}
}

func (q *IdentifierQuery) ensure(ctx context.Context) error {
	return q.conn.EnsureViews(ctx, "cards", "card_identifiers")
}

func (q *IdentifierQuery) findBy(ctx context.Context, idColumn, value string) ([]models.CardSet, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(
		"SELECT c.* FROM cards c JOIN card_identifiers ci ON c.uuid = ci.uuid WHERE ci.\"%s\" = $1",
		idColumn)
	var cards []models.CardSet
	if err := q.conn.ExecuteInto(ctx, &cards, sql, value); err != nil {
		return nil, err
	}
	return cards, nil
}

// FindBy performs a generic identifier lookup by column name.
// Returns an error if idType is not a known identifier column.
func (q *IdentifierQuery) FindBy(ctx context.Context, idType, value string) ([]models.CardSet, error) {
	if !KnownIDColumns[idType] {
		known := make([]string, 0, len(KnownIDColumns))
		for k := range KnownIDColumns {
			known = append(known, k)
		}
		sort.Strings(known)
		return nil, fmt.Errorf("unknown identifier type %q; known types: %s", idType, strings.Join(known, ", "))
	}
	return q.findBy(ctx, idType, value)
}

// FindByScryfallID finds cards by Scryfall UUID.
func (q *IdentifierQuery) FindByScryfallID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "scryfallId", id)
}

// FindByScryfallOracleID finds cards by Scryfall Oracle ID.
func (q *IdentifierQuery) FindByScryfallOracleID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "scryfallOracleId", id)
}

// FindByScryfallIllustrationID finds cards by Scryfall Illustration ID.
func (q *IdentifierQuery) FindByScryfallIllustrationID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "scryfallIllustrationId", id)
}

// FindByTCGPlayerID finds cards by TCGPlayer product ID.
func (q *IdentifierQuery) FindByTCGPlayerID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "tcgplayerProductId", id)
}

// FindByTCGPlayerEtchedID finds cards by TCGPlayer etched product ID.
func (q *IdentifierQuery) FindByTCGPlayerEtchedID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "tcgplayerEtchedProductId", id)
}

// FindByMTGOID finds cards by MTGO ID.
func (q *IdentifierQuery) FindByMTGOID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "mtgoId", id)
}

// FindByMTGOFoilID finds cards by MTGO foil ID.
func (q *IdentifierQuery) FindByMTGOFoilID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "mtgoFoilId", id)
}

// FindByMTGArenaID finds cards by MTG Arena ID.
func (q *IdentifierQuery) FindByMTGArenaID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "mtgArenaId", id)
}

// FindByMultiverseID finds cards by Gatherer multiverse ID.
func (q *IdentifierQuery) FindByMultiverseID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "multiverseId", id)
}

// FindByMCMID finds cards by Cardmarket (MCM) product ID.
func (q *IdentifierQuery) FindByMCMID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "mcmId", id)
}

// FindByMCMMetaID finds cards by Cardmarket (MCM) meta ID.
func (q *IdentifierQuery) FindByMCMMetaID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "mcmMetaId", id)
}

// FindByCardKingdomID finds cards by Card Kingdom product ID.
func (q *IdentifierQuery) FindByCardKingdomID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "cardKingdomId", id)
}

// FindByCardKingdomFoilID finds cards by Card Kingdom foil product ID.
func (q *IdentifierQuery) FindByCardKingdomFoilID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "cardKingdomFoilId", id)
}

// FindByCardKingdomEtchedID finds cards by Card Kingdom etched product ID.
func (q *IdentifierQuery) FindByCardKingdomEtchedID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "cardKingdomEtchedId", id)
}

// FindByCardsphereID finds cards by Cardsphere ID.
func (q *IdentifierQuery) FindByCardsphereID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "cardsphereId", id)
}

// FindByCardsphereFoilID finds cards by Cardsphere foil ID.
func (q *IdentifierQuery) FindByCardsphereFoilID(ctx context.Context, id string) ([]models.CardSet, error) {
	return q.findBy(ctx, "cardsphereFoilId", id)
}

// GetIdentifiers returns all external identifiers for a card UUID.
func (q *IdentifierQuery) GetIdentifiers(ctx context.Context, uuid string) (map[string]any, error) {
	if err := q.conn.EnsureViews(ctx, "card_identifiers"); err != nil {
		return nil, err
	}
	rows, err := q.conn.Execute(ctx, "SELECT * FROM card_identifiers WHERE uuid = $1", uuid)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}
