package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				SchemaPath:  "schema.graphql",
				OutputDir:   "./sdk",
				PackageName: "sdk",
			},
			wantErr: false,
		},
		{
			name: "missing schema path",
			config: &Config{
				OutputDir:   "./sdk",
				PackageName: "sdk",
			},
			wantErr: true,
		},
		{
			name: "defaults filled in",
			config: &Config{
				SchemaPath: "schema.graphql",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.config.OutputDir == "" {
				t.Error("Config.Validate() should set default OutputDir")
			}

			if !tt.wantErr && tt.config.PackageName == "" {
				t.Error("Config.Validate() should set default PackageName")
			}
		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	// Create a minimal test schema
	schemaContent := `
type Query {
	user(id: ID!): User
	users: [User!]!
}

type Mutation {
	createUser(name: String!): User
}

type User {
	id: ID!
	name: String!
	email: String
}

enum UserRole {
	ADMIN
	USER
}

input CreateUserInput {
	name: String!
	role: UserRole
}

scalar Time
`

	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "gqlsdk-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Write test schema
	schemaPath := filepath.Join(tempDir, "schema.graphql")
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema: %v", err)
	}

	// Create generator
	outputDir := filepath.Join(tempDir, "output")
	config := &Config{
		SchemaPath:  schemaPath,
		OutputDir:   outputDir,
		PackageName: "testapi",
		ModulePath:  "github.com/test/testapi",
	}

	gen, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Generate SDK
	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate SDK: %v", err)
	}

	// Verify generated files exist
	expectedFiles := []string{
		"go.mod",
		"scalars.go",
		"types.go",
		"enums.go",
		"inputs.go",
		"client.go",
		"builder.go",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(outputDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s not found", file)
		}
	}

	// Verify types.go contains User struct
	typesContent, err := os.ReadFile(filepath.Join(outputDir, "types.go"))
	if err != nil {
		t.Fatalf("Failed to read types.go: %v", err)
	}
	if !strings.Contains(string(typesContent), "type User struct") {
		t.Error("types.go should contain User struct")
	}

	// Verify enums.go contains UserRole
	enumsContent, err := os.ReadFile(filepath.Join(outputDir, "enums.go"))
	if err != nil {
		t.Fatalf("Failed to read enums.go: %v", err)
	}
	if !strings.Contains(string(enumsContent), "type UserRole string") {
		t.Error("enums.go should contain UserRole enum")
	}

	// Verify inputs.go contains CreateUserInput
	inputsContent, err := os.ReadFile(filepath.Join(outputDir, "inputs.go"))
	if err != nil {
		t.Fatalf("Failed to read inputs.go: %v", err)
	}
	if !strings.Contains(string(inputsContent), "type CreateUserInput struct") {
		t.Error("inputs.go should contain CreateUserInput struct")
	}

	// Verify builder.go contains query and mutation builders
	builderContent, err := os.ReadFile(filepath.Join(outputDir, "builder.go"))
	if err != nil {
		t.Fatalf("Failed to read builder.go: %v", err)
	}
	if !strings.Contains(string(builderContent), "type QueryRoot struct") {
		t.Error("builder.go should contain QueryRoot struct")
	}
	if !strings.Contains(string(builderContent), "type MutationRoot struct") {
		t.Error("builder.go should contain MutationRoot struct")
	}
	if !strings.Contains(string(builderContent), "func (q *QueryRoot) User") {
		t.Error("builder.go should contain User query builder method")
	}
	if !strings.Contains(string(builderContent), "CreateUserMutationBuilder") {
		t.Error("builder.go should contain CreateUser mutation builder")
	}

	// Verify client.go contains Client struct
	clientContent, err := os.ReadFile(filepath.Join(outputDir, "client.go"))
	if err != nil {
		t.Fatalf("Failed to read client.go: %v", err)
	}
	if !strings.Contains(string(clientContent), "type Client struct") {
		t.Error("client.go should contain Client struct")
	}
}

func TestGenerator_InvalidSchema(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gqlsdk-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Write invalid schema
	schemaPath := filepath.Join(tempDir, "schema.graphql")
	if err := os.WriteFile(schemaPath, []byte("invalid graphql schema {{{"), 0644); err != nil {
		t.Fatalf("Failed to write schema: %v", err)
	}

	config := &Config{
		SchemaPath:  schemaPath,
		OutputDir:   filepath.Join(tempDir, "output"),
		PackageName: "testapi",
	}

	_, err = New(config)
	if err == nil {
		t.Error("Expected error for invalid schema")
	}
}

func TestGenerator_NonexistentSchema(t *testing.T) {
	config := &Config{
		SchemaPath:  "/nonexistent/schema.graphql",
		OutputDir:   "/tmp/output",
		PackageName: "testapi",
	}

	_, err := New(config)
	if err == nil {
		t.Error("Expected error for nonexistent schema file")
	}
}
