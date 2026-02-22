package booster

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/mtgjson/mtgjson-sdk-go/db"
	"github.com/mtgjson/mtgjson-sdk-go/models"
)

// BoosterSimulator simulates opening booster packs using set booster configuration data.
// Uses weighted random selection based on the booster field in set data.
// Requires the booster column (present in AllPrintings, but NOT in the flat sets.parquet from CDN).
type BoosterSimulator struct {
	conn *db.Connection
}

func NewBoosterSimulator(conn *db.Connection) *BoosterSimulator {
	return &BoosterSimulator{conn: conn}
}

func (bs *BoosterSimulator) ensure(ctx context.Context) error {
	return bs.conn.EnsureViews(ctx, "sets", "cards")
}

// getBoosterConfig returns the booster configuration for a set.
func (bs *BoosterSimulator) getBoosterConfig(ctx context.Context, setCode string) (map[string]any, error) {
	if err := bs.ensure(ctx); err != nil {
		return nil, err
	}
	rows, err := bs.conn.Execute(ctx, "SELECT booster FROM sets WHERE code = $1", setCode)
	if err != nil {
		return nil, nil
	}
	if len(rows) == 0 {
		return nil, nil
	}
	boosterRaw := rows[0]["booster"]
	if boosterRaw == nil {
		return nil, nil
	}
	// May be a string (JSON), map, or DuckDB struct
	return extractBoosterConfig(boosterRaw), nil
}

// AvailableTypes lists available booster types for a set.
func (bs *BoosterSimulator) AvailableTypes(ctx context.Context, setCode string) ([]string, error) {
	config, err := bs.getBoosterConfig(ctx, setCode)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}
	types := make([]string, 0, len(config))
	for k := range config {
		types = append(types, k)
	}
	return types, nil
}

// OpenPack simulates opening a single booster pack.
func (bs *BoosterSimulator) OpenPack(ctx context.Context, setCode, boosterType string) ([]models.CardSet, error) {
	configs, err := bs.getBoosterConfig(ctx, setCode)
	if err != nil {
		return nil, err
	}
	if configs == nil {
		return nil, fmt.Errorf("mtgjson: no booster config for set %q", setCode)
	}
	configRaw, ok := configs[boosterType]
	if !ok {
		types := make([]string, 0, len(configs))
		for k := range configs {
			types = append(types, k)
		}
		return nil, fmt.Errorf("mtgjson: no booster type %q for set %q; available: %v", boosterType, setCode, types)
	}
	config, ok := configRaw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("mtgjson: invalid booster config type for %q/%q", setCode, boosterType)
	}

	boostersRaw, _ := config["boosters"].([]any)
	sheetsRaw, _ := config["sheets"].(map[string]any)

	// Pick a pack template
	packTemplate := pickPack(boostersRaw)
	if packTemplate == nil {
		return nil, nil
	}

	contents, _ := packTemplate["contents"].(map[string]any)
	var cardUUIDs []string
	for sheetName, countRaw := range contents {
		count := db.ToInt(countRaw)
		if count <= 0 {
			continue
		}
		sheetRaw, ok := sheetsRaw[sheetName]
		if !ok {
			continue
		}
		sheet, ok := sheetRaw.(map[string]any)
		if !ok {
			continue
		}
		picked := pickFromSheet(sheet, count)
		cardUUIDs = append(cardUUIDs, picked...)
	}

	if len(cardUUIDs) == 0 {
		return nil, nil
	}

	// Fetch card data
	placeholders := ""
	params := make([]any, len(cardUUIDs))
	for i, uuid := range cardUUIDs {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += fmt.Sprintf("$%d", i+1)
		params[i] = uuid
	}
	sql := fmt.Sprintf("SELECT * FROM cards WHERE uuid IN (%s)", placeholders)

	var cards []models.CardSet
	if err := bs.conn.ExecuteInto(ctx, &cards, sql, params...); err != nil {
		return nil, err
	}

	// Preserve pack order
	uuidToCard := make(map[string]models.CardSet, len(cards))
	for _, c := range cards {
		uuidToCard[c.UUID] = c
	}
	ordered := make([]models.CardSet, 0, len(cardUUIDs))
	for _, uuid := range cardUUIDs {
		if c, ok := uuidToCard[uuid]; ok {
			ordered = append(ordered, c)
		}
	}
	return ordered, nil
}

// OpenBox simulates opening a booster box (multiple packs).
func (bs *BoosterSimulator) OpenBox(ctx context.Context, setCode, boosterType string, packs int) ([][]models.CardSet, error) {
	if packs <= 0 {
		packs = 36
	}
	box := make([][]models.CardSet, 0, packs)
	for i := 0; i < packs; i++ {
		pack, err := bs.OpenPack(ctx, setCode, boosterType)
		if err != nil {
			return nil, err
		}
		box = append(box, pack)
	}
	return box, nil
}

