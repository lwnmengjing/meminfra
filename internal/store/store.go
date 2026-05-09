package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lwnmengjing/ai-infra-operator/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

var (
	ErrResourceNotFound    = errors.New("resource not found")
	ErrObservationNotFound = errors.New("observation not found")
	ErrEventNotFound       = errors.New("event not found")
	ErrIncidentNotFound    = errors.New("incident not found")
)

type Store struct {
	db *gorm.DB
}

type ResourceInput struct {
	ResourceKey  string
	Kind         string
	Hostname     string
	IPv4         string
	IPv6         string
	Provider     string
	Region       string
	Source       string
	MetadataJSON string
}

type ObservationInput struct {
	ResourceKey  string
	Metric       string
	Value        float64
	Unit         string
	Source       string
	MetadataJSON string
	ObservedAt   time.Time
}

type EventInput struct {
	ResourceKey   string
	EventType     string
	EventDataJSON string
	Source        string
	CreatedAt     time.Time
}

type IncidentInput struct {
	Title        string
	Symptoms     string
	RootCause    string
	Solution     string
	Result       string
	Tags         string
	Source       string
	MetadataJSON string
	CreatedAt    time.Time
}

type ListOptions struct {
	Limit int
}

type ObservationListOptions struct {
	ResourceKey string
	Metric      string
	Limit       int
}

type EventListOptions struct {
	ResourceKey string
	EventType   string
	Limit       int
}

type IncidentListOptions struct {
	Limit int
}

