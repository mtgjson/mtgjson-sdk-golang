package db

import (
	"strings"
	"testing"
)

func TestBasicSelect(t *testing.T) {
	q := NewSQLBuilder("cards")
	sql, params := q.Build()
	if sql != "SELECT *\nFROM cards" {
		t.Errorf("unexpected SQL: %s", sql)
	}
	if len(params) != 0 {
		t.Errorf("expected no params, got %v", params)
	}
}

func TestWhereEq(t *testing.T) {
	q := NewSQLBuilder("cards").WhereEq("name", "Bolt")
	sql, params := q.Build()
	if !strings.Contains(sql, "WHERE name = $1") {
		t.Errorf("expected WHERE name = $1, got: %s", sql)
	}
	if len(params) != 1 || params[0] != "Bolt" {
		t.Errorf("unexpected params: %v", params)
	}
}

func TestWhereGteLte(t *testing.T) {
	q := NewSQLBuilder("cards").WhereGTE("manaValue", 2.0).WhereLTE("manaValue", 5.0)
	sql, params := q.Build()
	if !strings.Contains(sql, "manaValue >= $1") {
		t.Errorf("expected manaValue >= $1, got: %s", sql)
	}
	if !strings.Contains(sql, "manaValue <= $2") {
		t.Errorf("expected manaValue <= $2, got: %s", sql)
	}
	if len(params) != 2 || params[0] != 2.0 || params[1] != 5.0 {
		t.Errorf("unexpected params: %v", params)
	}
}

func TestWhereOr(t *testing.T) {
	q := NewSQLBuilder("cards").WhereOr(
		WhereOrCondition{"name = $1", "Lightning Bolt"},
		WhereOrCondition{"name = $1", "Counterspell"},
	)
	sql, params := q.Build()
	if !strings.Contains(sql, "(name = $1 OR name = $2)") {
		t.Errorf("expected (name = $1 OR name = $2), got: %s", sql)
	}
	if len(params) != 2 || params[0] != "Lightning Bolt" || params[1] != "Counterspell" {
		t.Errorf("unexpected params: %v", params)
	}
}

func TestWhereOrCombinedWithAnd(t *testing.T) {
	q := NewSQLBuilder("cards").
		WhereEq("setCode", "A25").
		WhereOr(
			WhereOrCondition{"rarity = $1", "rare"},
			WhereOrCondition{"rarity = $1", "mythic"},
		)
	sql, params := q.Build()
	if !strings.Contains(sql, "setCode = $1") {
		t.Errorf("expected setCode = $1, got: %s", sql)
	}
	if !strings.Contains(sql, "(rarity = $2 OR rarity = $3)") {
		t.Errorf("expected (rarity = $2 OR rarity = $3), got: %s", sql)
	}
	if len(params) != 3 || params[0] != "A25" || params[1] != "rare" || params[2] != "mythic" {
		t.Errorf("unexpected params: %v", params)
	}
}

func TestGroupBy(t *testing.T) {
	q := NewSQLBuilder("cards").Select("setCode", "COUNT(*)").GroupBy("setCode")
	sql, _ := q.Build()
	if !strings.Contains(sql, "GROUP BY setCode") {
		t.Errorf("expected GROUP BY setCode, got: %s", sql)
	}
}

func TestHaving(t *testing.T) {
	q := NewSQLBuilder("cards").
		Select("setCode", "COUNT(*) AS cnt").
		GroupBy("setCode").
		Having("COUNT(*) > $1", 10)
	sql, params := q.Build()
	if !strings.Contains(sql, "HAVING COUNT(*) > $1") {
		t.Errorf("expected HAVING COUNT(*) > $1, got: %s", sql)
	}
	if len(params) != 1 || params[0] != 10 {
		t.Errorf("unexpected params: %v", params)
	}
}

func TestDistinct(t *testing.T) {
	q := NewSQLBuilder("cards").Select("name").Distinct()
	sql, _ := q.Build()
	if !strings.HasPrefix(sql, "SELECT DISTINCT name") {
		t.Errorf("expected SELECT DISTINCT name prefix, got: %s", sql)
	}
}

