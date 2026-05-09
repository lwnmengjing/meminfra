package core

import (
	"context"
	"path/filepath"
	"testing"
)

func TestServiceOpenWriteAndSearch(t *testing.T) {
	ctx := context.Background()
	service, err := Open(ctx, filepath.Join(t.TempDir(), "meminfra.db"))
	if err != nil {
		t.Fatalf("open service: %v", err)
	}
	t.Cleanup(func() {
		if err := service.Close(); err != nil {
			t.Fatalf("close service: %v", err)
		}
	})

	if _, err := service.UpsertResource(ctx, ResourceInput{
		ResourceKey: "node/frankfurt-01",
		Kind:        "server",
		Hostname:    "frankfurt-01",
	}); err != nil {
		t.Fatalf("upsert resource: %v", err)
	}

	results, err := service.Search(ctx, "frankfurt", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 || results[0].DocType != "resource" {
		t.Fatalf("unexpected search results: %#v", results)
	}
}
