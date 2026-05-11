package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/mss-boot-io/meminfra/internal/core"
)

func TestServerListsAndCallsTools(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "meminfra.db")
	seedMeminfra(t, ctx, dbPath)

	server, err := NewServer(ctx, dbPath)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	defer server.Close()

	listResponse := callHandle(t, server, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	})
	tools := listResponse["result"].(map[string]any)["tools"].([]any)
	if len(tools) == 0 {
		t.Fatalf("expected tools")
	}

	searchResponse := callHandle(t, server, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "search_memory",
			"arguments": map[string]any{
				"query": "frankfurt wg_tunnel",
			},
		},
	})
	searchResult := searchResponse["result"].(map[string]any)
	if searchResult["isError"] != false {
		t.Fatalf("unexpected tool error: %#v", searchResult)
	}
	results := searchResult["structuredContent"].(map[string]any)["results"].([]any)
	if len(results) == 0 {
		t.Fatalf("expected search results: %#v", searchResult)
	}

	topologyResponse := callHandle(t, server, map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "query_topology",
			"arguments": map[string]any{
				"resource":  "node/frankfurt-01",
				"direction": "out",
			},
		},
	})
	edges := topologyResponse["result"].(map[string]any)["structuredContent"].(map[string]any)["edges"].([]any)
	if len(edges) != 1 {
		t.Fatalf("unexpected topology edges: %#v", edges)
	}
	edge := edges[0].(map[string]any)
	src := edge["src_resource"].(map[string]any)
	if src["resource_key"] != "node/frankfurt-01" {
		t.Fatalf("unexpected src resource: %#v", edge)
	}
}

func TestServerUsesContentLengthFraming(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "meminfra.db")
	seedMeminfra(t, ctx, dbPath)

	server, err := NewServer(ctx, dbPath)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	defer server.Close()

	request := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	var input bytes.Buffer
	if err := writeMessage(&input, request); err != nil {
		t.Fatalf("write request frame: %v", err)
	}

	var output bytes.Buffer
	if err := server.Serve(&input, &output); err != nil {
		t.Fatalf("serve: %v", err)
	}

	payload, err := readMessage(bufio.NewReader(bytes.NewBuffer(output.Bytes())))
	if err != nil {
		t.Fatalf("read response frame: %v", err)
	}
	var response map[string]any
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("response json: %v\n%s", err, string(payload))
	}
	if response["id"].(float64) != 1 {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func callHandle(t *testing.T, server *Server, request map[string]any) map[string]any {
	t.Helper()

	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	responsePayload, ok := server.handle(context.Background(), payload)
	if !ok {
		t.Fatalf("expected response")
	}
	var response map[string]any
	if err := json.Unmarshal(responsePayload, &response); err != nil {
		t.Fatalf("response json: %v\n%s", err, string(responsePayload))
	}
	if response["error"] != nil {
		t.Fatalf("unexpected error response: %#v", response)
	}
	return response
}

func seedMeminfra(t *testing.T, ctx context.Context, dbPath string) {
	t.Helper()

	service, err := core.Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("open seed service: %v", err)
	}
	defer service.Close()

	if _, err := service.UpsertResource(ctx, core.ResourceInput{
		ResourceKey:  "node/frankfurt-01",
		Kind:         "server",
		Hostname:     "frankfurt-01",
		MetadataJSON: `{"role":"edge"}`,
	}); err != nil {
		t.Fatalf("seed src resource: %v", err)
	}
	if _, err := service.UpsertResource(ctx, core.ResourceInput{
		ResourceKey: "node/london-01",
		Kind:        "server",
		Hostname:    "london-01",
	}); err != nil {
		t.Fatalf("seed dst resource: %v", err)
	}
	if _, err := service.AddRelationship(ctx, core.RelationshipInput{
		SrcResourceKey: "node/frankfurt-01",
		DstResourceKey: "node/london-01",
		RelationType:   "wg_tunnel",
		MetadataJSON:   `{"interface":"wg0"}`,
	}); err != nil {
		t.Fatalf("seed relationship: %v", err)
	}
}
