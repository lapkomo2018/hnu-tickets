// Package apidocs rebuilds the OpenAPI specification from the annotations and
// DTO types of this module.
package apidocs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/swaggo/swag"
	"github.com/swaggo/swag/gen"
)

// Generate writes docs/ and returns the specification it produced. The
// generator reads Go sources, so it only works in a checkout: the container
// image ships the specification compiled into the binary instead.
func Generate() (string, error) {
	root, err := moduleRoot()
	if err != nil {
		return "", err
	}
	if err := gen.New().Build(config(root)); err != nil {
		return "", fmt.Errorf("build specification: %w", err)
	}
	specification, err := os.ReadFile(filepath.Join(root, "docs", "swagger.json"))
	if err != nil {
		return "", fmt.Errorf("read generated specification: %w", err)
	}
	return string(specification), nil
}

func config(root string) *gen.Config {
	return &gen.Config{
		SearchDir:          root,
		MainAPIFile:        filepath.Join("cmd", "api", "main.go"),
		OutputDir:          filepath.Join(root, "docs"),
		OutputTypes:        []string{"go", "json", "yaml"},
		OverridesFile:      filepath.Join(root, gen.DefaultOverridesFile),
		PropNamingStrategy: swag.CamelCase,
		CollectionFormat:   "csv",
		ParseDepth:         100,
		ParseInternal:      true,
		ParseGoList:        true,
	}
}

// moduleRoot walks up from the working directory so the generator produces the
// same files whether it runs from go:generate, make, or the server itself.
func moduleRoot() (string, error) {
	directory, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", errors.New("no go.mod above the working directory")
		}
		directory = parent
	}
}
