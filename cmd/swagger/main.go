// Command swagger regenerates the OpenAPI specification in docs/ from the
// annotations and DTO types of this module. It runs the same generator the swag
// CLI runs, so the project needs no external tool.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/swaggo/swag"
	"github.com/swaggo/swag/gen"
)

func main() {
	root, err := moduleRoot()
	if err != nil {
		log.Fatalf("locate module root: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		log.Fatalf("enter module root: %v", err)
	}
	if err := gen.New().Build(config()); err != nil {
		log.Fatalf("generate specification: %v", err)
	}
}

func config() *gen.Config {
	return &gen.Config{
		SearchDir:          "./",
		MainAPIFile:        filepath.Join("cmd", "api", "main.go"),
		OutputDir:          "docs",
		OutputTypes:        []string{"go", "json", "yaml"},
		OverridesFile:      gen.DefaultOverridesFile,
		PropNamingStrategy: swag.CamelCase,
		CollectionFormat:   "csv",
		ParseDepth:         100,
		ParseInternal:      true,
		ParseGoList:        true,
	}
}

// moduleRoot walks up from the working directory so the generator produces the
// same files whether it is started by go:generate, make, or by hand.
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
			return "", fmt.Errorf("no go.mod above the working directory")
		}
		directory = parent
	}
}
