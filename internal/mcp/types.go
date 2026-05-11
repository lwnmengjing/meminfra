package mcp

import "encoding/json"

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcErrorObject `json:"error,omitempty"`
}

type rpcErrorObject struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func rpcResult(id json.RawMessage, result any) rpcResponse {
	return rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

func rpcError(id json.RawMessage, code int, message string, data any) rpcResponse {
	return rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &rpcErrorObject{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

type initializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    serverCapabilities `json:"capabilities"`
	ServerInfo      implementation     `json:"serverInfo"`
	Instructions    string             `json:"instructions,omitempty"`
}

type serverCapabilities struct {
	Tools toolsCapability `json:"tools"`
}

type toolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

type implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type toolDefinition struct {
	Name        string         `json:"name"`
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type callToolRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

type toolResult struct {
	Content           []contentBlock `json:"content"`
	StructuredContent map[string]any `json:"structuredContent,omitempty"`
	IsError           bool           `json:"isError"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type searchMemoryArgs struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

type listArgs struct {
	Limit int `json:"limit,omitempty"`
}

type getResourceArgs struct {
	Key string `json:"key"`
}

type queryTopologyArgs struct {
	Resource  string `json:"resource"`
	Type      string `json:"type,omitempty"`
	Direction string `json:"direction,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type listObservationsArgs struct {
	Resource string `json:"resource,omitempty"`
	Metric   string `json:"metric,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

type listEventsArgs struct {
	Resource string `json:"resource,omitempty"`
	Type     string `json:"type,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}
