package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/lwnmengjing/ai-infra-operator/internal/core"
)

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
		return errors.New("usage: meminfra event <add|get|list> [flags]")
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

	event, err := mem.AddEvent(ctx, core.EventInput{
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

	events, err := mem.ListEvents(ctx, core.EventListOptions{
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
