package domain

import (
	"encoding"
	"encoding/json"
	"testing"
)

func TestDomainSerializationInterfaces(t *testing.T) {
	var (
		_ encoding.TextMarshaler   = WorkspaceID{}
		_ encoding.TextUnmarshaler = (*WorkspaceID)(nil)
		_ json.Marshaler           = WorkspaceID{}
		_ json.Unmarshaler         = (*WorkspaceID)(nil)

		_ encoding.TextMarshaler   = ResourceKey{}
		_ encoding.TextUnmarshaler = (*ResourceKey)(nil)
		_ json.Marshaler           = ResourceKey{}
		_ json.Unmarshaler         = (*ResourceKey)(nil)

		_ encoding.TextMarshaler   = EvidenceKind("")
		_ encoding.TextUnmarshaler = (*EvidenceKind)(nil)
		_ json.Marshaler           = EvidenceKind("")
		_ json.Unmarshaler         = (*EvidenceKind)(nil)

		_ encoding.TextMarshaler   = EvidenceStatus("")
		_ encoding.TextUnmarshaler = (*EvidenceStatus)(nil)
		_ json.Marshaler           = EvidenceStatus("")
		_ json.Unmarshaler         = (*EvidenceStatus)(nil)

		_ encoding.TextMarshaler   = Confidence{}
		_ encoding.TextUnmarshaler = (*Confidence)(nil)
		_ json.Marshaler           = Confidence{}
		_ json.Unmarshaler         = (*Confidence)(nil)

		_ encoding.TextMarshaler   = ObservedAt{}
		_ encoding.TextUnmarshaler = (*ObservedAt)(nil)
		_ json.Marshaler           = ObservedAt{}
		_ json.Unmarshaler         = (*ObservedAt)(nil)

		_ encoding.TextMarshaler   = ValidityInterval{}
		_ encoding.TextUnmarshaler = (*ValidityInterval)(nil)
		_ json.Marshaler           = ValidityInterval{}
		_ json.Unmarshaler         = (*ValidityInterval)(nil)

		_ encoding.TextMarshaler   = SourceType{}
		_ encoding.TextUnmarshaler = (*SourceType)(nil)
		_ json.Marshaler           = SourceType{}
		_ json.Unmarshaler         = (*SourceType)(nil)
	)
}
