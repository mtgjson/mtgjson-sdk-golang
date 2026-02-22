package queries

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/mtgjson/mtgjson-sdk-go/db"
	"github.com/mtgjson/mtgjson-sdk-go/models"
)

// DeckQuery provides methods to query preconstructed deck data.
// Decks are loaded from DeckList.json on the CDN (not parquet).
type DeckQuery struct {
	cache  *db.CacheManager
	data   []map[string]any
	loaded bool
}

func NewDeckQuery(cache *db.CacheManager) *DeckQuery {
	return &DeckQuery{cache: cache}
}

func (q *DeckQuery) ensure(ctx context.Context) error {
	if q.loaded {
		return nil
	}
	raw, err := q.cache.LoadJSON(ctx, "deck_list")
	if err != nil {
		// If file not found, treat as empty
		q.data = nil
		q.loaded = true
		return nil
	}
	dataRaw, ok := raw["data"]
	if !ok {
		q.data = nil
		q.loaded = true
		return nil
	}
	// data is an array of objects
	jsonBytes, err := json.Marshal(dataRaw)
	if err != nil {
		q.data = nil
		q.loaded = true
		return nil
	}
	var decks []map[string]any
	if err := json.Unmarshal(jsonBytes, &decks); err != nil {
		q.data = nil
		q.loaded = true
		return nil
	}
	q.data = decks
	q.loaded = true
	return nil
}

// List returns available decks with optional filters.
func (q *DeckQuery) List(ctx context.Context, params ListDecksParams) ([]models.DeckList, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}
	results := q.data
	if params.SetCode != "" {
		codeUpper := strings.ToUpper(params.SetCode)
		var filtered []map[string]any
		for _, d := range results {
			if code, _ := d["code"].(string); strings.ToUpper(code) == codeUpper {
				filtered = append(filtered, d)
			}
		}
		results = filtered
	}
	if params.DeckType != "" {
		var filtered []map[string]any
		for _, d := range results {
			if dt, _ := d["type"].(string); dt == params.DeckType {
				filtered = append(filtered, d)
			}
		}
		results = filtered
	}
	return marshalDeckLists(results)
}

// Search searches decks by name substring with optional set code filter.
func (q *DeckQuery) Search(ctx context.Context, params SearchDecksParams) ([]models.DeckList, error) {
	if err := q.ensure(ctx); err != nil {
		return nil, err
	}
	results := q.data
	if params.Name != "" {
		nameLower := strings.ToLower(params.Name)
		var filtered []map[string]any
		for _, d := range results {
			if name, _ := d["name"].(string); strings.Contains(strings.ToLower(name), nameLower) {
				filtered = append(filtered, d)
			}
		}
		results = filtered
	}
	if params.SetCode != "" {
		codeUpper := strings.ToUpper(params.SetCode)
		var filtered []map[string]any
		for _, d := range results {
			if code, _ := d["code"].(string); strings.ToUpper(code) == codeUpper {
				filtered = append(filtered, d)
			}
		}
		results = filtered
	}
	return marshalDeckLists(results)
}

// Count returns the total number of available decks.
func (q *DeckQuery) Count(ctx context.Context) (int, error) {
	if err := q.ensure(ctx); err != nil {
		return 0, err
	}
	if q.data == nil {
		return 0, nil
	}
	return len(q.data), nil
}

// ListDecksParams contains optional filters for listing decks.
type ListDecksParams struct {
	SetCode  string
	DeckType string
}

// SearchDecksParams contains optional filters for searching decks.
type SearchDecksParams struct {
	Name    string
	SetCode string
}

func marshalDeckLists(data []map[string]any) ([]models.DeckList, error) {
	if len(data) == 0 {
		return nil, nil
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var result []models.DeckList
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}
	return result, nil
}
