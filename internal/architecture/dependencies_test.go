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

const (
	moduleImportPrefix = "github.com/mss-boot-io/meminfra/"
	coreImportPrefix   = moduleImportPrefix + "internal/core"
)

var forbiddenCoreImports = []string{
	"database/sql",
	"flag",
	"github.com/mark3labs/mcp-go",
	"github.com/mattn/go-sqlite3",
	"github.com/modelcontextprotocol/go-sdk",
	"github.com/spf13/cobra",
	"google.golang.org/grpc",
	"gorm.io/",
	"modernc.org/sqlite",
	"net/http",
}

func TestCoreHasNoOutwardDependencies(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate architecture test")
	}
	coreDirectory := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "core"))

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
			if !isForbiddenCoreImport(importPath) {
				continue
			}

			relative, relativeErr := filepath.Rel(coreDirectory, path)
			if relativeErr != nil {
				relative = path
			}
			violations = append(violations, fmt.Sprintf("%s imports %q", relative, importPath))
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

func TestForbiddenCoreImportPolicy(t *testing.T) {
	tests := []struct {
		name       string
		importPath string
		forbidden  bool
	}{
		{name: "core subpackage", importPath: coreImportPrefix + "/domain", forbidden: false},
		{name: "standard context", importPath: "context", forbidden: false},
		{name: "legacy service", importPath: moduleImportPrefix + "internal/legacy/core", forbidden: true},
		{name: "future local adapter", importPath: moduleImportPrefix + "internal/adapters/sqlite", forbidden: true},
		{name: "old index", importPath: moduleImportPrefix + "internal/index", forbidden: true},
		{name: "database sql", importPath: "database/sql", forbidden: true},
		{name: "gorm", importPath: "gorm.io/gorm", forbidden: true},
		{name: "mcp sdk", importPath: "github.com/modelcontextprotocol/go-sdk/mcp", forbidden: true},
		{name: "http adapter", importPath: "net/http", forbidden: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isForbiddenCoreImport(test.importPath); got != test.forbidden {
				t.Fatalf("isForbiddenCoreImport(%q) = %v, want %v", test.importPath, got, test.forbidden)
			}
		})
	}
}

func isForbiddenCoreImport(importPath string) bool {
	if strings.HasPrefix(importPath, moduleImportPrefix) {
		return importPath != coreImportPrefix && !strings.HasPrefix(importPath, coreImportPrefix+"/")
	}

	for _, prefix := range forbiddenCoreImports {
		if matchesImport(importPath, prefix) {
			return true
		}
	}
	return false
}

func matchesImport(importPath, prefix string) bool {
	if strings.HasSuffix(prefix, "/") {
		return strings.HasPrefix(importPath, prefix)
	}
	return importPath == prefix || strings.HasPrefix(importPath, prefix+"/")
}
