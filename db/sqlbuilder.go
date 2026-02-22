package db

import (
	"fmt"
	"strings"
)

// SQLBuilder builds parameterized SQL queries safely.
// All user-supplied values go through DuckDB's parameter binding (?),
// never through string interpolation. Methods return the builder for chaining.
type SQLBuilder struct {
	selectCols []string
	isDistinct bool
	from       string
	joins      []string
	wheres     []string
	params     []any
	groupBys   []string
	havings    []string
	orderBys   []string
	limitVal   *int
	offsetVal  *int
}

// NewSQLBuilder creates a builder targeting the given table or view.
func NewSQLBuilder(table string) *SQLBuilder {
	return &SQLBuilder{
		selectCols: []string{"*"},
		from:       table,
	}
}

// Select sets the columns to select (replaces the default *).
func (b *SQLBuilder) Select(cols ...string) *SQLBuilder {
	b.selectCols = cols
	return b
}

// Distinct adds DISTINCT to the SELECT clause.
func (b *SQLBuilder) Distinct() *SQLBuilder {
	b.isDistinct = true
	return b
}

// Join adds a JOIN clause.
func (b *SQLBuilder) Join(clause string) *SQLBuilder {
	b.joins = append(b.joins, clause)
	return b
}

// Where adds a WHERE condition with positional params using $N placeholders.
// Placeholders are remapped automatically to the global parameter index.
func (b *SQLBuilder) Where(condition string, params ...any) *SQLBuilder {
	offset := len(b.params)
	remapped := condition
	for i := len(params); i >= 1; i-- {
		remapped = strings.ReplaceAll(remapped, fmt.Sprintf("$%d", i), fmt.Sprintf("$%d", offset+i))
	}
	b.wheres = append(b.wheres, remapped)
	b.params = append(b.params, params...)
	return b
}

// WhereLike adds a case-insensitive LIKE condition.
func (b *SQLBuilder) WhereLike(column, value string) *SQLBuilder {
	idx := len(b.params) + 1
	b.wheres = append(b.wheres, fmt.Sprintf("LOWER(%s) LIKE LOWER($%d)", column, idx))
	b.params = append(b.params, value)
	return b
}

// WhereIn adds an IN condition with parameterized values.
// An empty values slice produces FALSE.
func (b *SQLBuilder) WhereIn(column string, values []any) *SQLBuilder {
	if len(values) == 0 {
		b.wheres = append(b.wheres, "FALSE")
		return b
	}
	placeholders := make([]string, len(values))
	for i, v := range values {
		idx := len(b.params) + 1
		placeholders[i] = fmt.Sprintf("$%d", idx)
		b.params = append(b.params, v)
	}
	b.wheres = append(b.wheres, fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", ")))
	return b
}

// WhereEq adds an equality condition.
func (b *SQLBuilder) WhereEq(column string, value any) *SQLBuilder {
	idx := len(b.params) + 1
	b.wheres = append(b.wheres, fmt.Sprintf("%s = $%d", column, idx))
	b.params = append(b.params, value)
	return b
}

// WhereGTE adds a greater-than-or-equal condition.
func (b *SQLBuilder) WhereGTE(column string, value any) *SQLBuilder {
	idx := len(b.params) + 1
	b.wheres = append(b.wheres, fmt.Sprintf("%s >= $%d", column, idx))
	b.params = append(b.params, value)
	return b
}

// WhereLTE adds a less-than-or-equal condition.
func (b *SQLBuilder) WhereLTE(column string, value any) *SQLBuilder {
	idx := len(b.params) + 1
	b.wheres = append(b.wheres, fmt.Sprintf("%s <= $%d", column, idx))
	b.params = append(b.params, value)
	return b
}

// WhereRegex adds a regex match condition (DuckDB regexp_matches).
func (b *SQLBuilder) WhereRegex(column, pattern string) *SQLBuilder {
	idx := len(b.params) + 1
	b.wheres = append(b.wheres, fmt.Sprintf("regexp_matches(%s, $%d)", column, idx))
	b.params = append(b.params, pattern)
	return b
}

// WhereFuzzy adds a fuzzy string match condition using Jaro-Winkler similarity.
// threshold must be between 0 and 1 (default 0.8).
func (b *SQLBuilder) WhereFuzzy(column, value string, threshold float64) *SQLBuilder {
	if threshold < 0 || threshold > 1 {
		panic(fmt.Sprintf("mtgjson: threshold must be between 0 and 1, got %v", threshold))
	}
	idx := len(b.params) + 1
	b.wheres = append(b.wheres, fmt.Sprintf("jaro_winkler_similarity(%s, $%d) > %g", column, idx, threshold))
	b.params = append(b.params, value)
	return b
}

