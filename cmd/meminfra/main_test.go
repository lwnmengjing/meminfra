package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestSubcommandHelpIsVisible(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run(context.Background(), []string{"resource", "upsert", "-h"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("help returned error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Usage of resource upsert") {
		t.Fatalf("expected subcommand usage, got %q", output)
	}
	if !strings.Contains(output, "-metadata") || !strings.Contains(output, "-key") {
		t.Fatalf("expected subcommand flags, got %q", output)
	}
}
