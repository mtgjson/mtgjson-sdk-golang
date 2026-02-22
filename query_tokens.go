package mtgjson

import (
	"context"
	"fmt"
)

// SearchTokensParams contains filters for searching tokens.
type SearchTokensParams struct {
	Name    string
	SetCode string
	Colors  []string
	Types   string
	Artist  string
	Limit   int // 0 means default (100)
	Offset  int
}

// TokenQuery provides methods to search and retrieve token card data.
type TokenQuery struct {
	conn *Connection
}

func newTokenQuery(conn *Connection) *TokenQuery {
	return &TokenQuery{conn: conn}
}

// GetByUUID returns a single token by its MTGJSON UUID, or nil if not found.
func (q *TokenQuery) GetByUUID(ctx context.Context, uuid string) (*CardToken, error) {
	if err := q.conn.EnsureViews(ctx, "tokens"); err != nil {
		return nil, err
	}
	var tokens []CardToken
	if err := q.conn.ExecuteInto(ctx, &tokens, "SELECT * FROM tokens WHERE uuid = $1", uuid); err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, nil
	}
	return &tokens[0], nil
}

// GetByUUIDs fetches multiple tokens by UUID in a single query.
func (q *TokenQuery) GetByUUIDs(ctx context.Context, uuids []string) ([]CardToken, error) {
	if len(uuids) == 0 {
		return []CardToken{}, nil
	}
	if err := q.conn.EnsureViews(ctx, "tokens"); err != nil {
		return nil, err
	}
	b := NewSQLBuilder("tokens")
	vals := make([]any, len(uuids))
	for i, u := range uuids {
		vals[i] = u
	}
	b.WhereIn("uuid", vals)
	sql, params := b.Build()
	var tokens []CardToken
	if err := q.conn.ExecuteInto(ctx, &tokens, sql, params...); err != nil {
		return nil, err
	}
	return tokens, nil
}

// GetByName returns all tokens matching an exact name.
func (q *TokenQuery) GetByName(ctx context.Context, name string, setCode ...string) ([]CardToken, error) {
	if err := q.conn.EnsureViews(ctx, "tokens"); err != nil {
		return nil, err
	}
	b := NewSQLBuilder("tokens").WhereEq("name", name)
	if len(setCode) > 0 && setCode[0] != "" {
		b.WhereEq("setCode", setCode[0])
	}
	b.OrderBy("setCode DESC", "number ASC")
	sql, params := b.Build()
	var tokens []CardToken
	if err := q.conn.ExecuteInto(ctx, &tokens, sql, params...); err != nil {
		return nil, err
	}
	return tokens, nil
}

// Search searches tokens with flexible filters.
func (q *TokenQuery) Search(ctx context.Context, p SearchTokensParams) ([]CardToken, error) {
	if err := q.conn.EnsureViews(ctx, "tokens"); err != nil {
		return nil, err
	}
	b := NewSQLBuilder("tokens")
	if p.Name != "" {
		if containsWildcard(p.Name) {
			b.WhereLike("name", p.Name)
		} else {
			b.WhereEq("name", p.Name)
		}
	}
	if p.SetCode != "" {
		b.WhereEq("setCode", p.SetCode)
	}
	if p.Types != "" {
		b.WhereLike("type", "%"+p.Types+"%")
	}
	if p.Artist != "" {
		b.WhereLike("artist", "%"+p.Artist+"%")
	}
	if len(p.Colors) > 0 {
		for _, color := range p.Colors {
			idx := b.AddParam(color)
			b.AddWhere(fmt.Sprintf("list_contains(colors, $%d)", idx))
		}
	}
	b.OrderBy("name ASC", "number ASC")
	limit := p.Limit
	if limit <= 0 {
		limit = 100
	}
	b.Limit(limit).Offset(p.Offset)
	sql, params := b.Build()
	var tokens []CardToken
	if err := q.conn.ExecuteInto(ctx, &tokens, sql, params...); err != nil {
		return nil, err
	}
	return tokens, nil
}

// ForSet returns all tokens for a specific set.
func (q *TokenQuery) ForSet(ctx context.Context, setCode string) ([]CardToken, error) {
	return q.Search(ctx, SearchTokensParams{SetCode: setCode, Limit: 1000})
}

// Count returns the number of tokens matching optional column filters.
func (q *TokenQuery) Count(ctx context.Context, filters ...Filter) (int, error) {
	if err := q.conn.EnsureViews(ctx, "tokens"); err != nil {
		return 0, err
	}
	if len(filters) == 0 {
		val, err := q.conn.ExecuteScalar(ctx, "SELECT COUNT(*) FROM tokens")
		if err != nil {
			return 0, err
		}
		return scalarToInt(val), nil
	}
	b := NewSQLBuilder("tokens").Select("COUNT(*)")
	for _, f := range filters {
		b.WhereEq(f.Column, f.Value)
	}
	sql, params := b.Build()
	val, err := q.conn.ExecuteScalar(ctx, sql, params...)
	if err != nil {
		return 0, err
	}
	return scalarToInt(val), nil
}
