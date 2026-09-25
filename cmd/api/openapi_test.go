package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestOpenAPISpecificationIsValidYAML(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "docs", "openapi.yaml"))
	require.NoError(t, err)

	var specification map[string]any
	require.NoError(t, yaml.Unmarshal(content, &specification))
	require.Equal(t, "3.1.0", specification["openapi"])
	require.Contains(t, specification["paths"], "/tickets/{id}/status")
	require.Contains(t, specification["paths"], "/analytics/overview")
}
