package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/lwnmengjing/ai-infra-operator/internal/store"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}

	switch args[0] {
	case "init":
		return runInit(ctx, args[1:], stdout)
	case "resource":
		return runResource(ctx, args[1:], stdout)
	case "observe":
		return runObserve(ctx, args[1:], stdout)
	case "event":
		return runEvent(ctx, args[1:], stdout)
	case "search":
		return runSearch(ctx, args[1:], stdout)
	case "help", "-h", "--help":
		printUsage(stdout)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runInit(ctx context.Context, args []string, stdout io.Writer) error {
	fs := newFlagSet("init", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
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

	fmt.Fprintf(stdout, "initialized %s\n", *dbPath)
	return nil
}

func runResource(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "upsert" {
		return errors.New("usage: meminfra resource upsert --db PATH --key KEY --kind KIND [flags]")
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

	resource, err := mem.UpsertResource(ctx, store.ResourceInput{
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

	fmt.Fprintf(stdout, "resource %s id=%d\n", resource.ResourceKey, resource.ID)
	return nil
}

func runObserve(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "add" {
		return errors.New("usage: meminfra observe add --db PATH --resource KEY --metric METRIC --value VALUE [flags]")
	}

	fs := newFlagSet("observe add", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	resourceKey := fs.String("resource", "", "resource key")
	metric := fs.String("metric", "", "metric")
	value := fs.String("value", "", "numeric value")
	unit := fs.String("unit", "", "unit")
	source := fs.String("source", "manual", "source")
	metadata := fs.String("metadata", "", "metadata JSON")
	if err := fs.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	parsedValue, err := strconv.ParseFloat(*value, 64)
	if err != nil {
		return fmt.Errorf("value must be numeric: %w", err)
	}

	mem, err := openAndMigrate(ctx, *dbPath)
	if err != nil {
		return err
	}
	defer mem.Close()

	observation, err := mem.AddObservation(ctx, store.ObservationInput{
		ResourceKey:  *resourceKey,
		Metric:       *metric,
		Value:        parsedValue,
		Unit:         *unit,
		Source:       *source,
		MetadataJSON: *metadata,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "observation id=%d resource=%s metric=%s value=%g\n", observation.ID, *resourceKey, observation.Metric, observation.Value)
	return nil
}

func runEvent(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "add" {
		return errors.New("usage: meminfra event add --db PATH --resource KEY --type TYPE [flags]")
	}

	fs := newFlagSet("event add", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	resourceKey := fs.String("resource", "", "resource key")
	eventType := fs.String("type", "", "event type")
	data := fs.String("data", "", "event data JSON")
	source := fs.String("source", "manual", "source")
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

	event, err := mem.AddEvent(ctx, store.EventInput{
		ResourceKey:   *resourceKey,
		EventType:     *eventType,
		EventDataJSON: *data,
		Source:        *source,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "event id=%d resource=%s type=%s\n", event.ID, *resourceKey, event.EventType)
	return nil
}

func runSearch(ctx context.Context, args []string, stdout io.Writer) error {
	fs := newFlagSet("search", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	limit := fs.Int("limit", 10, "result limit")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	query := strings.Join(fs.Args(), " ")
	if strings.TrimSpace(query) == "" {
		return errors.New("search query is required")
	}

	mem, err := openAndMigrate(ctx, *dbPath)
	if err != nil {
		return err
	}
	defer mem.Close()

	results, err := mem.Search(ctx, query, *limit)
	if err != nil {
		return err
	}
	for _, result := range results {
		fmt.Fprintf(stdout, "[%s:%d] %s\n%s\n\n", result.DocType, result.RefID, result.Title, result.Body)
	}
	return nil
}

func openAndMigrate(ctx context.Context, dbPath string) (*store.Store, error) {
	mem, err := store.Open(dbPath)
	if err != nil {
		return nil, err
	}
	if err := mem.Migrate(ctx); err != nil {
		mem.Close()
		return nil, err
	}
	return mem, nil
}

func newFlagSet(name string, output io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(output)
	return fs
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `meminfra - AI-native infrastructure memory

Commands:
  init --db PATH
  resource upsert --db PATH --key KEY --kind KIND [--hostname NAME] [--ipv4 IP] [--ipv6 IP] [--provider NAME] [--region NAME]
  observe add --db PATH --resource KEY --metric METRIC --value VALUE [--unit UNIT]
  event add --db PATH --resource KEY --type TYPE [--data JSON]
  search --db PATH [--limit N] QUERY`)
}