func Open(path string) (*Store, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("db path is required")
	}

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *Store) Migrate(ctx context.Context) error {
	if err := s.db.WithContext(ctx).AutoMigrate(
		&model.Resource{},
		&model.Observation{},
		&model.Event{},
		&model.Incident{},
		&model.MemoryDocument{},
	); err != nil {
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS memory_fts USING fts5(title, body, tags)`).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *Store) UpsertResource(ctx context.Context, input ResourceInput) (*model.Resource, error) {
	if strings.TrimSpace(input.ResourceKey) == "" {
		return nil, fmt.Errorf("resource key is required")
	}
	if strings.TrimSpace(input.Kind) == "" {
		return nil, fmt.Errorf("resource kind is required")
	}
	metadata, err := normalizeJSON(input.MetadataJSON, "metadata_json")
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	resource := model.Resource{
		ResourceKey:  input.ResourceKey,
		Kind:         input.Kind,
		Hostname:     input.Hostname,
		IPv4:         input.IPv4,
		IPv6:         input.IPv6,
		Provider:     input.Provider,
		Region:       input.Region,
		Source:       defaultSource(input.Source),
		MetadataJSON: metadata,
		FirstSeen:    now,
		LastSeen:     now,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "resource_key"}},
			DoUpdates: clause.Assignments(map[string]any{
				"kind":          resource.Kind,
				"hostname":      resource.Hostname,
				"ipv4":          resource.IPv4,
				"ipv6":          resource.IPv6,
				"provider":      resource.Provider,
				"region":        resource.Region,
				"source":        resource.Source,
				"metadata_json": resource.MetadataJSON,
				"last_seen":     resource.LastSeen,
			}),
		}).Create(&resource).Error; err != nil {
			return err
		}

		if err := tx.Where("resource_key = ?", input.ResourceKey).First(&resource).Error; err != nil {
			return err
		}
		return upsertMemoryDocument(tx, resourceDocument(resource))
	})
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

func (s *Store) AddObservation(ctx context.Context, input ObservationInput) (*model.Observation, error) {
	if strings.TrimSpace(input.ResourceKey) == "" {
		return nil, fmt.Errorf("resource key is required")
	}
	if strings.TrimSpace(input.Metric) == "" {
		return nil, fmt.Errorf("metric is required")
	}
	metadata, err := normalizeJSON(input.MetadataJSON, "metadata_json")
	if err != nil {
		return nil, err
	}

	var observation model.Observation
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		resource, err := findResource(tx, input.ResourceKey)
		if err != nil {
			return err
		}

		observedAt := input.ObservedAt.UTC()
		if observedAt.IsZero() {
			observedAt = time.Now().UTC()
		}

		observation = model.Observation{
			ResourceID:   resource.ID,
			Metric:       input.Metric,
			Value:        input.Value,
			Unit:         input.Unit,
			Source:       defaultSource(input.Source),
			MetadataJSON: metadata,
			ObservedAt:   observedAt,
		}
		if err := tx.Create(&observation).Error; err != nil {
			return err
		}

		return upsertMemoryDocument(tx, observationDocument(observation, *resource))
	})
	if err != nil {
		return nil, err
	}
	return &observation, nil
}

func (s *Store) AddEvent(ctx context.Context, input EventInput) (*model.Event, error) {
	if strings.TrimSpace(input.ResourceKey) == "" {
		return nil, fmt.Errorf("resource key is required")
	}
	if strings.TrimSpace(input.EventType) == "" {
		return nil, fmt.Errorf("event type is required")
	}
	eventData, err := normalizeJSON(input.EventDataJSON, "event_data_json")
	if err != nil {
		return nil, err
	}

	var event model.Event
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		resource, err := findResource(tx, input.ResourceKey)
		if err != nil {
			return err
		}

		createdAt := input.CreatedAt.UTC()
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}

		event = model.Event{
			ResourceID:    resource.ID,
			EventType:     input.EventType,
			EventDataJSON: eventData,
			Source:        defaultSource(input.Source),
			CreatedAt:     createdAt,
		}
		if err := tx.Create(&event).Error; err != nil {
			return err
		}

		return upsertMemoryDocument(tx, eventDocument(event, *resource))
	})
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *Store) AddIncident(ctx context.Context, input IncidentInput) (*model.Incident, error) {
	if strings.TrimSpace(input.Title) == "" {
		return nil, fmt.Errorf("incident title is required")
	}
	metadata, err := normalizeJSON(input.MetadataJSON, "metadata_json")
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	createdAt := input.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = now
	}

	incident := model.Incident{
		Title:        input.Title,
		Symptoms:     input.Symptoms,
		RootCause:    input.RootCause,
		Solution:     input.Solution,
		Result:       input.Result,
		Tags:         input.Tags,
		Source:       defaultSource(input.Source),
		MetadataJSON: metadata,
		CreatedAt:    createdAt,
		UpdatedAt:    now,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&incident).Error; err != nil {
			return err
		}
		return upsertMemoryDocument(tx, incidentDocument(incident))
	})
	if err != nil {
		return nil, err
	}
	return &incident, nil
}

func (s *Store) Search(ctx context.Context, query string, limit int) ([]model.SearchResult, error) {
	matchQuery := safeFTSQuery(query)
	if matchQuery == "" {
		return nil, fmt.Errorf("query is required")
	}
	if limit <= 0 {
		limit = 10
	}

	var results []model.SearchResult
	err := s.db.WithContext(ctx).Raw(`
SELECT
	memory_documents.id,
	memory_documents.doc_type,
	memory_documents.ref_id,
	memory_documents.title,
	memory_documents.body,
	memory_documents.tags,
	bm25(memory_fts) AS rank
FROM memory_fts
JOIN memory_documents ON memory_documents.id = memory_fts.rowid
WHERE memory_fts MATCH ?
ORDER BY rank
LIMIT ?`, matchQuery, limit).Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *Store) ResourceByKey(ctx context.Context, key string) (*model.Resource, error) {
	return findResource(s.db.WithContext(ctx), key)
}

func (s *Store) ListResources(ctx context.Context, options ListOptions) ([]model.Resource, error) {
	var resources []model.Resource
	err := s.db.WithContext(ctx).
		Order("last_seen DESC").
		Limit(normalizeLimit(options.Limit)).
		Find(&resources).Error
	return resources, err
}

func (s *Store) ObservationByID(ctx context.Context, id uint) (*model.Observation, error) {
	var observation model.Observation
	err := s.db.WithContext(ctx).First(&observation, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrObservationNotFound
	}
	if err != nil {
		return nil, err
	}
	return &observation, nil
}

func (s *Store) ListObservations(ctx context.Context, options ObservationListOptions) ([]model.Observation, error) {
	query := s.db.WithContext(ctx).Model(&model.Observation{})
	if strings.TrimSpace(options.ResourceKey) != "" {
		resource, err := findResource(s.db.WithContext(ctx), options.ResourceKey)
		if errors.Is(err, ErrResourceNotFound) {
			return []model.Observation{}, nil
		}
		if err != nil {
			return nil, err
		}
		query = query.Where("resource_id = ?", resource.ID)
	}
	if strings.TrimSpace(options.Metric) != "" {
		query = query.Where("metric = ?", options.Metric)
	}

	var observations []model.Observation
	err := query.
		Order("observed_at DESC").
		Limit(normalizeLimit(options.Limit)).
		Find(&observations).Error
	return observations, err
}

func (s *Store) EventByID(ctx context.Context, id uint) (*model.Event, error) {
	var event model.Event
	err := s.db.WithContext(ctx).First(&event, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrEventNotFound
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *Store) ListEvents(ctx context.Context, options EventListOptions) ([]model.Event, error) {
	query := s.db.WithContext(ctx).Model(&model.Event{})
	if strings.TrimSpace(options.ResourceKey) != "" {
		resource, err := findResource(s.db.WithContext(ctx), options.ResourceKey)
		if errors.Is(err, ErrResourceNotFound) {
			return []model.Event{}, nil
		}
		if err != nil {
			return nil, err
		}
		query = query.Where("resource_id = ?", resource.ID)
	}
	if strings.TrimSpace(options.EventType) != "" {
		query = query.Where("event_type = ?", options.EventType)
	}

	var events []model.Event
	err := query.
		Order("created_at DESC").
		Limit(normalizeLimit(options.Limit)).
		Find(&events).Error
	return events, err
}

func (s *Store) IncidentByID(ctx context.Context, id uint) (*model.Incident, error) {
	var incident model.Incident
	err := s.db.WithContext(ctx).First(&incident, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrIncidentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &incident, nil
}

func (s *Store) ListIncidents(ctx context.Context, options IncidentListOptions) ([]model.Incident, error) {
	var incidents []model.Incident
	err := s.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(normalizeLimit(options.Limit)).
		Find(&incidents).Error
	return incidents, err
}

func findResource(db *gorm.DB, key string) (*model.Resource, error) {
	var resource model.Resource
	err := db.Where("resource_key = ?", key).First(&resource).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrResourceNotFound
	}
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > 500 {
		return 500
	}
	return limit
}

func upsertMemoryDocument(tx *gorm.DB, doc model.MemoryDocument) error {
	now := time.Now().UTC()
	doc.UpdatedAt = now
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = now
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "doc_type"}, {Name: "ref_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"title":      doc.Title,
			"body":       doc.Body,
			"tags":       doc.Tags,
			"updated_at": doc.UpdatedAt,
		}),
	}).Create(&doc).Error; err != nil {
		return err
	}

	if err := tx.Where("doc_type = ? AND ref_id = ?", doc.DocType, doc.RefID).First(&doc).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM memory_fts WHERE rowid = ?", doc.ID).Error; err != nil {
		return err
	}
	return tx.Exec("INSERT INTO memory_fts(rowid, title, body, tags) VALUES (?, ?, ?, ?)", doc.ID, doc.Title, doc.Body, doc.Tags).Error
}

func resourceDocument(resource model.Resource) model.MemoryDocument {
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
			string(resource.MetadataJSON),
		}, " ")),
		Tags: strings.TrimSpace(strings.Join([]string{resource.Kind, resource.Provider, resource.Region, resource.Source}, " ")),
	}
}

func observationDocument(observation model.Observation, resource model.Resource) model.MemoryDocument {
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
			string(observation.MetadataJSON),
		}, " ")),
		Tags: strings.TrimSpace(strings.Join([]string{"observation", observation.Metric, observation.Unit, observation.Source}, " ")),
	}
}

func eventDocument(event model.Event, resource model.Resource) model.MemoryDocument {
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
			string(event.EventDataJSON),
		}, " ")),
		Tags: strings.TrimSpace(strings.Join([]string{"event", event.EventType, event.Source}, " ")),
	}
}

func incidentDocument(incident model.Incident) model.MemoryDocument {
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

func defaultSource(source string) string {
	if strings.TrimSpace(source) == "" {
		return "manual"
	}
	return source
}

func normalizeJSON(value string, field string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "null", nil
	}
	if !json.Valid([]byte(value)) {
		return "", fmt.Errorf("%s must be valid JSON", field)
	}
	return value, nil
}

func safeFTSQuery(query string) string {
	tokens := strings.Fields(query)
	if len(tokens) == 0 {
		return ""
	}

	phrases := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		phrases = append(phrases, `"`+strings.ReplaceAll(token, `"`, `""`)+`"`)
	}
	if len(phrases) == 0 {
		return ""
	}
	return strings.Join(phrases, " AND ")
}
