package index

import (
	"strings"
	"testing"

	"github.com/lwnmengjing/ai-infra-operator/internal/model"
)

func TestRelationshipDocumentIncludesBothEndpointsAndType(t *testing.T) {
	doc := RelationshipDocument(
		model.Relationship{
			ID:            7,
			RelationType:  "wg_tunnel",
			Source:        "manual",
			MetadataJSON:  `{"interface":"wg0"}`,
			SrcResourceID: 1,
			DstResourceID: 2,
		},
		model.Resource{
			ResourceKey: "node/frankfurt-01",
			Hostname:    "fra-01",
		},
		model.Resource{
			ResourceKey: "node/london-01",
			Hostname:    "lon-01",
		},
	)

	if doc.DocType != "relationship" {
		t.Fatalf("DocType = %q, want relationship", doc.DocType)
	}
	if doc.RefID != 7 {
		t.Fatalf("RefID = %d, want 7", doc.RefID)
	}
	for _, want := range []string{"node/frankfurt-01", "wg_tunnel", "node/london-01", "wg0"} {
		if !strings.Contains(doc.Body, want) && !strings.Contains(doc.Title, want) {
			t.Fatalf("document does not include %q: title=%q body=%q", want, doc.Title, doc.Body)
		}
	}
}
