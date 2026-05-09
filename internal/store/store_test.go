package store

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close store: %v", err)
		}
	})
	if err := store.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return store
}

func TestUpsertResourcePreservesFirstSeen(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	first, err := store.UpsertResource(ctx, ResourceInput{
		ResourceKey: "node/frankfurt-01",
		Kind:        "server",
		Hostname:    "frankfurt-01",
		Provider:    "ovh",
		Region:      "fra",
	})
	if err != nil {
		t.Fatalf("create resource: %v", err)
	}

	time.Sleep(time.Millisecond)

	second, err := store.UpsertResource(ctx, ResourceInput{
		ResourceKey: "node/frankfurt-01",
		Kind:        "server",
		Hostname:    "frankfurt-01",
		Provider:    "ovh",
		Region:      "fra",
		IPv4:        "192.0.2.10",
	})
	if err != nil {
		t.Fatalf("update resource: %v", err)
	}

	if !second.FirstSeen.Equal(first.FirstSeen) {
		t.Fatalf("first_seen changed: first=%s second=%s", first.FirstSeen, second.FirstSeen)
	}
	if !second.LastSeen.After(first.LastSeen) {
		t.Fatalf("last_seen did not advance: first=%s second=%s", first.LastSeen, second.LastSeen)
	}
	if second.IPv4 != "192.0.2.10" {
		t.Fatalf("ipv4 not updated: %q", second.IPv4)
	}
}

