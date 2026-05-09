package main

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strconv"
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

	stdout.Reset()
	if err := run(ctx, []string{
		"incident", "add",
		"--db", dbPath,
		"--title", "Frankfurt RTT spike",
		"--symptoms", "RTT increased",
		"--root-cause", "OVH upstream congestion",
		"--solution", "Shift traffic to London",
		"--result", "Latency recovered",
		"--tags", "frankfurt rtt ovh",
		"--metadata", `{"region":"fra"}`,
		"--output", "json",
	}, &stdout, &stderr); err != nil {
		t.Fatalf("incident add: %v", err)
	}
	var incidentPayload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &incidentPayload); err != nil {
		t.Fatalf("incident json: %v\n%s", err, stdout.String())
	}
	if incidentPayload["title"] != "Frankfurt RTT spike" || incidentPayload["root_cause"] != "OVH upstream congestion" {
		t.Fatalf("unexpected incident json: %#v", incidentPayload)
	}
	incidentMetadata, ok := incidentPayload["metadata_json"].(map[string]any)
	if !ok || incidentMetadata["region"] != "fra" {
		t.Fatalf("incident metadata_json should be embedded JSON, got %#v", incidentPayload["metadata_json"])
	}
}

func TestParentCommandHelpIsVisible(t *testing.T) {
	tests := [][]string{
		{"resource", "-h"},
		{"observe", "-h"},
		{"event", "-h"},
		{"incident", "-h"},
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

func TestGetAndListCommandsReturnJSON(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "meminfra.db")
	ctx := context.Background()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := run(ctx, []string{"init", "--db", dbPath}, &stdout, &stderr); err != nil {
		t.Fatalf("init: %v", err)
	}
	stdout.Reset()
	if err := run(ctx, []string{
		"resource", "upsert",
		"--db", dbPath,
		"--key", "node/frankfurt-01",
		"--kind", "server",
		"--hostname", "frankfurt-01",
	}, &stdout, &stderr); err != nil {
		t.Fatalf("resource upsert: %v", err)
	}

	stdout.Reset()
	if err := run(ctx, []string{"resource", "get", "--db", dbPath, "--key", "node/frankfurt-01", "--output", "json"}, &stdout, &stderr); err != nil {
		t.Fatalf("resource get: %v", err)
	}
	var resource map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &resource); err != nil {
		t.Fatalf("resource get json: %v\n%s", err, stdout.String())
	}
	if resource["resource_key"] != "node/frankfurt-01" {
		t.Fatalf("unexpected resource: %#v", resource)
	}

	stdout.Reset()
	if err := run(ctx, []string{"resource", "list", "--db", dbPath, "--output", "json"}, &stdout, &stderr); err != nil {
		t.Fatalf("resource list: %v", err)
	}
	var resources []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &resources); err != nil {
		t.Fatalf("resource list json: %v\n%s", err, stdout.String())
	}
	if len(resources) != 1 {
		t.Fatalf("unexpected resource list: %#v", resources)
	}

	stdout.Reset()
	if err := run(ctx, []string{
		"observe", "add",
		"--db", dbPath,
		"--resource", "node/frankfurt-01",
		"--metric", "rtt_ms",
		"--value", "82",
		"--output", "json",
	}, &stdout, &stderr); err != nil {
		t.Fatalf("observe add: %v", err)
	}
	observationID := jsonID(t, stdout.Bytes())

	stdout.Reset()
	if err := run(ctx, []string{"observe", "get", "--db", dbPath, "--id", observationID, "--output", "json"}, &stdout, &stderr); err != nil {
		t.Fatalf("observe get: %v", err)
	}
	var observation map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &observation); err != nil {
		t.Fatalf("observe get json: %v\n%s", err, stdout.String())
	}
	if observation["metric"] != "rtt_ms" {
		t.Fatalf("unexpected observation: %#v", observation)
	}

	stdout.Reset()
	if err := run(ctx, []string{"observe", "list", "--db", dbPath, "--resource", "node/frankfurt-01", "--metric", "rtt_ms", "--output", "json"}, &stdout, &stderr); err != nil {
		t.Fatalf("observe list: %v", err)
	}
	var observations []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &observations); err != nil {
		t.Fatalf("observe list json: %v\n%s", err, stdout.String())
	}
	if len(observations) != 1 {
		t.Fatalf("unexpected observation list: %#v", observations)
	}

	stdout.Reset()
	if err := run(ctx, []string{
		"event", "add",
		"--db", dbPath,
		"--resource", "node/frankfurt-01",
		"--type", "rtt_spike",
		"--output", "json",
	}, &stdout, &stderr); err != nil {
		t.Fatalf("event add: %v", err)
	}
	eventID := jsonID(t, stdout.Bytes())

	stdout.Reset()
	if err := run(ctx, []string{"event", "get", "--db", dbPath, "--id", eventID, "--output", "json"}, &stdout, &stderr); err != nil {
		t.Fatalf("event get: %v", err)
	}
	var event map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &event); err != nil {
		t.Fatalf("event get json: %v\n%s", err, stdout.String())
	}
	if event["event_type"] != "rtt_spike" {
		t.Fatalf("unexpected event: %#v", event)
	}

	stdout.Reset()
	if err := run(ctx, []string{"event", "list", "--db", dbPath, "--resource", "node/frankfurt-01", "--type", "rtt_spike", "--output", "json"}, &stdout, &stderr); err != nil {
		t.Fatalf("event list: %v", err)
	}
	var events []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &events); err != nil {
		t.Fatalf("event list json: %v\n%s", err, stdout.String())
	}
	if len(events) != 1 {
		t.Fatalf("unexpected event list: %#v", events)
	}

	stdout.Reset()
	if err := run(ctx, []string{"incident", "add", "--db", dbPath, "--title", "Frankfurt RTT spike", "--output", "json"}, &stdout, &stderr); err != nil {
		t.Fatalf("incident add: %v", err)
	}
	incidentID := jsonID(t, stdout.Bytes())

	stdout.Reset()
	if err := run(ctx, []string{"incident", "get", "--db", dbPath, "--id", incidentID, "--output", "json"}, &stdout, &stderr); err != nil {
		t.Fatalf("incident get: %v", err)
	}
	var incident map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &incident); err != nil {
		t.Fatalf("incident get json: %v\n%s", err, stdout.String())
	}
	if incident["title"] != "Frankfurt RTT spike" {
		t.Fatalf("unexpected incident: %#v", incident)
	}

	stdout.Reset()
	if err := run(ctx, []string{"incident", "list", "--db", dbPath, "--output", "json"}, &stdout, &stderr); err != nil {
		t.Fatalf("incident list: %v", err)
	}
	var incidents []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &incidents); err != nil {
		t.Fatalf("incident list json: %v\n%s", err, stdout.String())
	}
	if len(incidents) != 1 {
		t.Fatalf("unexpected incident list: %#v", incidents)
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

func jsonID(t *testing.T, data []byte) string {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("json id payload: %v\n%s", err, string(data))
	}
	id, ok := payload["id"].(float64)
	if !ok || id <= 0 {
		t.Fatalf("missing json id: %#v", payload)
	}
	return strconv.FormatUint(uint64(id), 10)
}

func TestUnsupportedOutputFormatFails(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run(context.Background(), []string{"init", "--db", filepath.Join(t.TempDir(), "meminfra.db"), "--output", "yaml"}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), `unsupported output format "yaml"`) {
		t.Fatalf("expected unsupported output error, got %v", err)
	}
}
