package mcp

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/mss-boot-io/meminfra/internal/model"
)

type resourceOutput struct {
	ID           uint            `json:"id"`
	ResourceKey  string          `json:"resource_key"`
	Kind         string          `json:"kind"`
	Hostname     string          `json:"hostname"`
	IPv4         string          `json:"ipv4"`
	IPv6         string          `json:"ipv6"`
	Provider     string          `json:"provider"`
	Region       string          `json:"region"`
	Source       string          `json:"source"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
	FirstSeen    time.Time       `json:"first_seen"`
	LastSeen     time.Time       `json:"last_seen"`
}

type observationOutput struct {
	ID           uint            `json:"id"`
	ResourceID   uint            `json:"resource_id"`
	Metric       string          `json:"metric"`
	Value        float64         `json:"value"`
	Unit         string          `json:"unit"`
	Source       string          `json:"source"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
	ObservedAt   time.Time       `json:"observed_at"`
}

type eventOutput struct {
	ID            uint            `json:"id"`
	ResourceID    uint            `json:"resource_id"`
	EventType     string          `json:"event_type"`
	EventDataJSON json.RawMessage `json:"event_data_json"`
	Source        string          `json:"source"`
	CreatedAt     time.Time       `json:"created_at"`
}

type incidentOutput struct {
	ID           uint            `json:"id"`
	Title        string          `json:"title"`
	Symptoms     string          `json:"symptoms"`
	RootCause    string          `json:"root_cause"`
	Solution     string          `json:"solution"`
	Result       string          `json:"result"`
	Tags         string          `json:"tags"`
	Source       string          `json:"source"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type relationshipOutput struct {
	ID            uint            `json:"id"`
	SrcResourceID uint            `json:"src_resource_id"`
	DstResourceID uint            `json:"dst_resource_id"`
	RelationType  string          `json:"relation_type"`
	Source        string          `json:"source"`
	MetadataJSON  json.RawMessage `json:"metadata_json"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type topologyEdgeOutput struct {
	Relationship relationshipOutput `json:"relationship"`
	SrcResource  resourceOutput     `json:"src_resource"`
	DstResource  resourceOutput     `json:"dst_resource"`
}

type searchResultOutput struct {
	ID      uint    `json:"id"`
	DocType string  `json:"doc_type"`
	RefID   uint    `json:"ref_id"`
	Title   string  `json:"title"`
	Body    string  `json:"body"`
	Tags    string  `json:"tags"`
	Rank    float64 `json:"rank"`
}

func resourceView(resource *model.Resource) resourceOutput {
	return resourceOutput{
		ID:           resource.ID,
		ResourceKey:  resource.ResourceKey,
		Kind:         resource.Kind,
		Hostname:     resource.Hostname,
		IPv4:         resource.IPv4,
		IPv6:         resource.IPv6,
		Provider:     resource.Provider,
		Region:       resource.Region,
		Source:       resource.Source,
		MetadataJSON: rawJSON(resource.MetadataJSON),
		FirstSeen:    resource.FirstSeen,
		LastSeen:     resource.LastSeen,
	}
}

func resourceViews(resources []model.Resource) []resourceOutput {
	outputs := make([]resourceOutput, 0, len(resources))
	for i := range resources {
		outputs = append(outputs, resourceView(&resources[i]))
	}
	return outputs
}

func observationView(observation *model.Observation) observationOutput {
	return observationOutput{
		ID:           observation.ID,
		ResourceID:   observation.ResourceID,
		Metric:       observation.Metric,
		Value:        observation.Value,
		Unit:         observation.Unit,
		Source:       observation.Source,
		MetadataJSON: rawJSON(observation.MetadataJSON),
		ObservedAt:   observation.ObservedAt,
	}
}

func observationViews(observations []model.Observation) []observationOutput {
	outputs := make([]observationOutput, 0, len(observations))
	for i := range observations {
		outputs = append(outputs, observationView(&observations[i]))
	}
	return outputs
}

func eventView(event *model.Event) eventOutput {
	return eventOutput{
		ID:            event.ID,
		ResourceID:    event.ResourceID,
		EventType:     event.EventType,
		EventDataJSON: rawJSON(event.EventDataJSON),
		Source:        event.Source,
		CreatedAt:     event.CreatedAt,
	}
}

func eventViews(events []model.Event) []eventOutput {
	outputs := make([]eventOutput, 0, len(events))
	for i := range events {
		outputs = append(outputs, eventView(&events[i]))
	}
	return outputs
}

func incidentView(incident *model.Incident) incidentOutput {
	return incidentOutput{
		ID:           incident.ID,
		Title:        incident.Title,
		Symptoms:     incident.Symptoms,
		RootCause:    incident.RootCause,
		Solution:     incident.Solution,
		Result:       incident.Result,
		Tags:         incident.Tags,
		Source:       incident.Source,
		MetadataJSON: rawJSON(incident.MetadataJSON),
		CreatedAt:    incident.CreatedAt,
		UpdatedAt:    incident.UpdatedAt,
	}
}

func incidentViews(incidents []model.Incident) []incidentOutput {
	outputs := make([]incidentOutput, 0, len(incidents))
	for i := range incidents {
		outputs = append(outputs, incidentView(&incidents[i]))
	}
	return outputs
}

func relationshipView(relationship *model.Relationship) relationshipOutput {
	return relationshipOutput{
		ID:            relationship.ID,
		SrcResourceID: relationship.SrcResourceID,
		DstResourceID: relationship.DstResourceID,
		RelationType:  relationship.RelationType,
		Source:        relationship.Source,
		MetadataJSON:  rawJSON(relationship.MetadataJSON),
		CreatedAt:     relationship.CreatedAt,
		UpdatedAt:     relationship.UpdatedAt,
	}
}

func topologyEdgeViews(edges []model.TopologyEdge) []topologyEdgeOutput {
	outputs := make([]topologyEdgeOutput, 0, len(edges))
	for _, edge := range edges {
		outputs = append(outputs, topologyEdgeOutput{
			Relationship: relationshipView(&edge.Relationship),
			SrcResource:  resourceView(&edge.SrcResource),
			DstResource:  resourceView(&edge.DstResource),
		})
	}
	return outputs
}

func searchResultViews(results []model.SearchResult) []searchResultOutput {
	outputs := make([]searchResultOutput, 0, len(results))
	for _, result := range results {
		outputs = append(outputs, searchResultOutput{
			ID:      result.ID,
			DocType: result.DocType,
			RefID:   result.RefID,
			Title:   result.Title,
			Body:    result.Body,
			Tags:    result.Tags,
			Rank:    result.Rank,
		})
	}
	return outputs
}

func rawJSON(value string) json.RawMessage {
	if strings.TrimSpace(value) == "" {
		return json.RawMessage("null")
	}
	return json.RawMessage(value)
}
