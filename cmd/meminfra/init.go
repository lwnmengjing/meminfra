package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
)

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
