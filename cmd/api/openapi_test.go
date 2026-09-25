package main

import (
	"encoding/json"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"hnu-tickets/internal/platform/config"
)

const apiBasePath = "/api/v1"

type document struct {
	Swagger  string                       `json:"swagger" yaml:"swagger"`
	BasePath string                       `json:"basePath" yaml:"basePath"`
	Paths    map[string]map[string]any    `json:"paths" yaml:"paths"`
	Security map[string]map[string]string `json:"securityDefinitions" yaml:"securityDefinitions"`
}

func TestGeneratedDocumentDescribesTheAPI(t *testing.T) {
	doc := readDocument(t)

	require.Equal(t, "2.0", doc.Swagger)
	require.Equal(t, apiBasePath, doc.BasePath)
	require.Contains(t, doc.Security, "BearerAuth")
}

// TestEveryRouteIsDocumented keeps the generated document and the router in
// sync: regenerating with "make swagger" is the only way to satisfy it.
func TestEveryRouteIsDocumented(t *testing.T) {
	doc := readDocument(t)
	documented := make([]string, 0, len(doc.Paths))
	for path, operations := range doc.Paths {
		for method := range operations {
			documented = append(documented, strings.ToUpper(method)+" "+path)
		}
	}

	registered := make([]string, 0, len(documented))
	router, ok := testRouter().(chi.Routes)
	require.True(t, ok)
	require.NoError(t, chi.Walk(router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		path, isAPI := strings.CutPrefix(route, apiBasePath)
		if !isAPI {
			return nil
		}
		registered = append(registered, method+" "+strings.TrimSuffix(path, "/"))
		return nil
	}))

	require.ElementsMatch(t, documented, registered)
}

func TestGeneratedYAMLCoversTheSamePaths(t *testing.T) {
	var fromYAML document
	content, err := os.ReadFile(filepath.Join("..", "..", "docs", "swagger.yaml"))
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(content, &fromYAML))

	require.ElementsMatch(t, slices.Collect(maps.Keys(readDocument(t).Paths)), slices.Collect(maps.Keys(fromYAML.Paths)))
}

// TestDevelopmentStartupRegeneratesTheDocument covers the startup path that
// rebuilds the specification from the sources of a checkout.
func TestDevelopmentStartupRegeneratesTheDocument(t *testing.T) {
	cfg := testConfig()
	cfg.App.Environment = config.EnvironmentDevelopment

	var regenerated document
	require.NoError(t, json.Unmarshal([]byte(specification(cfg)), &regenerated))

	require.ElementsMatch(t,
		slices.Collect(maps.Keys(readDocument(t).Paths)),
		slices.Collect(maps.Keys(regenerated.Paths)),
	)
}

func readDocument(t *testing.T) document {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "..", "docs", "swagger.json"))
	require.NoError(t, err)

	var doc document
	require.NoError(t, json.Unmarshal(content, &doc))
	return doc
}
