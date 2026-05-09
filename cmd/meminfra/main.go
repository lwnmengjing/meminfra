package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lwnmengjing/ai-infra-operator/internal/model"
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

	if *output == "json" {
		return writeJSON(stdout, map[string]any{
			"status": "initialized",
			"db":     *dbPath,
		})
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	fmt.Fprintf(stdout, "initialized %s\n", *dbPath)
	return nil
}

func runResource(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Fprintln(stdout, "Usage: meminfra resource upsert --db PATH --key KEY --kind KIND [flags]")
		return nil
	}
	if args[0] != "upsert" {
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

	if *output == "json" {
		return writeJSON(stdout, newResourceOutput(resource))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	fmt.Fprintf(stdout, "resource %s id=%d\n", resource.ResourceKey, resource.ID)
	return nil
}

func runObserve(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Fprintln(stdout, "Usage: meminfra observe add --db PATH --resource KEY --metric METRIC --value VALUE [flags]")
		return nil
	}
	if args[0] != "add" {
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
	output := fs.String("output", "text", "output format: text or json")
	if err := fs.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(*value) == "" {
		return errors.New("--value is required")
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

	if *output == "json" {
		return writeJSON(stdout, newObservationOutput(observation))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	fmt.Fprintf(stdout, "observation id=%d resource=%s metric=%s value=%g\n", observation.ID, *resourceKey, observation.Metric, observation.Value)
	return nil
}

func runEvent(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Fprintln(stdout, "Usage: meminfra event add --db PATH --resource KEY --type TYPE [flags]")
		return nil
	}
	if args[0] != "add" {
		return errors.New("usage: meminfra event add --db PATH --resource KEY --type TYPE [flags]")
	}

	fs := newFlagSet("event add", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	resourceKey := fs.String("resource", "", "resource key")
	eventType := fs.String("type", "", "event type")
	data := fs.String("data", "", "event data JSON")
	source := fs.String("source", "manual", "source")
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

	event, err := mem.AddEvent(ctx, store.EventInput{
		ResourceKey:   *resourceKey,
		EventType:     *eventType,
		EventDataJSON: *data,
		Source:        *source,
	})
	if err != nil {
		return err
	}

	if *output == "json" {
		return writeJSON(stdout, newEventOutput(event))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	fmt.Fprintf(stdout, "event id=%d resource=%s type=%s\n", event.ID, *resourceKey, event.EventType)
	return nil
}

func runSearch(ctx context.Context, args []string, stdout io.Writer) error {
	fs := newFlagSet("search", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	limit := fs.Int("limit", 10, "result limit")
	output := fs.String("output", "text", "output format: text or json")
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
	if *output == "json" {
		return writeJSON(stdout, results)
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
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

func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "help"
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

type resourceOutput struct {
	ID           uint            `json:"id"`
	ResourceKey  string          `json:"resource_key"`
	Kind         string          `json:"kind"`
	Hostname     string          `json:"hostname"`
	IPv4         string          `json:"ipv4"`
	IPv6         string          `json:"ipv6"`
	Provider     string          `json:"provider"`
	Region       string          `json:"region"`
	Source       string          `json:"source"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
	FirstSeen    time.Time       `json:"first_seen"`
	LastSeen     time.Time       `json:"last_seen"`
}

type observationOutput struct {
	ID           uint            `json:"id"`
	ResourceID   uint            `json:"resource_id"`
	Metric       string          `json:"metric"`
	Value        float64         `json:"value"`
	Unit         string          `json:"unit"`
	Source       string          `json:"source"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
	ObservedAt   time.Time       `json:"observed_at"`
}

type eventOutput struct {
	ID            uint            `json:"id"`
	ResourceID    uint            `json:"resource_id"`
	EventType     string          `json:"event_type"`
	EventDataJSON json.RawMessage `json:"event_data_json"`
	Source        string          `json:"source"`
	CreatedAt     time.Time       `json:"created_at"`
}

func newResourceOutput(resource *model.Resource) resourceOutput {
	return resourceOutput{
		ID:           resource.ID,
		ResourceKey:  resource.ResourceKey,
		Kind:         resource.Kind,
		Hostname:     resource.Hostname,
		IPv4:         resource.IPv4,
		IPv6:         resource.IPv6,
		Provider:     resource.Provider,
		Region:       resource.Region,
		Source:       resource.Source,
		MetadataJSON: rawJSON(resource.MetadataJSON),
		FirstSeen:    resource.FirstSeen,
		LastSeen:     resource.LastSeen,
	}
}

func newObservationOutput(observation *model.Observation) observationOutput {
	return observationOutput{
		ID:           observation.ID,
		ResourceID:   observation.ResourceID,
		Metric:       observation.Metric,
		Value:        observation.Value,
		Unit:         observation.Unit,
		Source:       observation.Source,
		MetadataJSON: rawJSON(observation.MetadataJSON),
		ObservedAt:   observation.ObservedAt,
	}
}

func newEventOutput(event *model.Event) eventOutput {
	return eventOutput{
		ID:            event.ID,
		ResourceID:    event.ResourceID,
		EventType:     event.EventType,
		EventDataJSON: rawJSON(event.EventDataJSON),
		Source:        event.Source,
		CreatedAt:     event.CreatedAt,
	}
}

func rawJSON(value string) json.RawMessage {
	if strings.TrimSpace(value) == "" {
		return json.RawMessage("null")
	}
	return json.RawMessage(value)
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `meminfra - AI-native infrastructure memory

Commands:
  init --db PATH
  resource upsert --db PATH --key KEY --kind KIND [--hostname NAME] [--ipv4 IP] [--ipv6 IP] [--provider NAME] [--region NAME]
  observe add --db PATH --resource KEY --metric METRIC --value VALUE [--unit UNIT]
  event add --db PATH --resource KEY --type TYPE [--data JSON]
  search --db PATH [--limit N] QUERY

All commands support --output text|json.`)
}