// SheetContents returns the card UUIDs and weights for a specific booster sheet.
func (bs *BoosterSimulator) SheetContents(ctx context.Context, setCode, boosterType, sheetName string) (map[string]int, error) {
	configs, err := bs.getBoosterConfig(ctx, setCode)
	if err != nil {
		return nil, err
	}
	if configs == nil {
		return nil, nil
	}
	configRaw, ok := configs[boosterType]
	if !ok {
		return nil, nil
	}
	config, ok := configRaw.(map[string]any)
	if !ok {
		return nil, nil
	}
	sheetsRaw, _ := config["sheets"].(map[string]any)
	sheetRaw, ok := sheetsRaw[sheetName]
	if !ok {
		return nil, nil
	}
	sheet, ok := sheetRaw.(map[string]any)
	if !ok {
		return nil, nil
	}
	cardsRaw, _ := sheet["cards"].(map[string]any)
	if cardsRaw == nil {
		return nil, nil
	}
	result := make(map[string]int, len(cardsRaw))
	for uuid, weightRaw := range cardsRaw {
		result[uuid] = db.ToInt(weightRaw)
	}
	return result, nil
}

// pickPack does a weighted random selection of a pack template.
func pickPack(boosters []any) map[string]any {
	if len(boosters) == 0 {
		return nil
	}
	type entry struct {
		pack   map[string]any
		weight float64
	}
	var entries []entry
	totalWeight := 0.0
	for _, b := range boosters {
		m, ok := b.(map[string]any)
		if !ok {
			continue
		}
		w := db.ToFloat64(m["weight"])
		if w <= 0 {
			w = 1
		}
		entries = append(entries, entry{pack: m, weight: w})
		totalWeight += w
	}
	if len(entries) == 0 {
		return nil
	}
	r := rand.Float64() * totalWeight
	cumulative := 0.0
	for _, e := range entries {
		cumulative += e.weight
		if r < cumulative {
			return e.pack
		}
	}
	return entries[len(entries)-1].pack
}

// pickFromSheet does weighted random selection of cards from a sheet.
func pickFromSheet(sheet map[string]any, count int) []string {
	cardsRaw, _ := sheet["cards"].(map[string]any)
	if cardsRaw == nil {
		return nil
	}
	allowDuplicates, _ := sheet["allowDuplicates"].(bool)

	uuids := make([]string, 0, len(cardsRaw))
	weights := make([]float64, 0, len(cardsRaw))
	for uuid, weightRaw := range cardsRaw {
		uuids = append(uuids, uuid)
		weights = append(weights, db.ToFloat64(weightRaw))
	}

	if allowDuplicates {
		return weightedChoicesWithReplacement(uuids, weights, count)
	}

	if count >= len(uuids) {
		result := make([]string, len(uuids))
		copy(result, uuids)
		rand.Shuffle(len(result), func(i, j int) { result[i], result[j] = result[j], result[i] })
		return result
	}

	return weightedChoicesWithoutReplacement(uuids, weights, count)
}

func weightedChoicesWithReplacement(items []string, weights []float64, count int) []string {
	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}
	result := make([]string, count)
	for i := 0; i < count; i++ {
		r := rand.Float64() * totalWeight
		cumulative := 0.0
		for j, w := range weights {
			cumulative += w
			if r < cumulative {
				result[i] = items[j]
				break
			}
		}
		if result[i] == "" {
			result[i] = items[len(items)-1]
		}
	}
	return result
}

func weightedChoicesWithoutReplacement(items []string, weights []float64, count int) []string {
	remaining := make([]string, len(items))
	copy(remaining, items)
	remainingWeights := make([]float64, len(weights))
	copy(remainingWeights, weights)

	picked := make([]string, 0, count)
	for i := 0; i < count && len(remaining) > 0; i++ {
		totalWeight := 0.0
		for _, w := range remainingWeights {
			totalWeight += w
		}
		r := rand.Float64() * totalWeight
		cumulative := 0.0
		idx := len(remaining) - 1
		for j, w := range remainingWeights {
			cumulative += w
			if r < cumulative {
				idx = j
				break
			}
		}
		picked = append(picked, remaining[idx])
		remaining = append(remaining[:idx], remaining[idx+1:]...)
		remainingWeights = append(remainingWeights[:idx], remainingWeights[idx+1:]...)
	}
	return picked
}

func extractBoosterConfig(v any) map[string]any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	if s, ok := v.(string); ok {
		var m map[string]any
		if err := json.Unmarshal([]byte(s), &m); err == nil {
			return m
		}
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}