func TestWhereRegex(t *testing.T) {
	q := NewSQLBuilder("cards").WhereRegex("text", `deals \d+ damage`)
	sql, params := q.Build()
	if !strings.Contains(sql, "regexp_matches(text, $1)") {
		t.Errorf("expected regexp_matches(text, $1), got: %s", sql)
	}
	if len(params) != 1 || params[0] != `deals \d+ damage` {
		t.Errorf("unexpected params: %v", params)
	}
}

func TestWhereRegexWithOtherConditions(t *testing.T) {
	q := NewSQLBuilder("cards").WhereEq("setCode", "A25").WhereRegex("text", "^Draw")
	sql, params := q.Build()
	if !strings.Contains(sql, "setCode = $1") {
		t.Errorf("expected setCode = $1, got: %s", sql)
	}
	if !strings.Contains(sql, "regexp_matches(text, $2)") {
		t.Errorf("expected regexp_matches(text, $2), got: %s", sql)
	}
	if len(params) != 2 || params[0] != "A25" || params[1] != "^Draw" {
		t.Errorf("unexpected params: %v", params)
	}
}

func TestWhereFuzzy(t *testing.T) {
	q := NewSQLBuilder("cards").WhereFuzzy("name", "Ligtning Bolt", 0.8)
	sql, params := q.Build()
	if !strings.Contains(sql, "jaro_winkler_similarity(name, $1) > 0.8") {
		t.Errorf("expected jaro_winkler_similarity(name, $1) > 0.8, got: %s", sql)
	}
	if len(params) != 1 || params[0] != "Ligtning Bolt" {
		t.Errorf("unexpected params: %v", params)
	}
}

func TestWhereFuzzyCustomThreshold(t *testing.T) {
	q := NewSQLBuilder("cards").WhereFuzzy("name", "Bolt", 0.9)
	sql, params := q.Build()
	if !strings.Contains(sql, "jaro_winkler_similarity(name, $1) > 0.9") {
		t.Errorf("expected jaro_winkler_similarity(name, $1) > 0.9, got: %s", sql)
	}
	if len(params) != 1 || params[0] != "Bolt" {
		t.Errorf("unexpected params: %v", params)
	}
}

func TestWhereFuzzyWithOtherConditions(t *testing.T) {
	q := NewSQLBuilder("cards").
		WhereEq("setCode", "A25").
		WhereFuzzy("name", "Ligtning Bolt", 0.8)
	sql, params := q.Build()
	if !strings.Contains(sql, "setCode = $1") {
		t.Errorf("expected setCode = $1, got: %s", sql)
	}
	if !strings.Contains(sql, "jaro_winkler_similarity(name, $2) > 0.8") {
		t.Errorf("expected jaro_winkler_similarity(name, $2) > 0.8, got: %s", sql)
	}
	if len(params) != 2 || params[0] != "A25" || params[1] != "Ligtning Bolt" {
		t.Errorf("unexpected params: %v", params)
	}
}

func TestFullQuery(t *testing.T) {
	q := NewSQLBuilder("prices_today").
		Select("provider", "AVG(price) AS avg_price").
		WhereEq("uuid", "abc-123").
		WhereGTE("date", "2024-01-01").
		GroupBy("provider").
		Having("AVG(price) > $1", 1.0).
		OrderBy("avg_price DESC").
		Limit(10)
	sql, params := q.Build()
	if !strings.Contains(sql, "SELECT provider, AVG(price) AS avg_price") {
		t.Errorf("expected SELECT clause, got: %s", sql)
	}
	if !strings.Contains(sql, "WHERE uuid = $1 AND date >= $2") {
		t.Errorf("expected WHERE clause, got: %s", sql)
	}
	if !strings.Contains(sql, "GROUP BY provider") {
		t.Errorf("expected GROUP BY, got: %s", sql)
	}
	if !strings.Contains(sql, "HAVING AVG(price) > $3") {
		t.Errorf("expected HAVING, got: %s", sql)
	}
	if !strings.Contains(sql, "ORDER BY avg_price DESC") {
		t.Errorf("expected ORDER BY, got: %s", sql)
	}
	if !strings.Contains(sql, "LIMIT 10") {
		t.Errorf("expected LIMIT 10, got: %s", sql)
	}
	if len(params) != 3 || params[0] != "abc-123" || params[1] != "2024-01-01" || params[2] != 1.0 {
		t.Errorf("unexpected params: %v", params)
	}
}

