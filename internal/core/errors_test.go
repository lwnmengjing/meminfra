package core

import (
	"errors"
	"testing"
)

func TestApplicationErrorPreservesCodeAndCause(t *testing.T) {
	cause := errors.New("database unavailable")
	err := WrapError(CodeProjectionUnavailable, "query.state", "load projection", cause)

	if got := err.Error(); got != "query.state: load projection: database unavailable" {
		t.Fatalf("unexpected error string: %q", got)
	}
	if !errors.Is(err, cause) {
		t.Fatal("expected wrapped cause")
	}
	if !IsErrorCode(err, CodeProjectionUnavailable) {
		t.Fatalf("unexpected code: %#v", err)
	}
}

func TestFieldErrorIncludesField(t *testing.T) {
	err := NewFieldError(CodeInvalidArgument, "append.evidence", "source_ref", "source_ref is required")

	if got := err.Error(); got != "append.evidence: source_ref: source_ref is required" {
		t.Fatalf("unexpected error string: %q", got)
	}
	if code, ok := ErrorCodeOf(err); !ok || code != CodeInvalidArgument {
		t.Fatalf("unexpected code: %q, %v", code, ok)
	}
}
