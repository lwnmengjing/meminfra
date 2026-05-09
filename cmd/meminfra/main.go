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
	case "incident":
		return runIncident(ctx, args[1:], stdout)
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

	resources, err := mem.ListResources(ctx, store.ListOptions{Limit: *limit})
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

func runObserve(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Fprintln(stdout, "Usage: meminfra observe <add|get|list> [flags]")
		return nil
	}
	if args[0] == "get" {
		return runObserveGet(ctx, args[1:], stdout)
	}
	if args[0] == "list" {
		return runObserveList(ctx, args[1:], stdout)
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

func runObserveGet(ctx context.Context, args []string, stdout io.Writer) error {
	fs := newFlagSet("observe get", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	id := fs.String("id", "", "observation id")
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

	observation, err := mem.ObservationByID(ctx, parsedID)
	if err != nil {
		return err
	}
	if *output == "json" {
		return writeJSON(stdout, newObservationOutput(observation))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	fmt.Fprintf(stdout, "observation id=%d resource_id=%d metric=%s value=%g\n", observation.ID, observation.ResourceID, observation.Metric, observation.Value)
	return nil
}

func runObserveList(ctx context.Context, args []string, stdout io.Writer) error {
	fs := newFlagSet("observe list", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	resourceKey := fs.String("resource", "", "resource key")
	metric := fs.String("metric", "", "metric")
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

	observations, err := mem.ListObservations(ctx, store.ObservationListOptions{
		ResourceKey: *resourceKey,
		Metric:      *metric,
		Limit:       *limit,
	})
	if err != nil {
		return err
	}
	if *output == "json" {
		return writeJSON(stdout, newObservationOutputs(observations))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	for _, observation := range observations {
		fmt.Fprintf(stdout, "observation id=%d resource_id=%d metric=%s value=%g\n", observation.ID, observation.ResourceID, observation.Metric, observation.Value)
	}
	return nil
}

func runEvent(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Fprintln(stdout, "Usage: meminfra event <add|get|list> [flags]")
		return nil
	}
	if args[0] == "get" {
		return runEventGet(ctx, args[1:], stdout)
	}
	if args[0] == "list" {
		return runEventList(ctx, args[1:], stdout)
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

func runEventGet(ctx context.Context, args []string, stdout io.Writer) error {
	fs := newFlagSet("event get", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	id := fs.String("id", "", "event id")
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

	event, err := mem.EventByID(ctx, parsedID)
	if err != nil {
		return err
	}
	if *output == "json" {
		return writeJSON(stdout, newEventOutput(event))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	fmt.Fprintf(stdout, "event id=%d resource_id=%d type=%s\n", event.ID, event.ResourceID, event.EventType)
	return nil
}

func runEventList(ctx context.Context, args []string, stdout io.Writer) error {
	fs := newFlagSet("event list", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	resourceKey := fs.String("resource", "", "resource key")
	eventType := fs.String("type", "", "event type")
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

	events, err := mem.ListEvents(ctx, store.EventListOptions{
		ResourceKey: *resourceKey,
		EventType:   *eventType,
		Limit:       *limit,
	})
	if err != nil {
		return err
	}
	if *output == "json" {
		return writeJSON(stdout, newEventOutputs(events))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	for _, event := range events {
		fmt.Fprintf(stdout, "event id=%d resource_id=%d type=%s\n", event.ID, event.ResourceID, event.EventType)
	}
	return nil
}

func runIncident(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Fprintln(stdout, "Usage: meminfra incident <add|get|list> [flags]")
		return nil
	}
	if args[0] == "get" {
		return runIncidentGet(ctx, args[1:], stdout)
	}
	if args[0] == "list" {
		return runIncidentList(ctx, args[1:], stdout)
	}
	if args[0] != "add" {
		return errors.New("usage: meminfra incident add --db PATH --title TITLE [flags]")
	}

	fs := newFlagSet("incident add", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	title := fs.String("title", "", "incident title")
	symptoms := fs.String("symptoms", "", "symptoms")
	rootCause := fs.String("root-cause", "", "root cause")
	solution := fs.String("solution", "", "solution")
	result := fs.String("result", "", "result")
	tags := fs.String("tags", "", "space or comma separated tags")
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

	incident, err := mem.AddIncident(ctx, store.IncidentInput{
		Title:        *title,
		Symptoms:     *symptoms,
		RootCause:    *rootCause,
		Solution:     *solution,
		Result:       *result,
		Tags:         *tags,
		Source:       *source,
		MetadataJSON: *metadata,
	})
	if err != nil {
		return err
	}

	if *output == "json" {
		return writeJSON(stdout, newIncidentOutput(incident))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	fmt.Fprintf(stdout, "incident id=%d title=%s\n", incident.ID, incident.Title)
	return nil
}

func runIncidentGet(ctx context.Context, args []string, stdout io.Writer) error {
	fs := newFlagSet("incident get", stdout)
	dbPath := fs.String("db", "meminfra.db", "SQLite database path")
	id := fs.String("id", "", "incident id")
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

	incident, err := mem.IncidentByID(ctx, parsedID)
	if err != nil {
		return err
	}
	if *output == "json" {
		return writeJSON(stdout, newIncidentOutput(incident))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	fmt.Fprintf(stdout, "incident id=%d title=%s\n", incident.ID, incident.Title)
	return nil
}

func runIncidentList(ctx context.Context, args []string, stdout io.Writer) error {
	fs := newFlagSet("incident list", stdout)
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

	incidents, err := mem.ListIncidents(ctx, store.IncidentListOptions{Limit: *limit})
	if err != nil {
		return err
	}
	if *output == "json" {
		return writeJSON(stdout, newIncidentOutputs(incidents))
	}
	if *output != "text" {
		return fmt.Errorf("unsupported output format %q", *output)
	}
	for _, incident := range incidents {
		fmt.Fprintf(stdout, "incident id=%d title=%s\n", incident.ID, incident.Title)
	}
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

func parseUintFlag(value string, name string) (uint, error) {
	if strings.TrimSpace(value) == "" {
		return 0, fmt.Errorf("%s is required", name)
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return uint(parsed), nil
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

type incidentOutput struct {
	ID           uint            `json:"id"`
	Title        string          `json:"title"`
	Symptoms     string          `json:"symptoms"`
	RootCause    string          `json:"root_cause"`
	Solution     string          `json:"solution"`
	Result       string          `json:"result"`
	Tags         string          `json:"tags"`
	Source       string          `json:"source"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
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

func newResourceOutputs(resources []model.Resource) []resourceOutput {
	outputs := make([]resourceOutput, 0, len(resources))
	for i := range resources {
		outputs = append(outputs, newResourceOutput(&resources[i]))
	}
	return outputs
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

func newObservationOutputs(observations []model.Observation) []observationOutput {
	outputs := make([]observationOutput, 0, len(observations))
	for i := range observations {
		outputs = append(outputs, newObservationOutput(&observations[i]))
	}
	return outputs
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

func newEventOutputs(events []model.Event) []eventOutput {
	outputs := make([]eventOutput, 0, len(events))
	for i := range events {
		outputs = append(outputs, newEventOutput(&events[i]))
	}
	return outputs
}

func newIncidentOutput(incident *model.Incident) incidentOutput {
	return incidentOutput{
		ID:           incident.ID,
		Title:        incident.Title,
		Symptoms:     incident.Symptoms,
		RootCause:    incident.RootCause,
		Solution:     incident.Solution,
		Result:       incident.Result,
		Tags:         incident.Tags,
		Source:       incident.Source,
		MetadataJSON: rawJSON(incident.MetadataJSON),
		CreatedAt:    incident.CreatedAt,
		UpdatedAt:    incident.UpdatedAt,
	}
}

func newIncidentOutputs(incidents []model.Incident) []incidentOutput {
	outputs := make([]incidentOutput, 0, len(incidents))
	for i := range incidents {
		outputs = append(outputs, newIncidentOutput(&incidents[i]))
	}
	return outputs
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
  resource upsert|get|list
  observe add|get|list
  event add|get|list
  incident add|get|list
  search --db PATH [--limit N] QUERY

All commands support --output text|json.`)
}
