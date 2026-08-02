package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/mss-boot-io/meminfra/internal/legacy/core"
)

func runResource(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Fprintln(stdout, "Usage: meminfra resource <upsert|get|list> [flags]")
		return nil
	}
	if args[0] == "get" {
		return runResourceGet(ctx, args[1:], stdout)
	}
	if args[0] == "list" {
		return runResourceList(ctx, args[1:], stdout)
	}
	if args[0] != "upsert" {
		return errors.New("usage: meminfra resource <upsert|get|list> [flags]")
	}

	fs := newFlagSet("resource upsert", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	key := fs.String("key", "", "resource key")
	kind := fs.String("kind", "", "resource kind")
	hostname := fs.String("hostname", "", "hostname")
	ipv4 := fs.String("ipv4", "", "IPv4 address")
	ipv6 := fs.String("ipv6", "", "IPv6 address")
	provider := fs.String("provider", "", "provider")
	region := fs.String("region", "", "region")
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

	resource, err := mem.UpsertResource(ctx, core.ResourceInput{
		ResourceKey:  *key,
		Kind:         *kind,
		Hostname:     *hostname,
		IPv4:         *ipv4,
		IPv6:         *ipv6,
		Provider:     *provider,
		Region:       *region,
		Source:       *source,
		MetadataJSON: *metadata,
	})
	if err != nil {
		return err
	}

	if *output == "json" {
		return writeJSON(stdout, newResourceOutput(resource))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	fmt.Fprintf(stdout, "resource %s id=%d\n", resource.ResourceKey, resource.ID)
	return nil
}

func runResourceGet(ctx context.Context, args []string, stdout io.Writer) error {
	fs := newFlagSet("resource get", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	key := fs.String("key", "", "resource key")
	output := fs.String("output", "text", "output format: text or json")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(*key) == "" {
		return errors.New("--key is required")
	}

	mem, err := openAndMigrate(ctx, *dbPath)
	if err != nil {
		return err
	}
	defer mem.Close()

	resource, err := mem.ResourceByKey(ctx, *key)
	if err != nil {
		return err
	}
	if *output == "json" {
		return writeJSON(stdout, newResourceOutput(resource))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	fmt.Fprintf(stdout, "resource %s id=%d kind=%s hostname=%s\n", resource.ResourceKey, resource.ID, resource.Kind, resource.Hostname)
	return nil
}

func runResourceList(ctx context.Context, args []string, stdout io.Writer) error {
	fs := newFlagSet("resource list", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
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

	resources, err := mem.ListResources(ctx, core.ListOptions{Limit: *limit})
	if err != nil {
		return err
	}
	if *output == "json" {
		return writeJSON(stdout, newResourceOutputs(resources))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	for _, resource := range resources {
		fmt.Fprintf(stdout, "resource %s id=%d kind=%s hostname=%s\n", resource.ResourceKey, resource.ID, resource.Kind, resource.Hostname)
	}
	return nil
}
