package main

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestSubcommandHelpIsVisible(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run(context.Background(), []string{"resource", "upsert", "-h"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("help returned error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Usage of resource upsert") {
		t.Fatalf("expected subcommand usage, got %q", output)
	}
	if !strings.Contains(output, "-metadata") || !strings.Contains(output, "-key") {
		t.Fatalf("expected subcommand flags, got %q", output)
	}
}

func TestJSONOutputForWriteAndSearchCommands(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "meminfra.db")
	ctx := context.Background()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := run(ctx, []string{"init", "--db", dbPath, "--output", "json"}, &stdout, &stderr); err != nil {
		t.Fatalf("init: %v", err)
	}
	var initPayload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &initPayload); err != nil {
		t.Fatalf("init json: %v\n%s", err, stdout.String())
	}
	if initPayload["status"] != "initialized" || initPayload["db"] != dbPath {
		t.Fatalf("unexpected init json: %#v", initPayload)
	}

	stdout.Reset()
	if err := run(ctx, []string{
		"resource", "upsert",
		"--db", dbPath,
		"--key", "node/frankfurt-01",
		"--kind", "server",
		"--hostname", "frankfurt-01",
		"--metadata", `{"role":"edge"}`,
		"--output", "json",
	}, &stdout, &stderr); err != nil {
		t.Fatalf("resource upsert: %v", err)
	}
	var resourcePayload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &resourcePayload); err != nil {
		t.Fatalf("resource json: %v\n%s", err, stdout.String())
	}
	if resourcePayload["resource_key"] != "node/frankfurt-01" {
		t.Fatalf("unexpected resource json: %#v", resourcePayload)
	}
	metadata, ok := resourcePayload["metadata_json"].(map[string]any)
	if !ok || metadata["role"] != "edge" {
		t.Fatalf("metadata_json should be embedded JSON, got %#v", resourcePayload["metadata_json"])
	}

	stdout.Reset()
	if err := run(ctx, []string{
		"search",
		"--db", dbPath,
		"--output", "json",
		"frankfurt",
	}, &stdout, &stderr); err != nil {
		t.Fatalf("search: %v", err)
	}
	var results []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("search json: %v\n%s", err, stdout.String())
	}
	if len(results) != 1 || results[0]["DocType"] != "resource" {
		t.Fatalf("unexpected search json: %#v", results)
	}
}

func TestParentCommandHelpIsVisible(t *testing.T) {
	tests := [][]string{
		{"resource", "-h"},
		{"observe", "-h"},
		{"event", "-h"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			err := run(context.Background(), args, &stdout, &stderr)
			if err != nil {
				t.Fatalf("help returned error: %v", err)
			}
			if !strings.Contains(stdout.String(), "Usage: meminfra "+args[0]) {
				t.Fatalf("expected parent help, got %q", stdout.String())
			}
		})
	}
}

func TestObserveValueIsRequired(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run(context.Background(), []string{
		"observe", "add",
		"--db", filepath.Join(t.TempDir(), "meminfra.db"),
		"--resource", "node/frankfurt-01",
		"--metric", "rtt_ms",
	}, &stdout, &stderr)
	if err == nil || err.Error() != "--value is required" {
		t.Fatalf("expected required value error, got %v", err)
	}
}

func TestUnsupportedOutputFormatFails(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run(context.Background(), []string{"init", "--db", filepath.Join(t.TempDir(), "meminfra.db"), "--output", "yaml"}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), `unsupported output format "yaml"`) {
		t.Fatalf("expected unsupported output error, got %v", err)
	}
}
