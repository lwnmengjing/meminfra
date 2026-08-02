package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/mss-boot-io/meminfra/internal/legacy/core"
)

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
		return errors.New("usage: meminfra incident <add|get|list> [flags]")
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

	incident, err := mem.AddIncident(ctx, core.IncidentInput{
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

	incidents, err := mem.ListIncidents(ctx, core.IncidentListOptions{Limit: *limit})
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
