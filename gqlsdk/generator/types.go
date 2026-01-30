package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

// TypeGenerator generates Go types from GraphQL schema
type TypeGenerator struct {
	schema     *ast.Schema
	typeMapper *TypeMapper
}

// NewTypeGenerator creates a new TypeGenerator
func NewTypeGenerator(schema *ast.Schema, tm *TypeMapper) *TypeGenerator {
	return &TypeGenerator{
		schema:     schema,
		typeMapper: tm,
	}
}

// GenerateTypes generates all Go struct definitions for GraphQL types
func (tg *TypeGenerator) GenerateTypes(packageName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// Collect imports
	imports := tg.collectTypeImports()
	if len(imports) > 0 {
		sb.WriteString("import (\n")
		for _, imp := range imports {
			sb.WriteString(fmt.Sprintf("\t%q\n", imp))
		}
		sb.WriteString(")\n\n")
	}

	// Get sorted type names for deterministic output
	typeNames := tg.getSortedTypeNames()

	// Generate structs for each type
	for _, name := range typeNames {
		def := tg.schema.Types[name]
		if def.BuiltIn || strings.HasPrefix(name, "__") {
			continue
		}

		switch def.Kind {
		case ast.Object:
			if name != "Query" && name != "Mutation" && name != "Subscription" {
				sb.WriteString(tg.generateStruct(def))
			}
		case ast.Interface:
			sb.WriteString(tg.generateInterface(def))
		}
	}

	return sb.String()
}

// GenerateEnums generates all enum definitions
func (tg *TypeGenerator) GenerateEnums(packageName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	typeNames := tg.getSortedTypeNames()

	for _, name := range typeNames {
		def := tg.schema.Types[name]
		if def.BuiltIn || strings.HasPrefix(name, "__") {
			continue
		}

		if def.Kind == ast.Enum {
			sb.WriteString(tg.generateEnum(def))
		}
	}

	return sb.String()
}

// GenerateInputs generates all input type definitions
func (tg *TypeGenerator) GenerateInputs(packageName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// Collect imports
	imports := tg.collectInputImports()
	if len(imports) > 0 {
		sb.WriteString("import (\n")
		for _, imp := range imports {
			sb.WriteString(fmt.Sprintf("\t%q\n", imp))
		}
		sb.WriteString(")\n\n")
	}

	typeNames := tg.getSortedTypeNames()

	for _, name := range typeNames {
		def := tg.schema.Types[name]
		if def.BuiltIn || strings.HasPrefix(name, "__") {
			continue
		}

		if def.Kind == ast.InputObject {
			sb.WriteString(tg.generateInputStruct(def))
		}
	}

	return sb.String()
}

// GenerateScalars generates custom scalar type definitions
func (tg *TypeGenerator) GenerateScalars(packageName string) string {
	var sb strings.Builder

	// Known scalar mappings
	knownScalars := map[string]string{
		"Time":     "time.Time",
		"DateTime": "time.Time",
		"Date":     "time.Time",
		"Cursor":   "string",
		"UUID":     "string",
		"JSON":     "json.RawMessage",
		"Map":      "map[string]interface{}",
		"Any":      "interface{}",
		"Password": "string",
		"Upload":   "string",
	}

	// Built-in scalars to skip
	builtInScalars := map[string]bool{
		"String":  true,
		"Int":     true,
		"Float":   true,
		"Boolean": true,
		"ID":      true,
	}

	// First pass: collect which imports are needed based on actual scalars in schema
	needsTime := false
	needsJSON := false

	typeNames := tg.getSortedTypeNames()
	for _, name := range typeNames {
		def := tg.schema.Types[name]
		if def.BuiltIn || builtInScalars[name] {
			continue
		}

		if def.Kind == ast.Scalar {
			goType, ok := knownScalars[name]
			if ok {
				if strings.Contains(goType, "time.Time") {
					needsTime = true
				}
				if strings.Contains(goType, "json.RawMessage") {
					needsJSON = true
				}
			}
		}
	}

	// Write package declaration
	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// Write imports only if needed
	if needsTime || needsJSON {
		sb.WriteString("import (\n")
		if needsJSON {
			sb.WriteString("\t\"encoding/json\"\n")
		}
		if needsTime {
			sb.WriteString("\t\"time\"\n")
		}
		sb.WriteString(")\n\n")
	}

	// Generate type aliases for custom scalars
	sb.WriteString("// Custom scalar type aliases\n")
	sb.WriteString("type (\n")

	for _, name := range typeNames {
		def := tg.schema.Types[name]
		if def.BuiltIn || builtInScalars[name] {
			continue
		}

		if def.Kind == ast.Scalar {
			goType, ok := knownScalars[name]
			if !ok {
				// Default unknown scalars to string
				goType = "string"
			}

			if def.Description != "" {
				lines := strings.Split(def.Description, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" {
						sb.WriteString(fmt.Sprintf("\t// %s\n", line))
					}
				}
			} else {
				sb.WriteString(fmt.Sprintf("\t// %s is a custom GraphQL scalar\n", name))
			}
			sb.WriteString(fmt.Sprintf("\t%s = %s\n\n", name, goType))
		}
	}

	sb.WriteString(")\n")

	return sb.String()
}

// generateStruct generates a Go struct for a GraphQL object type
func (tg *TypeGenerator) generateStruct(def *ast.Definition) string {
	var sb strings.Builder

	// Write description as comment
	if def.Description != "" {
		sb.WriteString(formatDescription(def.Name, def.Description))
	} else {
		sb.WriteString(fmt.Sprintf("// %s represents the GraphQL type %s\n", def.Name, def.Name))
	}

	sb.WriteString(fmt.Sprintf("type %s struct {\n", def.Name))

	for _, field := range def.Fields {
		sb.WriteString(tg.generateField(field))
	}

	sb.WriteString("}\n\n")

	return sb.String()
}

