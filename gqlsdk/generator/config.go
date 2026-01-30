package generator

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
	return nil
}
