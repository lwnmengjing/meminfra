package index

import (
	"fmt"
	"strings"

	"github.com/mss-boot-io/meminfra/internal/model"
)

func ResourceDocument(resource model.Resource) model.MemoryDocument {
	return model.MemoryDocument{
		DocType: "resource",
		RefID:   resource.ID,
		Title:   strings.TrimSpace(strings.Join([]string{resource.Kind, resource.ResourceKey, resource.Hostname}, " ")),
		Body: strings.TrimSpace(strings.Join([]string{
			resource.ResourceKey,
			resource.Kind,
			resource.Hostname,
			resource.IPv4,
			resource.IPv6,
			resource.Provider,
			resource.Region,
			resource.Source,
			resource.MetadataJSON,
		}, " ")),
		Tags: strings.TrimSpace(strings.Join([]string{resource.Kind, resource.Provider, resource.Region, resource.Source}, " ")),
	}
}

func ObservationDocument(observation model.Observation, resource model.Resource) model.MemoryDocument {
	value := fmt.Sprintf("%g", observation.Value)
	return model.MemoryDocument{
		DocType: "observation",
		RefID:   observation.ID,
		Title:   strings.TrimSpace(strings.Join([]string{resource.ResourceKey, observation.Metric, value, observation.Unit}, " ")),
		Body: strings.TrimSpace(strings.Join([]string{
			resource.ResourceKey,
			resource.Hostname,
			resource.Provider,
			resource.Region,
			observation.Metric,
			value,
			observation.Unit,
			observation.Source,
			observation.MetadataJSON,
		}, " ")),
		Tags: strings.TrimSpace(strings.Join([]string{"observation", observation.Metric, observation.Unit, observation.Source}, " ")),
	}
}

func EventDocument(event model.Event, resource model.Resource) model.MemoryDocument {
	return model.MemoryDocument{
		DocType: "event",
		RefID:   event.ID,
		Title:   strings.TrimSpace(strings.Join([]string{resource.ResourceKey, event.EventType}, " ")),
		Body: strings.TrimSpace(strings.Join([]string{
			resource.ResourceKey,
			resource.Hostname,
			resource.Provider,
			resource.Region,
			event.EventType,
			event.Source,
			event.EventDataJSON,
		}, " ")),
		Tags: strings.TrimSpace(strings.Join([]string{"event", event.EventType, event.Source}, " ")),
	}
}

func IncidentDocument(incident model.Incident) model.MemoryDocument {
	return model.MemoryDocument{
		DocType: "incident",
		RefID:   incident.ID,
		Title:   incident.Title,
		Body: strings.TrimSpace(strings.Join([]string{
			incident.Title,
			incident.Symptoms,
			incident.RootCause,
			incident.Solution,
			incident.Result,
			incident.Source,
			incident.MetadataJSON,
		}, " ")),
		Tags: strings.TrimSpace(strings.Join([]string{"incident", incident.Tags, incident.Source}, " ")),
	}
}

func RelationshipDocument(relationship model.Relationship, src model.Resource, dst model.Resource) model.MemoryDocument {
	return model.MemoryDocument{
		DocType: "relationship",
		RefID:   relationship.ID,
		Title: strings.TrimSpace(strings.Join([]string{
			src.ResourceKey,
			relationship.RelationType,
			dst.ResourceKey,
		}, " ")),
		Body: strings.TrimSpace(strings.Join([]string{
			src.ResourceKey,
			src.Hostname,
			dst.ResourceKey,
			dst.Hostname,
			relationship.RelationType,
			relationship.Source,
			relationship.MetadataJSON,
		}, " ")),
		Tags: strings.TrimSpace(strings.Join([]string{"relationship", relationship.RelationType, relationship.Source}, " ")),
	}
}
