package generator

import (
	"fmt"
	"os"
	"strings"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// extractLocalPackageName extracts the local package name from an import path
// e.g., "testsdk/api" -> "api", "mypackage" -> "mypackage"
func extractLocalPackageName(importPath string) string {
	if idx := strings.LastIndex(importPath, "/"); idx != -1 {
		return importPath[idx+1:]
	}
	return importPath
}

// Generator orchestrates the SDK generation process
type Generator struct {
	config     *Config
	schema     *ast.Schema
	typeMapper *TypeMapper
	writer     *Writer
}

// New creates a new Generator from the given configuration
func New(config *Config) (*Generator, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Parse schema
	schema, err := parseSchemaFile(config.SchemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	return &Generator{
		config:     config,
		schema:     schema,
		typeMapper: NewTypeMapper(),
		writer:     NewWriter(config.OutputDir),
	}, nil
}

// Generate generates the SDK
func (g *Generator) Generate() error {
	fmt.Printf("Generating SDK from %s\n", g.config.SchemaPath)
	fmt.Printf("Output directory: %s\n", g.config.OutputDir)
	fmt.Printf("Package name: %s\n", g.config.PackageName)

	// Ensure output directory exists
	if err := g.writer.EnsureDir(); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate go.mod
	// if err := g.writer.WriteGoMod(g.config.ModulePath, g.config.PackageName); err != nil {
	// 	return fmt.Errorf("failed to write go.mod: %w", err)
	// }
	// fmt.Println("  Generated: go.mod")

	// Generate scalars
	if err := g.generateScalars(); err != nil {
		return fmt.Errorf("failed to generate scalars: %w", err)
	}
	fmt.Println("  Generated: scalars.go")

	// Generate types
	if err := g.generateTypes(); err != nil {
		return fmt.Errorf("failed to generate types: %w", err)
	}
	fmt.Println("  Generated: types.go")

	// Generate enums
	if err := g.generateEnums(); err != nil {
		return fmt.Errorf("failed to generate enums: %w", err)
	}
	fmt.Println("  Generated: enums.go")

	// Generate inputs
	if err := g.generateInputs(); err != nil {
		return fmt.Errorf("failed to generate inputs: %w", err)
	}
	fmt.Println("  Generated: inputs.go")

	// Generate client
	if err := g.generateClient(); err != nil {
		return fmt.Errorf("failed to generate client: %w", err)
	}
	fmt.Println("  Generated: client.go")

	// Generate builder files (separate files for each query/mutation)
	if err := g.generateBuilderFiles(); err != nil {
		return fmt.Errorf("failed to generate builder files: %w", err)
	}

	fmt.Printf("SDK generated successfully in %s\n", g.config.OutputDir)
	return nil
}

func (g *Generator) generateScalars() error {
	tg := NewTypeGenerator(g.schema, g.typeMapper)
	localPkgName := extractLocalPackageName(g.config.PackageName)
	content := tg.GenerateScalars(localPkgName)
	return g.writer.WriteFileWithHeader("scalars.go", content)
}

func (g *Generator) generateTypes() error {
	tg := NewTypeGenerator(g.schema, g.typeMapper)
	localPkgName := extractLocalPackageName(g.config.PackageName)
	content := tg.GenerateTypes(localPkgName)
	return g.writer.WriteFileWithHeader("types.go", content)
}

func (g *Generator) generateEnums() error {
	tg := NewTypeGenerator(g.schema, g.typeMapper)
	localPkgName := extractLocalPackageName(g.config.PackageName)
	content := tg.GenerateEnums(localPkgName)
	return g.writer.WriteFileWithHeader("enums.go", content)
}

func (g *Generator) generateInputs() error {
	tg := NewTypeGenerator(g.schema, g.typeMapper)
	localPkgName := extractLocalPackageName(g.config.PackageName)
	content := tg.GenerateInputs(localPkgName)
	return g.writer.WriteFileWithHeader("inputs.go", content)
}

func (g *Generator) generateClient() error {
	cg := NewClientGenerator()
	localPkgName := extractLocalPackageName(g.config.PackageName)
	content := cg.GenerateClient(localPkgName)
	return g.writer.WriteFileWithHeader("client.go", content)
}

func (g *Generator) generateBuilderFiles() error {
	bg := NewBuilderGenerator(g.schema, g.typeMapper)
	files := bg.GenerateBuilderFiles(g.config.PackageName)

	for _, file := range files {
		if err := g.writer.WriteFileWithHeader(file.Filename, file.Content); err != nil {
			return fmt.Errorf("failed to write %s: %w", file.Filename, err)
		}
		fmt.Printf("  Generated: %s\n", file.Filename)
	}
	return nil
}

// parseSchemaFile reads and parses a GraphQL schema file
func parseSchemaFile(filepath string) (*ast.Schema, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	schema, gqlErr := gqlparser.LoadSchema(&ast.Source{
		Name:  filepath,
		Input: string(content),
	})
	if gqlErr != nil {
		return nil, fmt.Errorf("failed to parse schema: %v", gqlErr)
	}

	return schema, nil
}

// GetSchema returns the parsed schema
func (g *Generator) GetSchema() *ast.Schema {
	return g.schema
}

// GetTypeMapper returns the type mapper
func (g *Generator) GetTypeMapper() *TypeMapper {
	return g.typeMapper
}
