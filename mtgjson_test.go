package mtgjson

import (
	"context"
	"testing"
)

func TestSDKNewAndClose(t *testing.T) {
	sdk, err := New(WithCacheDir(t.TempDir()), WithOffline(true))
	if err != nil {
		t.Fatal(err)
	}
	if err := sdk.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSDKViews(t *testing.T) {
	sdk := setupSampleSDK(t)
	views := sdk.Views()
	if len(views) == 0 {
		t.Fatal("expected some views registered")
	}
	found := false
	for _, v := range views {
		if v == "cards" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'cards' in views, got %v", views)
	}
}

func TestSDKSQL(t *testing.T) {
	sdk := setupSampleSDK(t)
	ctx := context.Background()

	rows, err := sdk.SQL(ctx, "SELECT name FROM cards ORDER BY name LIMIT 2")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2, got %d", len(rows))
	}
}

func TestSDKString(t *testing.T) {
	sdk := setupSampleSDK(t)
	s := sdk.String()
	if s == "" {
		t.Fatal("expected non-empty string")
	}
}
