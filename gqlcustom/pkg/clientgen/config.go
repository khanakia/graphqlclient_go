package clientgen

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tidwall/jsonc"
)

// Config holds the configuration for the SDK generator
type Config struct {
	// SchemaPath is the path to the GraphQL SDL file
	SchemaPath string
	// OutputDir is the directory where the generated SDK will be written
	OutputDir string
	// PackageName is the Go package name for the generated SDK
	PackageName string
	// ModulePath is the Go module path for the generated SDK
	ModulePath string

	// Config is the configuration for the generator
	ConfigPath string

	// Package is the Go package name for the generated SDK
	Package string
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.SchemaPath == "" {
		return ErrSchemaPathRequired
	}
	if c.OutputDir == "" {
		c.OutputDir = "./sdk"
	}
	if c.PackageName == "" {
		c.PackageName = "sdk"
	}
	if c.ConfigPath == "" {
		return ErrConfigPathRequired
	}
	return nil
}

func loadClientConfig(path string) (*ClientConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var config ClientConfig
	err = json.Unmarshal(jsonc.ToJSON(content), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}
