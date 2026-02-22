package mtgjson

import (
	"context"
	"testing"
)

func TestTokenGetByUUID(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	token, err := sdk.Tokens().GetByUUID(ctx, "token-uuid-001")
	if err != nil {
		t.Fatal(err)
	}
	if token == nil {
		t.Fatal("expected token, got nil")
	}
	if token.Name != "Soldier Token" {
		t.Fatalf("expected Soldier Token, got %s", token.Name)
	}
}

func TestTokenGetByUUIDNotFound(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	token, err := sdk.Tokens().GetByUUID(ctx, "nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if token != nil {
		t.Fatalf("expected nil, got %v", token)
	}
}

func TestTokenGetByName(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	tokens, err := sdk.Tokens().GetByName(ctx, "Soldier Token")
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1, got %d", len(tokens))
	}
	if tokens[0].Name != "Soldier Token" {
		t.Fatalf("expected Soldier Token, got %s", tokens[0].Name)
	}
}

func TestTokenSearchByName(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	tokens, err := sdk.Tokens().Search(ctx, SearchTokensParams{Name: "%Token"})
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected 2, got %d", len(tokens))
	}
}

func TestTokenSearchBySet(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	tokens, err := sdk.Tokens().Search(ctx, SearchTokensParams{SetCode: "A25"})
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1, got %d", len(tokens))
	}
	if tokens[0].SetCode != "A25" {
		t.Fatalf("expected A25, got %s", tokens[0].SetCode)
	}
}

func TestTokenSearchByColors(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	tokens, err := sdk.Tokens().Search(ctx, SearchTokensParams{Colors: []string{"G"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1, got %d", len(tokens))
	}
	if tokens[0].Name != "Beast Token" {
		t.Fatalf("expected Beast Token, got %s", tokens[0].Name)
	}
}

func TestTokenForSet(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	tokens, err := sdk.Tokens().ForSet(ctx, "MH2")
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1, got %d", len(tokens))
	}
	if tokens[0].Name != "Beast Token" {
		t.Fatalf("expected Beast Token, got %s", tokens[0].Name)
	}
}

func TestTokenCount(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	count, err := sdk.Tokens().Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

func TestTokenCountFiltered(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	count, err := sdk.Tokens().Count(ctx, Filter{Column: "setCode", Value: "A25"})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1, got %d", count)
	}
}

func TestTokenGetByUUIDs(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	tokens, err := sdk.Tokens().GetByUUIDs(ctx, []string{"token-uuid-001", "token-uuid-002"})
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected 2, got %d", len(tokens))
	}
	names := make(map[string]bool)
	for _, tok := range tokens {
		names[tok.Name] = true
	}
	if !names["Soldier Token"] || !names["Beast Token"] {
		t.Fatalf("expected both tokens, got %v", names)
	}
}

func TestTokenGetByUUIDsEmpty(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	tokens, err := sdk.Tokens().GetByUUIDs(ctx, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 0 {
		t.Fatalf("expected 0, got %d", len(tokens))
	}
}