// generateInterface generates a Go interface for a GraphQL interface type
func (tg *TypeGenerator) generateInterface(def *ast.Definition) string {
	var sb strings.Builder

	if def.Description != "" {
		sb.WriteString(formatDescription(def.Name, def.Description))
	} else {
		sb.WriteString(fmt.Sprintf("// %s is a GraphQL interface type\n", def.Name))
	}

	// Generate as an interface with Is<Name> method
	sb.WriteString(fmt.Sprintf("type %s interface {\n", def.Name))
	sb.WriteString(fmt.Sprintf("\tIs%s()\n", def.Name))
	sb.WriteString("}\n\n")

	return sb.String()
}

// generateEnum generates a Go enum type
func (tg *TypeGenerator) generateEnum(def *ast.Definition) string {
	var sb strings.Builder

	if def.Description != "" {
		sb.WriteString(formatDescription(def.Name, def.Description))
	} else {
		sb.WriteString(fmt.Sprintf("// %s is a GraphQL enum type\n", def.Name))
	}

	sb.WriteString(fmt.Sprintf("type %s string\n\n", def.Name))

	sb.WriteString(fmt.Sprintf("// %s enum values\n", def.Name))
	sb.WriteString("const (\n")

	for _, val := range def.EnumValues {
		constName := fmt.Sprintf("%s%s", def.Name, ToPascalCase(val.Name))
		if val.Description != "" {
			// Add indented comment for enum constant
			lines := strings.Split(val.Description, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					sb.WriteString(fmt.Sprintf("\t// %s\n", line))
				}
			}
		}
		sb.WriteString(fmt.Sprintf("\t%s %s = %q\n", constName, def.Name, val.Name))
	}

	sb.WriteString(")\n\n")

	return sb.String()
}

// generateInputStruct generates a Go struct for a GraphQL input type
func (tg *TypeGenerator) generateInputStruct(def *ast.Definition) string {
	var sb strings.Builder

	if def.Description != "" {
		sb.WriteString(formatDescription(def.Name, def.Description))
	} else {
		sb.WriteString(fmt.Sprintf("// %s is a GraphQL input type\n", def.Name))
	}

	sb.WriteString(fmt.Sprintf("type %s struct {\n", def.Name))

	for _, field := range def.Fields {
		sb.WriteString(tg.generateField(field))
	}

	sb.WriteString("}\n\n")

	return sb.String()
}

// generateField generates a Go struct field from a GraphQL field
func (tg *TypeGenerator) generateField(field *ast.FieldDefinition) string {
	fieldName := ToPascalCase(field.Name)
	goType := tg.typeMapper.GraphQLToGoType(field.Type)

	// Determine if omitempty should be used
	omitempty := !field.Type.NonNull

	jsonTag := JSONTag(field.Name, omitempty)

	// Add field comment if there's a description
	comment := formatInlineComment(field.Description)

	return fmt.Sprintf("\t%s %s %s%s\n", fieldName, goType, jsonTag, comment)
}

// collectTypeImports collects necessary imports for types
func (tg *TypeGenerator) collectTypeImports() []string {
	imports := make(map[string]bool)

	for _, def := range tg.schema.Types {
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") {
			continue
		}

		if def.Kind == ast.Object || def.Kind == ast.Interface {
			for _, field := range def.Fields {
				tg.checkTypeForImports(field.Type, imports)
			}
		}
	}

	return sortedKeys(imports)
}

// collectInputImports collects necessary imports for input types
func (tg *TypeGenerator) collectInputImports() []string {
	imports := make(map[string]bool)

	for _, def := range tg.schema.Types {
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") {
			continue
		}

		if def.Kind == ast.InputObject {
			for _, field := range def.Fields {
				tg.checkTypeForImports(field.Type, imports)
			}
		}
	}

	return sortedKeys(imports)
}

// checkTypeForImports checks a type and adds necessary imports
func (tg *TypeGenerator) checkTypeForImports(t *ast.Type, imports map[string]bool) {
	if t.Elem != nil {
		tg.checkTypeForImports(t.Elem, imports)
		return
	}

	if tg.typeMapper.NeedsTimeImport(t.NamedType) {
		imports["time"] = true
	}
	if tg.typeMapper.NeedsJSONImport(t.NamedType) {
		imports["encoding/json"] = true
	}
}

// getSortedTypeNames returns sorted type names for deterministic output
func (tg *TypeGenerator) getSortedTypeNames() []string {
	names := make([]string, 0, len(tg.schema.Types))
	for name := range tg.schema.Types {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// sortedKeys returns sorted keys from a map
func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// formatDescription formats a description as Go comments
func formatDescription(name, description string) string {
	if description == "" {
		return ""
	}

	// Replace newlines with comment continuation
	lines := strings.Split(description, "\n")
	var sb strings.Builder

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if i == 0 {
			sb.WriteString(fmt.Sprintf("// %s %s\n", name, line))
		} else {
			sb.WriteString(fmt.Sprintf("// %s\n", line))
		}
	}

	return sb.String()
}

// formatInlineComment formats a description for inline comments (single line)
func formatInlineComment(description string) string {
	if description == "" {
		return ""
	}

	// Take only the first line and trim it
	lines := strings.Split(description, "\n")
	firstLine := strings.TrimSpace(lines[0])
	if firstLine == "" {
		return ""
	}

	return fmt.Sprintf(" // %s", firstLine)
}
