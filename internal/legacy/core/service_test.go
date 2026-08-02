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
	if _, err := service.UpsertResource(ctx, ResourceInput{
		ResourceKey: "node/london-01",
		Kind:        "server",
		Hostname:    "london-01",
	}); err != nil {
		t.Fatalf("upsert second resource: %v", err)
	}
	if _, err := service.AddRelationship(ctx, RelationshipInput{
		SrcResourceKey: "node/frankfurt-01",
		DstResourceKey: "node/london-01",
		RelationType:   "wg_tunnel",
	}); err != nil {
		t.Fatalf("add relationship: %v", err)
	}

	results, err := service.Search(ctx, "frankfurt wg_tunnel", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 || results[0].DocType != "relationship" {
		t.Fatalf("unexpected search results: %#v", results)
	}
}
