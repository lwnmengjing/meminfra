package main

import (
	"context"
	"fmt"
	"io"
	"os"
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
	case "relationship":
		return runRelationship(ctx, args[1:], stdout)
	case "search":
		return runSearch(ctx, args[1:], stdout)
	case "help", "-h", "--help":
		printUsage(stdout)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `meminfra - AI-native infrastructure memory

Commands:
  init --db PATH
  resource upsert|get|list
  observe add|get|list
  event add|get|list
  incident add|get|list
  relationship add|get|list|topology
  search --db PATH [--limit N] QUERY

All commands support --output text|json.`)
}
