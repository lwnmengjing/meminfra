package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

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
