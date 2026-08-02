package architecture_test

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestCoreHasNoOutwardDependencies(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate architecture test")
	}
	coreDirectory := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "core"))

	forbidden := []string{
		"flag",
		"github.com/mattn/go-sqlite3",
		"github.com/mss-boot-io/meminfra/cmd/",
		"github.com/mss-boot-io/meminfra/internal/adapters/",
		"github.com/mss-boot-io/meminfra/internal/app/",
		"github.com/mss-boot-io/meminfra/internal/ingest/",
		"github.com/mss-boot-io/meminfra/internal/legacy/",
		"github.com/mss-boot-io/meminfra/internal/mcp",
		"github.com/mss-boot-io/meminfra/internal/model",
		"github.com/mss-boot-io/meminfra/internal/projection/",
		"github.com/mss-boot-io/meminfra/internal/store",
		"gorm.io/",
		"modernc.org/sqlite",
	}

	violations := make([]string, 0)
	err := filepath.WalkDir(coreDirectory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		for _, imported := range parsed.Imports {
			importPath, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				return fmt.Errorf("unquote import in %s: %w", path, err)
			}
			for _, prefix := range forbidden {
				if matchesForbiddenImport(importPath, prefix) {
					relative, relativeErr := filepath.Rel(coreDirectory, path)
					if relativeErr != nil {
						relative = path
					}
					violations = append(violations, fmt.Sprintf("%s imports %q", relative, importPath))
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan core dependencies: %v", err)
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("internal/core must remain adapter-independent:\n%s", strings.Join(violations, "\n"))
	}
}

func matchesForbiddenImport(importPath, prefix string) bool {
	if strings.HasSuffix(prefix, "/") {
		return strings.HasPrefix(importPath, prefix)
	}
	return importPath == prefix || strings.HasPrefix(importPath, prefix+"/")
}
