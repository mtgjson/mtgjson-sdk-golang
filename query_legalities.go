package mtgjson

import (
	"context"
	"fmt"
)

// LegalityQuery provides methods to query card format legalities.
type LegalityQuery struct {
	conn *Connection
}

func newLegalityQuery(conn *Connection) *LegalityQuery {
	return &LegalityQuery{conn: conn}
}

func (q *LegalityQuery) ensure(ctx context.Context) error {
	return q.conn.EnsureViews(ctx, "card_legalities")
}

// cardsByStatus returns cards with a specific legality status in a format.
func (q *LegalityQuery) cardsByStatus(ctx context.Context, formatName, status string, limit, offset int) ([]CardLegality, error) {
	if err := q.conn.EnsureViews(ctx, "cards", "card_legalities"); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 100
	}
	sql := fmt.Sprintf(
		"SELECT c.name, c.uuid FROM cards c "+
			"JOIN card_legalities cl ON c.uuid = cl.uuid "+
			"WHERE cl.format = $1 AND cl.status = $2 "+
			"ORDER BY c.name ASC "+
			"LIMIT %d OFFSET %d", limit, offset)
	var results []CardLegality
	if err := q.conn.ExecuteInto(ctx, &results, sql, formatName, status); err != nil {
		return nil, err
	}
	return results, nil
}

// FormatsForCard returns all format legalities for a card UUID.
// Returns a map of format name to status (e.g. {"modern": "Legal"}).
func (q *LegalityQuery) FormatsForCard(ctx context.Context, uuid string) (map[string]string, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}
	rows, err := q.conn.Execute(ctx, "SELECT format, status FROM card_legalities WHERE uuid = $1", uuid)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, r := range rows {
		f, _ := r["format"].(string)
		s, _ := r["status"].(string)
		if f != "" {
			result[f] = s
		}
	}
	return result, nil
}

// LegalIn returns all cards legal in a specific format.
func (q *LegalityQuery) LegalIn(ctx context.Context, formatName string, limit ...int) ([]CardSet, error) {
	if err := q.conn.EnsureViews(ctx, "cards", "card_legalities"); err != nil {
		return nil, err
	}
	lim := 100
	if len(limit) > 0 && limit[0] > 0 {
		lim = limit[0]
	}
	sql := fmt.Sprintf(
		"SELECT DISTINCT c.* FROM cards c "+
			"JOIN card_legalities cl ON c.uuid = cl.uuid "+
			"WHERE cl.format = $1 AND cl.status = 'Legal' "+
			"ORDER BY c.name ASC "+
			"LIMIT %d", lim)
	var cards []CardSet
	if err := q.conn.ExecuteInto(ctx, &cards, sql, formatName); err != nil {
		return nil, err
	}
	return cards, nil
}

// IsLegal checks if a card is legal in a specific format.
func (q *LegalityQuery) IsLegal(ctx context.Context, uuid, formatName string) (bool, error) {
	if err := q.ensure(ctx); err != nil {
		return false, err
	}
	val, err := q.conn.ExecuteScalar(ctx,
		"SELECT COUNT(*) FROM card_legalities WHERE uuid = $1 AND format = $2 AND status = 'Legal'",
		uuid, formatName)
	if err != nil {
		return false, err
	}
	return scalarToInt(val) > 0, nil
}

// BannedIn returns all cards banned in a specific format.
func (q *LegalityQuery) BannedIn(ctx context.Context, formatName string, limit ...int) ([]CardLegality, error) {
	lim := 100
	if len(limit) > 0 && limit[0] > 0 {
		lim = limit[0]
	}
	return q.cardsByStatus(ctx, formatName, "Banned", lim, 0)
}

// RestrictedIn returns all cards restricted in a specific format.
func (q *LegalityQuery) RestrictedIn(ctx context.Context, formatName string, limit ...int) ([]CardLegality, error) {
	lim := 100
	if len(limit) > 0 && limit[0] > 0 {
		lim = limit[0]
	}
	return q.cardsByStatus(ctx, formatName, "Restricted", lim, 0)
}

// SuspendedIn returns all cards suspended in a specific format.
func (q *LegalityQuery) SuspendedIn(ctx context.Context, formatName string, limit ...int) ([]CardLegality, error) {
	lim := 100
	if len(limit) > 0 && limit[0] > 0 {
		lim = limit[0]
	}
	return q.cardsByStatus(ctx, formatName, "Suspended", lim, 0)
}

// NotLegalIn returns all cards not legal in a specific format.
func (q *LegalityQuery) NotLegalIn(ctx context.Context, formatName string, limit ...int) ([]CardLegality, error) {
	lim := 100
	if len(limit) > 0 && limit[0] > 0 {
		lim = limit[0]
	}
	return q.cardsByStatus(ctx, formatName, "Not Legal", lim, 0)
}
