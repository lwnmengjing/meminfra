package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/mss-boot-io/meminfra/internal/mcp"
)

func main() {
	dbPath := flag.String("db", "meminfra.db", "SQLite database path")
	flag.Parse()

	server, err := mcp.NewServer(context.Background(), *dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer server.Close()

	if err := server.Serve(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
