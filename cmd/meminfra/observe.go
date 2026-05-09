package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/lwnmengjing/ai-infra-operator/internal/core"
)

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
		return errors.New("usage: meminfra observe <add|get|list> [flags]")
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

	observation, err := mem.AddObservation(ctx, core.ObservationInput{
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

	observations, err := mem.ListObservations(ctx, core.ObservationListOptions{
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