// WhereOr adds OR-combined conditions. Each condition is a tuple of
// (sql_fragment, param_value) where the fragment uses $1 as a placeholder.
func (b *SQLBuilder) WhereOr(conditions ...WhereOrCondition) *SQLBuilder {
	if len(conditions) == 0 {
		return b
	}
	parts := make([]string, len(conditions))
	for i, cond := range conditions {
		idx := len(b.params) + 1
		remapped := strings.ReplaceAll(cond.SQL, "$1", fmt.Sprintf("$%d", idx))
		parts[i] = remapped
		b.params = append(b.params, cond.Value)
	}
	b.wheres = append(b.wheres, "("+strings.Join(parts, " OR ")+")")
	return b
}

// WhereOrCondition is a single condition for WhereOr.
type WhereOrCondition struct {
	SQL   string
	Value any
}

// GroupBy adds GROUP BY columns.
func (b *SQLBuilder) GroupBy(columns ...string) *SQLBuilder {
	b.groupBys = append(b.groupBys, columns...)
	return b
}

// Having adds a HAVING condition (works like Where but for aggregates).
func (b *SQLBuilder) Having(condition string, params ...any) *SQLBuilder {
	offset := len(b.params)
	remapped := condition
	for i := len(params); i >= 1; i-- {
		remapped = strings.ReplaceAll(remapped, fmt.Sprintf("$%d", i), fmt.Sprintf("$%d", offset+i))
	}
	b.havings = append(b.havings, remapped)
	b.params = append(b.params, params...)
	return b
}

// OrderBy adds ORDER BY clauses.
func (b *SQLBuilder) OrderBy(clauses ...string) *SQLBuilder {
	b.orderBys = append(b.orderBys, clauses...)
	return b
}

// Limit sets the maximum number of rows to return.
// Panics if n is negative.
func (b *SQLBuilder) Limit(n int) *SQLBuilder {
	if n < 0 {
		panic(fmt.Sprintf("mtgjson: limit must be non-negative, got %d", n))
	}
	b.limitVal = &n
	return b
}

// Offset sets the number of rows to skip before returning results.
// Panics if n is negative.
func (b *SQLBuilder) Offset(n int) *SQLBuilder {
	if n < 0 {
		panic(fmt.Sprintf("mtgjson: offset must be non-negative, got %d", n))
	}
	b.offsetVal = &n
	return b
}

// Build returns the final SQL string and parameter list.
// The SQL uses $1, $2, ... placeholders matching the parameter positions.
func (b *SQLBuilder) Build() (string, []any) {
	var parts []string

	distinct := ""
	if b.isDistinct {
		distinct = "DISTINCT "
	}
	parts = append(parts, fmt.Sprintf("SELECT %s%s", distinct, strings.Join(b.selectCols, ", ")))
	parts = append(parts, "FROM "+b.from)

	for _, j := range b.joins {
		parts = append(parts, j)
	}

	if len(b.wheres) > 0 {
		parts = append(parts, "WHERE "+strings.Join(b.wheres, " AND "))
	}

	if len(b.groupBys) > 0 {
		parts = append(parts, "GROUP BY "+strings.Join(b.groupBys, ", "))
	}

	if len(b.havings) > 0 {
		parts = append(parts, "HAVING "+strings.Join(b.havings, " AND "))
	}

	if len(b.orderBys) > 0 {
		parts = append(parts, "ORDER BY "+strings.Join(b.orderBys, ", "))
	}

	if b.limitVal != nil {
		parts = append(parts, fmt.Sprintf("LIMIT %d", *b.limitVal))
	}

	if b.offsetVal != nil {
		parts = append(parts, fmt.Sprintf("OFFSET %d", *b.offsetVal))
	}

	return strings.Join(parts, "\n"), b.params
}

// AddWhere exposes direct addition of a WHERE clause for query modules.
func (b *SQLBuilder) AddWhere(cond string) {
	b.wheres = append(b.wheres, cond)
}

// AddParam adds a parameter and returns its 1-based index.
func (b *SQLBuilder) AddParam(val any) int {
	b.params = append(b.params, val)
	return len(b.params)
}

// Params returns the current parameter slice (for direct manipulation by query modules).
func (b *SQLBuilder) Params() []any {
	return b.params
}
