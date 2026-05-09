package core

import (
	"context"

	"github.com/lwnmengjing/ai-infra-operator/internal/model"
	"github.com/lwnmengjing/ai-infra-operator/internal/store"
)

type ResourceInput = store.ResourceInput
type ObservationInput = store.ObservationInput
type EventInput = store.EventInput
type IncidentInput = store.IncidentInput
type RelationshipInput = store.RelationshipInput
type ListOptions = store.ListOptions
type ObservationListOptions = store.ObservationListOptions
type EventListOptions = store.EventListOptions
type IncidentListOptions = store.IncidentListOptions
type RelationshipListOptions = store.RelationshipListOptions
type TopologyQueryOptions = store.TopologyQueryOptions

type Service struct {
	store *store.Store
}

func Open(ctx context.Context, dbPath string) (*Service, error) {
	mem, err := store.Open(dbPath)
	if err != nil {
		return nil, err
	}
	if err := mem.Migrate(ctx); err != nil {
		mem.Close()
		return nil, err
	}
	return &Service{store: mem}, nil
}

func (s *Service) Close() error {
	return s.store.Close()
}

func (s *Service) UpsertResource(ctx context.Context, input ResourceInput) (*model.Resource, error) {
	return s.store.UpsertResource(ctx, input)
}

func (s *Service) ResourceByKey(ctx context.Context, key string) (*model.Resource, error) {
	return s.store.ResourceByKey(ctx, key)
}

func (s *Service) ListResources(ctx context.Context, options ListOptions) ([]model.Resource, error) {
	return s.store.ListResources(ctx, options)
}

func (s *Service) AddObservation(ctx context.Context, input ObservationInput) (*model.Observation, error) {
	return s.store.AddObservation(ctx, input)
}

func (s *Service) ObservationByID(ctx context.Context, id uint) (*model.Observation, error) {
	return s.store.ObservationByID(ctx, id)
}

func (s *Service) ListObservations(ctx context.Context, options ObservationListOptions) ([]model.Observation, error) {
	return s.store.ListObservations(ctx, options)
}

func (s *Service) AddEvent(ctx context.Context, input EventInput) (*model.Event, error) {
	return s.store.AddEvent(ctx, input)
}

func (s *Service) EventByID(ctx context.Context, id uint) (*model.Event, error) {
	return s.store.EventByID(ctx, id)
}

func (s *Service) ListEvents(ctx context.Context, options EventListOptions) ([]model.Event, error) {
	return s.store.ListEvents(ctx, options)
}

func (s *Service) AddIncident(ctx context.Context, input IncidentInput) (*model.Incident, error) {
	return s.store.AddIncident(ctx, input)
}

func (s *Service) IncidentByID(ctx context.Context, id uint) (*model.Incident, error) {
	return s.store.IncidentByID(ctx, id)
}

func (s *Service) ListIncidents(ctx context.Context, options IncidentListOptions) ([]model.Incident, error) {
	return s.store.ListIncidents(ctx, options)
}

func (s *Service) AddRelationship(ctx context.Context, input RelationshipInput) (*model.Relationship, error) {
	return s.store.AddRelationship(ctx, input)
}

func (s *Service) RelationshipByID(ctx context.Context, id uint) (*model.Relationship, error) {
	return s.store.RelationshipByID(ctx, id)
}

func (s *Service) ListRelationships(ctx context.Context, options RelationshipListOptions) ([]model.Relationship, error) {
	return s.store.ListRelationships(ctx, options)
}

func (s *Service) QueryTopology(ctx context.Context, options TopologyQueryOptions) ([]model.TopologyEdge, error) {
	return s.store.QueryTopology(ctx, options)
}

func (s *Service) Search(ctx context.Context, query string, limit int) ([]model.SearchResult, error) {
	return s.store.Search(ctx, query, limit)
}
