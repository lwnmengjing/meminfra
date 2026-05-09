package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/lwnmengjing/ai-infra-operator/internal/core"
)

const protocolVersion = "2025-06-18"

type Server struct {
	service *core.Service
}

func NewServer(ctx context.Context, dbPath string) (*Server, error) {
	service, err := core.Open(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	return &Server{service: service}, nil
}

func (s *Server) Close() error {
	return s.service.Close()
}

func (s *Server) Serve(r io.Reader, w io.Writer) error {
	reader := bufio.NewReader(r)
	for {
		payload, err := readMessage(reader)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		response, ok := s.handle(context.Background(), payload)
		if !ok {
			continue
		}
		if err := writeMessage(w, response); err != nil {
			return err
		}
	}
}

func (s *Server) handle(ctx context.Context, payload []byte) ([]byte, bool) {
	var request rpcRequest
	if err := json.Unmarshal(payload, &request); err != nil {
		return mustJSON(rpcError(nil, -32700, "parse error", err.Error())), true
	}
	if len(request.ID) == 0 {
		return nil, false
	}

	switch request.Method {
	case "initialize":
		return mustJSON(rpcResult(request.ID, initializeResult{
			ProtocolVersion: protocolVersion,
			Capabilities: serverCapabilities{
				Tools: toolsCapability{ListChanged: false},
			},
			ServerInfo: implementation{
				Name:    "meminfra",
				Version: "0.1.0",
			},
			Instructions: "Use MemInfra tools to query local infrastructure memory, observations, incidents, and topology.",
		})), true
	case "ping":
		return mustJSON(rpcResult(request.ID, map[string]any{})), true
	case "tools/list":
		return mustJSON(rpcResult(request.ID, map[string]any{"tools": toolDefinitions()})), true
	case "tools/call":
		result, err := s.callTool(ctx, request.Params)
		if err != nil {
			return mustJSON(rpcError(request.ID, -32602, err.Error(), nil)), true
		}
		return mustJSON(rpcResult(request.ID, result)), true
	default:
		return mustJSON(rpcError(request.ID, -32601, "method not found", request.Method)), true
	}
}

func (s *Server) callTool(ctx context.Context, params json.RawMessage) (toolResult, error) {
	var request callToolRequest
	if err := json.Unmarshal(params, &request); err != nil {
		return toolResult{}, fmt.Errorf("invalid tools/call params: %w", err)
	}
	if strings.TrimSpace(request.Name) == "" {
		return toolResult{}, fmt.Errorf("tool name is required")
	}

	switch request.Name {
	case "search_memory":
		var args searchMemoryArgs
		if err := decodeArgs(request.Arguments, &args); err != nil {
			return toolResult{}, err
		}
		results, err := s.service.Search(ctx, args.Query, args.Limit)
		if err != nil {
			return toolResult{}, err
		}
		return structuredToolResult(map[string]any{"results": searchResultViews(results)})
	case "list_resources":
		var args listArgs
		if err := decodeArgs(request.Arguments, &args); err != nil {
			return toolResult{}, err
		}
		resources, err := s.service.ListResources(ctx, core.ListOptions{Limit: args.Limit})
		if err != nil {
			return toolResult{}, err
		}
		return structuredToolResult(map[string]any{"resources": resourceViews(resources)})
	case "get_resource":
		var args getResourceArgs
		if err := decodeArgs(request.Arguments, &args); err != nil {
			return toolResult{}, err
		}
		resource, err := s.service.ResourceByKey(ctx, args.Key)
		if err != nil {
			return toolResult{}, err
		}
		return structuredToolResult(map[string]any{"resource": resourceView(resource)})
	case "query_topology":
		var args queryTopologyArgs
		if err := decodeArgs(request.Arguments, &args); err != nil {
			return toolResult{}, err
		}
		edges, err := s.service.QueryTopology(ctx, core.TopologyQueryOptions{
			ResourceKey:  args.Resource,
			RelationType: args.Type,
			Direction:    args.Direction,
			Limit:        args.Limit,
		})
		if err != nil {
			return toolResult{}, err
		}
		return structuredToolResult(map[string]any{"edges": topologyEdgeViews(edges)})
	case "list_observations":
		var args listObservationsArgs
		if err := decodeArgs(request.Arguments, &args); err != nil {
			return toolResult{}, err
		}
		observations, err := s.service.ListObservations(ctx, core.ObservationListOptions{
			ResourceKey: args.Resource,
			Metric:      args.Metric,
			Limit:       args.Limit,
		})
		if err != nil {
			return toolResult{}, err
		}
		return structuredToolResult(map[string]any{"observations": observationViews(observations)})
	case "list_events":
		var args listEventsArgs
		if err := decodeArgs(request.Arguments, &args); err != nil {
			return toolResult{}, err
		}
		events, err := s.service.ListEvents(ctx, core.EventListOptions{
			ResourceKey: args.Resource,
			EventType:   args.Type,
			Limit:       args.Limit,
		})
		if err != nil {
			return toolResult{}, err
		}
		return structuredToolResult(map[string]any{"events": eventViews(events)})
	case "list_incidents":
		var args listArgs
		if err := decodeArgs(request.Arguments, &args); err != nil {
			return toolResult{}, err
		}
		incidents, err := s.service.ListIncidents(ctx, core.IncidentListOptions{Limit: args.Limit})
		if err != nil {
			return toolResult{}, err
		}
		return structuredToolResult(map[string]any{"incidents": incidentViews(incidents)})
	default:
		return toolResult{}, fmt.Errorf("unknown tool %q", request.Name)
	}
}

func decodeArgs(raw json.RawMessage, target any) error {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		raw = []byte("{}")
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("invalid tool arguments: %w", err)
	}
	return nil
}

func structuredToolResult(data map[string]any) (toolResult, error) {
	text, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return toolResult{}, err
	}
	return toolResult{
		Content: []contentBlock{{
			Type: "text",
			Text: string(text),
		}},
		StructuredContent: data,
		IsError:           false,
	}, nil
}

func readMessage(r *bufio.Reader) ([]byte, error) {
	contentLength := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("invalid header line %q", line)
		}
		if strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			parsed, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil || parsed < 0 {
				return nil, fmt.Errorf("invalid Content-Length %q", value)
			}
			contentLength = parsed
		}
	}
	if contentLength < 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func writeMessage(w io.Writer, payload []byte) error {
	if _, err := fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(payload)); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}

func mustJSON(value any) []byte {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return payload
}
