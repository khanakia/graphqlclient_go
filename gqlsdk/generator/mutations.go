package generator

import (
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

// MutationGenerator generates Go mutation functions from GraphQL schema
type MutationGenerator struct {
	schema     *ast.Schema
	typeMapper *TypeMapper
}

// NewMutationGenerator creates a new MutationGenerator
func NewMutationGenerator(schema *ast.Schema, tm *TypeMapper) *MutationGenerator {
	return &MutationGenerator{
		schema:     schema,
		typeMapper: tm,
	}
}

// GenerateMutations generates all mutation methods
func (mg *MutationGenerator) GenerateMutations(packageName string) string {
	if mg.schema.Mutation == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString(")\n\n")

	// Generate mutation methods
	for _, field := range mg.schema.Mutation.Fields {
		if strings.HasPrefix(field.Name, "__") {
			continue
		}
		sb.WriteString(mg.generateMutationMethod(field))
	}

	return sb.String()
}

// generateMutationMethod generates a mutation method for a field
func (mg *MutationGenerator) generateMutationMethod(field *ast.FieldDefinition) string {
	var sb strings.Builder

	// Add "Mutate" prefix to distinguish from queries with same name
	methodName := "Mutate" + ToPascalCase(field.Name)
	returnType := mg.typeMapper.GraphQLToGoType(field.Type)

	// Generate description comment
	if field.Description != "" {
		sb.WriteString(fmt.Sprintf("// %s %s\n", methodName, field.Description))
	}

	// Build method signature
	sb.WriteString(fmt.Sprintf("func (c *Client) %s(ctx context.Context", methodName))

	// Add arguments
	for _, arg := range field.Arguments {
		argName := ToCamelCase(arg.Name)
		argType := mg.typeMapper.GraphQLToGoType(arg.Type)
		sb.WriteString(fmt.Sprintf(", %s %s", argName, argType))
	}

	sb.WriteString(fmt.Sprintf(") (%s, error) {\n", returnType))

	// Generate mutation string
	mutationString := mg.buildMutationString(field)
	sb.WriteString(fmt.Sprintf("\tconst mutation = `%s`\n\n", mutationString))

	// Build variables map
	if len(field.Arguments) > 0 {
		sb.WriteString("\tvariables := map[string]interface{}{\n")
		for _, arg := range field.Arguments {
			argName := ToCamelCase(arg.Name)
			sb.WriteString(fmt.Sprintf("\t\t%q: %s,\n", arg.Name, argName))
		}
		sb.WriteString("\t}\n\n")
	} else {
		sb.WriteString("\tvariables := map[string]interface{}{}\n\n")
	}

	// Generate response handling
	sb.WriteString(fmt.Sprintf("\tvar response struct {\n"))
	sb.WriteString(fmt.Sprintf("\t\tData struct {\n"))
	sb.WriteString(fmt.Sprintf("\t\t\t%s %s `json:\"%s\"`\n", ToPascalCase(field.Name), returnType, field.Name))
	sb.WriteString(fmt.Sprintf("\t\t} `json:\"data\"`\n"))
	sb.WriteString(fmt.Sprintf("\t}\n\n"))

	sb.WriteString("\tif err := c.Execute(ctx, mutation, variables, &response); err != nil {\n")

	// Return zero value based on type
	zeroValue := mg.getZeroValue(returnType)
	sb.WriteString(fmt.Sprintf("\t\treturn %s, err\n", zeroValue))
	sb.WriteString("\t}\n\n")

	sb.WriteString(fmt.Sprintf("\treturn response.Data.%s, nil\n", ToPascalCase(field.Name)))
	sb.WriteString("}\n\n")

	return sb.String()
}

// buildMutationString builds a GraphQL mutation string for a field
func (mg *MutationGenerator) buildMutationString(field *ast.FieldDefinition) string {
	var sb strings.Builder

	// Mutation declaration with variables
	sb.WriteString("mutation ")
	sb.WriteString(ToPascalCase(field.Name))

	if len(field.Arguments) > 0 {
		sb.WriteString("(")
		args := make([]string, 0, len(field.Arguments))
		for _, arg := range field.Arguments {
			args = append(args, fmt.Sprintf("$%s: %s", arg.Name, formatGraphQLType(arg.Type)))
		}
		sb.WriteString(strings.Join(args, ", "))
		sb.WriteString(")")
	}

	sb.WriteString(" {\n")

	// Field with arguments
	sb.WriteString("\t")
	sb.WriteString(field.Name)

	if len(field.Arguments) > 0 {
		sb.WriteString("(")
		args := make([]string, 0, len(field.Arguments))
		for _, arg := range field.Arguments {
			args = append(args, fmt.Sprintf("%s: $%s", arg.Name, arg.Name))
		}
		sb.WriteString(strings.Join(args, ", "))
		sb.WriteString(")")
	}

	// Generate field selection
	sb.WriteString(" ")
	sb.WriteString(mg.generateFieldSelection(field.Type, 2))
	sb.WriteString("\n}")

	return sb.String()
}

// generateFieldSelection generates the field selection for a mutation result type
func (mg *MutationGenerator) generateFieldSelection(t *ast.Type, depth int) string {
	if depth > 3 {
		return ""
	}

	typeName := getBaseTypeName(t)
	typeDef := mg.schema.Types[typeName]

	if typeDef == nil || len(typeDef.Fields) == 0 {
		return ""
	}

	// Skip scalar types
	if typeDef.Kind == ast.Scalar || typeDef.Kind == ast.Enum {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("{\n")

	indent := strings.Repeat("\t", depth)

	for _, f := range typeDef.Fields {
		// Skip connection fields in mutation responses
		if strings.HasSuffix(getBaseTypeName(f.Type), "Connection") {
			continue
		}

		// Skip deeply nested object types to prevent infinite recursion
		if depth > 2 && isObjectType(mg.schema, f.Type) {
			continue
		}

		sb.WriteString(indent + f.Name)

		// Recursively add nested fields for object types
		if nested := mg.generateFieldSelection(f.Type, depth+1); nested != "" {
			sb.WriteString(" " + nested)
		}
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat("\t", depth-1) + "}")

	return sb.String()
}

// getZeroValue returns the zero value for a Go type
func (mg *MutationGenerator) getZeroValue(goType string) string {
	if strings.HasPrefix(goType, "*") {
		return "nil"
	}
	if strings.HasPrefix(goType, "[]") {
		return "nil"
	}
	switch goType {
	case "string":
		return `""`
	case "int", "int32", "int64", "float32", "float64":
		return "0"
	case "bool":
		return "false"
	default:
		return goType + "{}"
	}
}
