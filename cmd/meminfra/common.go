package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/mss-boot-io/meminfra/internal/legacy/core"
)

func openAndMigrate(ctx context.Context, dbPath string) (*core.Service, error) {
	return core.Open(ctx, dbPath)
}

func newFlagSet(name string, output io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(output)
	return fs
}

func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "help"
}

func parseUintFlag(value string, name string) (uint, error) {
	if strings.TrimSpace(value) == "" {
		return 0, fmt.Errorf("%s is required", name)
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return uint(parsed), nil
}