func TestAddObservationRequiresExistingResource(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	_, err := store.AddObservation(ctx, ObservationInput{
		ResourceKey: "node/missing",
		Metric:      "rtt_ms",
		Value:       82,
	})
	if !errors.Is(err, ErrResourceNotFound) {
		t.Fatalf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestAddObservationStoresMetricAndSearchDocument(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if _, err := store.UpsertResource(ctx, ResourceInput{
		ResourceKey: "node/frankfurt-01",
		Kind:        "server",
		Hostname:    "frankfurt-01",
		Region:      "fra",
	}); err != nil {
		t.Fatalf("resource: %v", err)
	}

	observedAt := time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC)
	observation, err := store.AddObservation(ctx, ObservationInput{
		ResourceKey: "node/frankfurt-01",
		Metric:      "rtt_ms",
		Value:       82,
		Unit:        "ms",
		ObservedAt:  observedAt,
	})
	if err != nil {
		t.Fatalf("add observation: %v", err)
	}
	if observation.Metric != "rtt_ms" || observation.Value != 82 || !observation.ObservedAt.Equal(observedAt) {
		t.Fatalf("unexpected observation: %#v", observation)
	}

	results, err := store.Search(ctx, "frankfurt rtt_ms", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected search result")
	}
}

func TestAddEventStoresEventAndSearchDocument(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if _, err := store.UpsertResource(ctx, ResourceInput{
		ResourceKey: "node/frankfurt-01",
		Kind:        "server",
		Hostname:    "frankfurt-01",
		Provider:    "ovh",
	}); err != nil {
		t.Fatalf("resource: %v", err)
	}

	event, err := store.AddEvent(ctx, EventInput{
		ResourceKey:   "node/frankfurt-01",
		EventType:     "ipv6_changed",
		EventDataJSON: `{"new_ipv6":"2001:db8::1"}`,
	})
	if err != nil {
		t.Fatalf("add event: %v", err)
	}
	if event.EventType != "ipv6_changed" {
		t.Fatalf("unexpected event type: %q", event.EventType)
	}

	results, err := store.Search(ctx, "ipv6_changed", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 || results[0].DocType != "event" {
		t.Fatalf("unexpected search results: %#v", results)
	}
}

func TestAddIncidentStoresIncidentAndSearchDocument(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	incident, err := store.AddIncident(ctx, IncidentInput{
		Title:        "Frankfurt RTT spike",
		Symptoms:     "RTT increased from 30ms to 180ms",
		RootCause:    "OVH upstream congestion",
		Solution:     "Shift traffic to London",
		Result:       "Latency recovered",
		Tags:         "frankfurt rtt ovh",
		MetadataJSON: `{"region":"fra"}`,
	})
	if err != nil {
		t.Fatalf("add incident: %v", err)
	}
	if incident.Title != "Frankfurt RTT spike" || incident.MetadataJSON != `{"region":"fra"}` {
		t.Fatalf("unexpected incident: %#v", incident)
	}

	results, err := store.Search(ctx, "OVH congestion", 10)
	if err != nil {
		t.Fatalf("search incident: %v", err)
	}
	if len(results) != 1 || results[0].DocType != "incident" {
		t.Fatalf("unexpected search results: %#v", results)
	}
}

func TestAddRelationshipStoresRelationshipAndSearchDocument(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if _, err := store.UpsertResource(ctx, ResourceInput{
		ResourceKey: "node/frankfurt-01",
		Kind:        "server",
		Hostname:    "frankfurt-01",
	}); err != nil {
		t.Fatalf("src resource: %v", err)
	}
	if _, err := store.UpsertResource(ctx, ResourceInput{
		ResourceKey: "node/london-01",
		Kind:        "server",
		Hostname:    "london-01",
	}); err != nil {
		t.Fatalf("dst resource: %v", err)
	}

	relationship, err := store.AddRelationship(ctx, RelationshipInput{
		SrcResourceKey: "node/frankfurt-01",
		DstResourceKey: "node/london-01",
		RelationType:   "wg_tunnel",
		MetadataJSON:   `{"interface":"wg0"}`,
	})
	if err != nil {
		t.Fatalf("add relationship: %v", err)
	}
	if relationship.RelationType != "wg_tunnel" || relationship.MetadataJSON != `{"interface":"wg0"}` {
		t.Fatalf("unexpected relationship: %#v", relationship)
	}

	results, err := store.Search(ctx, "frankfurt wg_tunnel london", 10)
	if err != nil {
		t.Fatalf("search relationship: %v", err)
	}
	if len(results) != 1 || results[0].DocType != "relationship" {
		t.Fatalf("unexpected search results: %#v", results)
	}
}

func TestQueryTopologyReturnsDirectedEdges(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	for _, key := range []string{"node/frankfurt-01", "node/london-01", "node/paris-01"} {
		if _, err := store.UpsertResource(ctx, ResourceInput{
			ResourceKey: key,
			Kind:        "server",
			Hostname:    strings.TrimPrefix(key, "node/"),
		}); err != nil {
			t.Fatalf("resource %s: %v", key, err)
		}
	}
	if _, err := store.AddRelationship(ctx, RelationshipInput{
		SrcResourceKey: "node/frankfurt-01",
		DstResourceKey: "node/london-01",
		RelationType:   "wg_tunnel",
	}); err != nil {
		t.Fatalf("outgoing relationship: %v", err)
	}
	if _, err := store.AddRelationship(ctx, RelationshipInput{
		SrcResourceKey: "node/paris-01",
		DstResourceKey: "node/frankfurt-01",
		RelationType:   "service_dependency",
	}); err != nil {
		t.Fatalf("incoming relationship: %v", err)
	}

	outgoing, err := store.QueryTopology(ctx, TopologyQueryOptions{
		ResourceKey: "node/frankfurt-01",
		Direction:   "out",
	})
	if err != nil {
		t.Fatalf("query outgoing topology: %v", err)
	}
	if len(outgoing) != 1 || outgoing[0].SrcResource.ResourceKey != "node/frankfurt-01" || outgoing[0].DstResource.ResourceKey != "node/london-01" {
		t.Fatalf("unexpected outgoing edges: %#v", outgoing)
	}

	incoming, err := store.QueryTopology(ctx, TopologyQueryOptions{
		ResourceKey: "node/frankfurt-01",
		Direction:   "in",
	})
	if err != nil {
		t.Fatalf("query incoming topology: %v", err)
	}
	if len(incoming) != 1 || incoming[0].SrcResource.ResourceKey != "node/paris-01" || incoming[0].DstResource.ResourceKey != "node/frankfurt-01" {
		t.Fatalf("unexpected incoming edges: %#v", incoming)
	}

	wgOnly, err := store.QueryTopology(ctx, TopologyQueryOptions{
		ResourceKey:  "node/frankfurt-01",
		RelationType: "wg_tunnel",
		Direction:    "both",
	})
	if err != nil {
		t.Fatalf("query filtered topology: %v", err)
	}
	if len(wgOnly) != 1 || wgOnly[0].Relationship.RelationType != "wg_tunnel" {
		t.Fatalf("unexpected filtered edges: %#v", wgOnly)
	}
}

func TestGetAndListMethods(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	resource, err := store.UpsertResource(ctx, ResourceInput{
		ResourceKey: "node/frankfurt-01",
		Kind:        "server",
		Hostname:    "frankfurt-01",
	})
	if err != nil {
		t.Fatalf("resource: %v", err)
	}
	resources, err := store.ListResources(ctx, ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("list resources: %v", err)
	}
	if len(resources) != 1 || resources[0].ID != resource.ID {
		t.Fatalf("unexpected resources: %#v", resources)
	}

	observation, err := store.AddObservation(ctx, ObservationInput{
		ResourceKey: "node/frankfurt-01",
		Metric:      "rtt_ms",
		Value:       82,
	})
	if err != nil {
		t.Fatalf("observation: %v", err)
	}
	gotObservation, err := store.ObservationByID(ctx, observation.ID)
	if err != nil {
		t.Fatalf("get observation: %v", err)
	}
	if gotObservation.ID != observation.ID {
		t.Fatalf("unexpected observation: %#v", gotObservation)
	}
	observations, err := store.ListObservations(ctx, ObservationListOptions{
		ResourceKey: "node/frankfurt-01",
		Metric:      "rtt_ms",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("list observations: %v", err)
	}
	if len(observations) != 1 || observations[0].ID != observation.ID {
		t.Fatalf("unexpected observations: %#v", observations)
	}

	event, err := store.AddEvent(ctx, EventInput{
		ResourceKey: "node/frankfurt-01",
		EventType:   "rtt_spike",
	})
	if err != nil {
		t.Fatalf("event: %v", err)
	}
	gotEvent, err := store.EventByID(ctx, event.ID)
	if err != nil {
		t.Fatalf("get event: %v", err)
	}
	if gotEvent.ID != event.ID {
		t.Fatalf("unexpected event: %#v", gotEvent)
	}
	events, err := store.ListEvents(ctx, EventListOptions{
		ResourceKey: "node/frankfurt-01",
		EventType:   "rtt_spike",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 || events[0].ID != event.ID {
		t.Fatalf("unexpected events: %#v", events)
	}

	incident, err := store.AddIncident(ctx, IncidentInput{Title: "Frankfurt RTT spike"})
	if err != nil {
		t.Fatalf("incident: %v", err)
	}
	gotIncident, err := store.IncidentByID(ctx, incident.ID)
	if err != nil {
		t.Fatalf("get incident: %v", err)
	}
	if gotIncident.ID != incident.ID {
		t.Fatalf("unexpected incident: %#v", gotIncident)
	}
	incidents, err := store.ListIncidents(ctx, IncidentListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("list incidents: %v", err)
	}
	if len(incidents) != 1 || incidents[0].ID != incident.ID {
		t.Fatalf("unexpected incidents: %#v", incidents)
	}

	if _, err := store.UpsertResource(ctx, ResourceInput{
		ResourceKey: "node/london-01",
		Kind:        "server",
	}); err != nil {
		t.Fatalf("second resource: %v", err)
	}
	relationship, err := store.AddRelationship(ctx, RelationshipInput{
		SrcResourceKey: "node/frankfurt-01",
		DstResourceKey: "node/london-01",
		RelationType:   "service_dependency",
	})
	if err != nil {
		t.Fatalf("relationship: %v", err)
	}
	gotRelationship, err := store.RelationshipByID(ctx, relationship.ID)
	if err != nil {
		t.Fatalf("get relationship: %v", err)
	}
	if gotRelationship.ID != relationship.ID {
		t.Fatalf("unexpected relationship: %#v", gotRelationship)
	}
	relationships, err := store.ListRelationships(ctx, RelationshipListOptions{
		ResourceKey:  "node/frankfurt-01",
		RelationType: "service_dependency",
		Limit:        10,
	})
	if err != nil {
		t.Fatalf("list relationships: %v", err)
	}
	if len(relationships) != 1 || relationships[0].ID != relationship.ID {
		t.Fatalf("unexpected relationships: %#v", relationships)
	}
}

func TestListWithMissingResourceFilterReturnsEmpty(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	observations, err := store.ListObservations(ctx, ObservationListOptions{
		ResourceKey: "node/missing",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("list observations with missing resource: %v", err)
	}
	if len(observations) != 0 {
		t.Fatalf("expected empty observations, got %#v", observations)
	}

	events, err := store.ListEvents(ctx, EventListOptions{
		ResourceKey: "node/missing",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("list events with missing resource: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected empty events, got %#v", events)
	}

	relationships, err := store.ListRelationships(ctx, RelationshipListOptions{
		ResourceKey: "node/missing",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("list relationships with missing resource: %v", err)
	}
	if len(relationships) != 0 {
		t.Fatalf("expected empty relationships, got %#v", relationships)
	}
}

func TestSearchEscapesFTSSpecialCharacters(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if _, err := store.UpsertResource(ctx, ResourceInput{
		ResourceKey: "node/ipv6-01",
		Kind:        "server",
		Hostname:    "ipv6-01",
		IPv6:        "2001:db8::1",
		MetadataJSON: `{
			"note": "contains colon:and-json"
		}`,
	}); err != nil {
		t.Fatalf("resource: %v", err)
	}

	results, err := store.Search(ctx, "2001:db8::1", 10)
	if err != nil {
		t.Fatalf("search IPv6: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected IPv6 search result")
	}

	results, err = store.Search(ctx, `colon:and-json`, 10)
	if err != nil {
		t.Fatalf("search punctuation token: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected punctuation search result")
	}
}

func TestRejectsInvalidJSONFields(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if _, err := store.UpsertResource(ctx, ResourceInput{
		ResourceKey:  "node/bad",
		Kind:         "server",
		MetadataJSON: "{bad-json}",
	}); err == nil || !strings.Contains(err.Error(), "metadata_json must be valid JSON") {
		t.Fatalf("expected invalid resource metadata error, got %v", err)
	}

	if _, err := store.UpsertResource(ctx, ResourceInput{
		ResourceKey: "node/frankfurt-01",
		Kind:        "server",
	}); err != nil {
		t.Fatalf("resource: %v", err)
	}

	if _, err := store.AddObservation(ctx, ObservationInput{
		ResourceKey:  "node/frankfurt-01",
		Metric:       "rtt_ms",
		Value:        82,
		MetadataJSON: "{bad-json}",
	}); err == nil || !strings.Contains(err.Error(), "metadata_json must be valid JSON") {
		t.Fatalf("expected invalid observation metadata error, got %v", err)
	}

	if _, err := store.AddEvent(ctx, EventInput{
		ResourceKey:   "node/frankfurt-01",
		EventType:     "rtt_spike",
		EventDataJSON: "{bad-json}",
	}); err == nil || !strings.Contains(err.Error(), "event_data_json must be valid JSON") {
		t.Fatalf("expected invalid event data error, got %v", err)
	}

	if _, err := store.AddIncident(ctx, IncidentInput{
		Title:        "Bad incident",
		MetadataJSON: "{bad-json}",
	}); err == nil || !strings.Contains(err.Error(), "metadata_json must be valid JSON") {
		t.Fatalf("expected invalid incident metadata error, got %v", err)
	}

	if _, err := store.AddRelationship(ctx, RelationshipInput{
		SrcResourceKey: "node/frankfurt-01",
		DstResourceKey: "node/frankfurt-01",
		RelationType:   "bad_json",
		MetadataJSON:   "{bad-json}",
	}); err == nil || !strings.Contains(err.Error(), "metadata_json must be valid JSON") {
		t.Fatalf("expected invalid relationship metadata error, got %v", err)
	}
}

func TestMigrateCanRunTwice(t *testing.T) {
	store := newTestStore(t)

	if err := store.Migrate(context.Background()); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
}
