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
)

const apiBasePath = "/api/v1"

type specification struct {
	Swagger  string                       `json:"swagger" yaml:"swagger"`
	BasePath string                       `json:"basePath" yaml:"basePath"`
	Paths    map[string]map[string]any    `json:"paths" yaml:"paths"`
	Security map[string]map[string]string `json:"securityDefinitions" yaml:"securityDefinitions"`
}

func TestGeneratedSpecificationDescribesTheAPI(t *testing.T) {
	spec := readSpecification(t)

	require.Equal(t, "2.0", spec.Swagger)
	require.Equal(t, apiBasePath, spec.BasePath)
	require.Contains(t, spec.Security, "BearerAuth")
}

// TestEveryRouteIsDocumented keeps the generated specification and the router in
// sync: regenerating with "make swagger" is the only way to satisfy it.
func TestEveryRouteIsDocumented(t *testing.T) {
	spec := readSpecification(t)
	documented := make([]string, 0, len(spec.Paths))
	for path, operations := range spec.Paths {
		for method := range operations {
			documented = append(documented, strings.ToUpper(method)+" "+path)
		}
	}

	registered := make([]string, 0, len(documented))
	router, ok := newRouter(nil, testConfig()).(chi.Routes)
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
	var fromYAML specification
	content, err := os.ReadFile(filepath.Join("..", "..", "docs", "swagger.yaml"))
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(content, &fromYAML))

	require.ElementsMatch(t, slices.Collect(maps.Keys(readSpecification(t).Paths)), slices.Collect(maps.Keys(fromYAML.Paths)))
}

func readSpecification(t *testing.T) specification {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "..", "docs", "swagger.json"))
	require.NoError(t, err)

	var spec specification
	require.NoError(t, json.Unmarshal(content, &spec))
	return spec
}
