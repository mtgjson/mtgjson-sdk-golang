package mtgjson

import (
	"context"
	"testing"
)

func testConnection(t *testing.T) *Connection {
	t.Helper()
	cache, err := newCacheManager(defaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	conn, err := NewConnection(cache)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestConnectionBasicQuery(t *testing.T) {
	conn := testConnection(t)
	ctx := context.Background()

	rows, err := conn.Execute(ctx, "SELECT 42 AS num, 'hello' AS msg")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	// DuckDB returns int32 for integer literals
	num, ok := rows[0]["num"].(int32)
	if !ok {
		t.Fatalf("expected int32, got %T: %v", rows[0]["num"], rows[0]["num"])
	}
	if num != 42 {
		t.Fatalf("expected 42, got %d", num)
	}
	msg, ok := rows[0]["msg"].(string)
	if !ok {
		t.Fatalf("expected string, got %T", rows[0]["msg"])
	}
	if msg != "hello" {
		t.Fatalf("expected 'hello', got %q", msg)
	}
}

func TestConnectionExecuteScalar(t *testing.T) {
	conn := testConnection(t)
	ctx := context.Background()

	val, err := conn.ExecuteScalar(ctx, "SELECT 42")
	if err != nil {
		t.Fatal(err)
	}
	num, ok := val.(int32)
	if !ok {
		t.Fatalf("expected int32, got %T: %v", val, val)
	}
	if num != 42 {
		t.Fatalf("expected 42, got %d", num)
	}
}

func TestConnectionExecuteJSON(t *testing.T) {
	conn := testConnection(t)
	ctx := context.Background()

	jsonStr, err := conn.ExecuteJSON(ctx, "SELECT 1 AS a, 'x' AS b UNION ALL SELECT 2, 'y'")
	if err != nil {
		t.Fatal(err)
	}
	if jsonStr == "[]" {
		t.Fatal("expected non-empty JSON")
	}
	// Should be a valid JSON array
	if jsonStr[0] != '[' || jsonStr[len(jsonStr)-1] != ']' {
		t.Fatalf("expected JSON array, got: %s", jsonStr)
	}
}

func TestConnectionRegisterTableFromData(t *testing.T) {
	conn := testConnection(t)
	ctx := context.Background()

	data := []map[string]any{
		{"name": "Alpha", "value": 1},
		{"name": "Beta", "value": 2},
		{"name": "Gamma", "value": 3},
	}
	if err := conn.RegisterTableFromData(ctx, "test_items", data); err != nil {
		t.Fatal(err)
	}

	rows, err := conn.Execute(ctx, "SELECT * FROM test_items ORDER BY name")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	if rows[0]["name"] != "Alpha" {
		t.Fatalf("expected 'Alpha', got %v", rows[0]["name"])
	}
}

func TestConnectionExecuteInto(t *testing.T) {
	conn := testConnection(t)
	ctx := context.Background()

	data := []map[string]any{
		{"name": "Alpha", "value": 1},
		{"name": "Beta", "value": 2},
	}
	if err := conn.RegisterTableFromData(ctx, "test_into", data); err != nil {
		t.Fatal(err)
	}

	type Item struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	var items []Item
	if err := conn.ExecuteInto(ctx, &items, "SELECT * FROM test_into ORDER BY name"); err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Name != "Alpha" {
		t.Fatalf("expected 'Alpha', got %q", items[0].Name)
	}
	if items[1].Value != 2 {
		t.Fatalf("expected 2, got %d", items[1].Value)
	}
}

func TestConnectionParameterizedQuery(t *testing.T) {
	conn := testConnection(t)
	ctx := context.Background()

	data := []map[string]any{
		{"name": "Alpha", "category": "A"},
		{"name": "Beta", "category": "B"},
		{"name": "Gamma", "category": "A"},
	}
	if err := conn.RegisterTableFromData(ctx, "test_params", data); err != nil {
		t.Fatal(err)
	}

	// Test with $1 style parameters (DuckDB native)
	rows, err := conn.Execute(ctx, "SELECT * FROM test_params WHERE category = $1", "A")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}

func TestConnectionViews(t *testing.T) {
	conn := testConnection(t)
	ctx := context.Background()

	data := []map[string]any{{"x": 1}}
	if err := conn.RegisterTableFromData(ctx, "test_view_table", data); err != nil {
		t.Fatal(err)
	}

	views := conn.Views()
	found := false
	for _, v := range views {
		if v == "test_view_table" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'test_view_table' in views, got %v", views)
	}
	if !conn.HasView("test_view_table") {
		t.Fatal("expected HasView to return true")
	}
}