func TestLimitRejectsNegative(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for negative limit")
		}
	}()
	NewSQLBuilder("t").Limit(-1)
}

func TestLimitAcceptsZero(t *testing.T) {
	q := NewSQLBuilder("t").Limit(0)
	sql, _ := q.Build()
	if !strings.Contains(sql, "LIMIT 0") {
		t.Errorf("expected LIMIT 0, got: %s", sql)
	}
}

func TestOffsetRejectsNegative(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for negative offset")
		}
	}()
	NewSQLBuilder("t").Offset(-5)
}

func TestFuzzyThresholdRejectsOutOfRange(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for out-of-range threshold")
		}
	}()
	NewSQLBuilder("t").WhereFuzzy("name", "Bolt", 2.0)
}

func TestWhereIn(t *testing.T) {
	q := NewSQLBuilder("cards").WhereIn("uuid", []any{"abc", "def", "ghi"})
	sql, params := q.Build()
	if !strings.Contains(sql, "uuid IN ($1, $2, $3)") {
		t.Errorf("expected uuid IN ($1, $2, $3), got: %s", sql)
	}
	if len(params) != 3 {
		t.Errorf("expected 3 params, got %d", len(params))
	}
}

func TestWhereInEmpty(t *testing.T) {
	q := NewSQLBuilder("cards").WhereIn("uuid", []any{})
	sql, params := q.Build()
	if !strings.Contains(sql, "FALSE") {
		t.Errorf("expected FALSE for empty IN, got: %s", sql)
	}
	if len(params) != 0 {
		t.Errorf("expected 0 params, got %d", len(params))
	}
}

func TestWhereLike(t *testing.T) {
	q := NewSQLBuilder("cards").WhereLike("name", "Lightning%")
	sql, params := q.Build()
	if !strings.Contains(sql, "LOWER(name) LIKE LOWER($1)") {
		t.Errorf("expected LOWER(name) LIKE LOWER($1), got: %s", sql)
	}
	if len(params) != 1 || params[0] != "Lightning%" {
		t.Errorf("unexpected params: %v", params)
	}
}

func TestJoin(t *testing.T) {
	q := NewSQLBuilder("cards").
		Join("JOIN sets s ON cards.setCode = s.code").
		WhereEq("s.type", "expansion")
	sql, params := q.Build()
	if !strings.Contains(sql, "JOIN sets s ON cards.setCode = s.code") {
		t.Errorf("expected JOIN clause, got: %s", sql)
	}
	if !strings.Contains(sql, "s.type = $1") {
		t.Errorf("expected s.type = $1, got: %s", sql)
	}
	if len(params) != 1 || params[0] != "expansion" {
		t.Errorf("unexpected params: %v", params)
	}
}

func TestOffsetOnly(t *testing.T) {
	q := NewSQLBuilder("t").Offset(10)
	sql, _ := q.Build()
	if !strings.Contains(sql, "OFFSET 10") {
		t.Errorf("expected OFFSET 10, got: %s", sql)
	}
}

func TestLimitAndOffset(t *testing.T) {
	q := NewSQLBuilder("t").Limit(10).Offset(20)
	sql, _ := q.Build()
	if !strings.Contains(sql, "LIMIT 10") {
		t.Errorf("expected LIMIT 10, got: %s", sql)
	}
	if !strings.Contains(sql, "OFFSET 20") {
		t.Errorf("expected OFFSET 20, got: %s", sql)
	}
}
