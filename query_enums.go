package mtgjson

import (
	"context"
)

// EnumQuery provides methods to query MTGJSON keywords, card types, and enum values.
// Data is loaded from JSON files on the CDN (not parquet).
type EnumQuery struct {
	cache *CacheManager
}

func newEnumQuery(cache *CacheManager) *EnumQuery {
	return &EnumQuery{cache: cache}
}

// Keywords returns all MTG keyword categories and their values.
// Returns a map like {"abilityWords": ["Addendum", ...], "keywordActions": [...]}.
func (q *EnumQuery) Keywords(ctx context.Context) (map[string]any, error) {
	raw, err := q.cache.LoadJSON(ctx, "keywords")
	if err != nil {
		return nil, err
	}
	if data, ok := raw["data"].(map[string]any); ok {
		return data, nil
	}
	return map[string]any{}, nil
}

// CardTypes returns all card types with their valid sub- and supertypes.
// Returns a map like {"creature": {"subTypes": [...], "superTypes": [...]}}.
func (q *EnumQuery) CardTypes(ctx context.Context) (map[string]any, error) {
	raw, err := q.cache.LoadJSON(ctx, "card_types")
	if err != nil {
		return nil, err
	}
	if data, ok := raw["data"].(map[string]any); ok {
		return data, nil
	}
	return map[string]any{}, nil
}

// EnumValues returns all enumerated values used by MTGJSON fields.
// Returns a map like {"colors": ["B", "G", "R", "U", "W"], ...}.
func (q *EnumQuery) EnumValues(ctx context.Context) (map[string]any, error) {
	raw, err := q.cache.LoadJSON(ctx, "enum_values")
	if err != nil {
		return nil, err
	}
	if data, ok := raw["data"].(map[string]any); ok {
		return data, nil
	}
	return map[string]any{}, nil
}
