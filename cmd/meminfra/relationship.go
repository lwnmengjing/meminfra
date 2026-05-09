package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/mss-boot-io/meminfra/internal/core"
)

func runRelationship(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Fprintln(stdout, "Usage: meminfra relationship <add|get|list|topology> [flags]")
		return nil
	}
	if args[0] == "get" {
		return runRelationshipGet(ctx, args[1:], stdout)
	}
	if args[0] == "list" {
		return runRelationshipList(ctx, args[1:], stdout)
	}
	if args[0] == "topology" {
		return runRelationshipTopology(ctx, args[1:], stdout)
	}
	if args[0] != "add" {
		return errors.New("usage: meminfra relationship <add|get|list|topology> [flags]")
	}

	fs := newFlagSet("relationship add", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	src := fs.String("src", "", "source resource key")
	dst := fs.String("dst", "", "destination resource key")
	relationType := fs.String("type", "", "relationship type")
	source := fs.String("source", "manual", "source")
	metadata := fs.String("metadata", "", "metadata JSON")
	output := fs.String("output", "text", "output format: text or json")
	if err := fs.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	mem, err := openAndMigrate(ctx, *dbPath)
	if err != nil {
		return err
	}
	defer mem.Close()

	relationship, err := mem.AddRelationship(ctx, core.RelationshipInput{
		SrcResourceKey: *src,
		DstResourceKey: *dst,
		RelationType:   *relationType,
		Source:         *source,
		MetadataJSON:   *metadata,
	})
	if err != nil {
		return err
	}

	if *output == "json" {
		return writeJSON(stdout, newRelationshipOutput(relationship))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	fmt.Fprintf(stdout, "relationship id=%d src_resource_id=%d dst_resource_id=%d type=%s\n", relationship.ID, relationship.SrcResourceID, relationship.DstResourceID, relationship.RelationType)
	return nil
}

func runRelationshipGet(ctx context.Context, args []string, stdout io.Writer) error {
	fs := newFlagSet("relationship get", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	id := fs.String("id", "", "relationship id")
	output := fs.String("output", "text", "output format: text or json")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	parsedID, err := parseUintFlag(*id, "--id")
	if err != nil {
		return err
	}

	mem, err := openAndMigrate(ctx, *dbPath)
	if err != nil {
		return err
	}
	defer mem.Close()

	relationship, err := mem.RelationshipByID(ctx, parsedID)
	if err != nil {
		return err
	}
	if *output == "json" {
		return writeJSON(stdout, newRelationshipOutput(relationship))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	fmt.Fprintf(stdout, "relationship id=%d src_resource_id=%d dst_resource_id=%d type=%s\n", relationship.ID, relationship.SrcResourceID, relationship.DstResourceID, relationship.RelationType)
	return nil
}

func runRelationshipList(ctx context.Context, args []string, stdout io.Writer) error {
	fs := newFlagSet("relationship list", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	resourceKey := fs.String("resource", "", "resource key matched on either side")
	relationType := fs.String("type", "", "relationship type")
	limit := fs.Int("limit", 50, "result limit")
	output := fs.String("output", "text", "output format: text or json")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	mem, err := openAndMigrate(ctx, *dbPath)
	if err != nil {
		return err
	}
	defer mem.Close()

	relationships, err := mem.ListRelationships(ctx, core.RelationshipListOptions{
		ResourceKey:  *resourceKey,
		RelationType: *relationType,
		Limit:        *limit,
	})
	if err != nil {
		return err
	}
	if *output == "json" {
		return writeJSON(stdout, newRelationshipOutputs(relationships))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	for _, relationship := range relationships {
		fmt.Fprintf(stdout, "relationship id=%d src_resource_id=%d dst_resource_id=%d type=%s\n", relationship.ID, relationship.SrcResourceID, relationship.DstResourceID, relationship.RelationType)
	}
	return nil
}

func runRelationshipTopology(ctx context.Context, args []string, stdout io.Writer) error {
	fs := newFlagSet("relationship topology", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	resourceKey := fs.String("resource", "", "resource key")
	relationType := fs.String("type", "", "relationship type")
	direction := fs.String("direction", "both", "direction: both, in, or out")
	limit := fs.Int("limit", 50, "result limit")
	output := fs.String("output", "text", "output format: text or json")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	mem, err := openAndMigrate(ctx, *dbPath)
	if err != nil {
		return err
	}
	defer mem.Close()

	edges, err := mem.QueryTopology(ctx, core.TopologyQueryOptions{
		ResourceKey:  *resourceKey,
		RelationType: *relationType,
		Direction:    *direction,
		Limit:        *limit,
	})
	if err != nil {
		return err
	}
	if *output == "json" {
		return writeJSON(stdout, newTopologyEdgeOutputs(edges))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	for _, edge := range edges {
		fmt.Fprintf(stdout, "%s --%s--> %s relationship_id=%d\n", edge.SrcResource.ResourceKey, edge.Relationship.RelationType, edge.DstResource.ResourceKey, edge.Relationship.ID)
	}
	return nil
}
