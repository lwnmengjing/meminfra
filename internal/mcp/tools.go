package mcp

func toolDefinitions() []toolDefinition {
	return []toolDefinition{
		{
			Name:        "search_memory",
			Title:       "Search Memory",
			Description: "Search MemInfra memory documents with safe full-text search.",
			InputSchema: objectSchema(map[string]any{
				"query": stringSchema("Search query text."),
				"limit": integerSchema("Maximum number of results."),
			}, []string{"query"}),
		},
		{
			Name:        "list_resources",
			Title:       "List Resources",
			Description: "List recently seen infrastructure resources.",
			InputSchema: objectSchema(map[string]any{
				"limit": integerSchema("Maximum number of resources."),
			}, nil),
		},
		{
			Name:        "get_resource",
			Title:       "Get Resource",
			Description: "Get one infrastructure resource by resource key.",
			InputSchema: objectSchema(map[string]any{
				"key": stringSchema("Resource key, for example node/frankfurt-01."),
			}, []string{"key"}),
		},
		{
			Name:        "query_topology",
			Title:       "Query Topology",
			Description: "Query one-hop topology edges around a resource.",
			InputSchema: objectSchema(map[string]any{
				"resource":  stringSchema("Center resource key."),
				"type":      stringSchema("Optional relationship type filter."),
				"direction": enumSchema("Direction filter.", []string{"both", "in", "out", "incoming", "outgoing"}),
				"limit":     integerSchema("Maximum number of edges."),
			}, []string{"resource"}),
		},
		{
			Name:        "list_observations",
			Title:       "List Observations",
			Description: "List observations, optionally filtered by resource and metric.",
			InputSchema: objectSchema(map[string]any{
				"resource": stringSchema("Optional resource key."),
				"metric":   stringSchema("Optional metric name."),
				"limit":    integerSchema("Maximum number of observations."),
			}, nil),
		},
		{
			Name:        "list_events",
			Title:       "List Events",
			Description: "List events, optionally filtered by resource and event type.",
			InputSchema: objectSchema(map[string]any{
				"resource": stringSchema("Optional resource key."),
				"type":     stringSchema("Optional event type."),
				"limit":    integerSchema("Maximum number of events."),
			}, nil),
		},
		{
			Name:        "list_incidents",
			Title:       "List Incidents",
			Description: "List recent operational incident memories.",
			InputSchema: objectSchema(map[string]any{
				"limit": integerSchema("Maximum number of incidents."),
			}, nil),
		},
	}
}

func objectSchema(properties map[string]any, required []string) map[string]any {
	schema := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func stringSchema(description string) map[string]any {
	return map[string]any{
		"type":        "string",
		"description": description,
	}
}

func integerSchema(description string) map[string]any {
	return map[string]any{
		"type":        "integer",
		"description": description,
		"minimum":     1,
	}
}

func enumSchema(description string, values []string) map[string]any {
	return map[string]any{
		"type":        "string",
		"description": description,
		"enum":        values,
	}
}
