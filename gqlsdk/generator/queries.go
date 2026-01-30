package generator

import (
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

// QueryGenerator generates Go query functions from GraphQL schema
type QueryGenerator struct {
	schema     *ast.Schema
	typeMapper *TypeMapper
}

// NewQueryGenerator creates a new QueryGenerator
func NewQueryGenerator(schema *ast.Schema, tm *TypeMapper) *QueryGenerator {
	return &QueryGenerator{
		schema:     schema,
		typeMapper: tm,
	}
}

// GenerateQueries generates all query methods
func (qg *QueryGenerator) GenerateQueries(packageName string) string {
	if qg.schema.Query == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString(")\n\n")

	// Generate query builder types and methods for each query field
	for _, field := range qg.schema.Query.Fields {
		if strings.HasPrefix(field.Name, "__") {
			continue
		}
		sb.WriteString(qg.generateQueryMethod(field))
	}

	return sb.String()
}

// generateQueryMethod generates a query method for a field
func (qg *QueryGenerator) generateQueryMethod(field *ast.FieldDefinition) string {
	var sb strings.Builder

	methodName := ToPascalCase(field.Name)
	returnType := qg.typeMapper.GraphQLToGoType(field.Type)

	// Generate description comment
	if field.Description != "" {
		sb.WriteString(fmt.Sprintf("// %s %s\n", methodName, field.Description))
	}

	// Build method signature
	sb.WriteString(fmt.Sprintf("func (c *Client) %s(ctx context.Context", methodName))

	// Add arguments
	for _, arg := range field.Arguments {
		argName := ToCamelCase(arg.Name)
		argType := qg.typeMapper.GraphQLToGoType(arg.Type)
		sb.WriteString(fmt.Sprintf(", %s %s", argName, argType))
	}

	sb.WriteString(fmt.Sprintf(") (%s, error) {\n", returnType))

	// Generate query string
	queryString := qg.buildQueryString(field)
	sb.WriteString(fmt.Sprintf("\tconst query = `%s`\n\n", queryString))

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
	baseType := getBaseTypeName(field.Type)

	sb.WriteString(fmt.Sprintf("\tvar response struct {\n"))
	sb.WriteString(fmt.Sprintf("\t\tData struct {\n"))
	sb.WriteString(fmt.Sprintf("\t\t\t%s %s `json:\"%s\"`\n", ToPascalCase(field.Name), returnType, field.Name))
	sb.WriteString(fmt.Sprintf("\t\t} `json:\"data\"`\n"))
	sb.WriteString(fmt.Sprintf("\t}\n\n"))

	sb.WriteString("\tif err := c.Execute(ctx, query, variables, &response); err != nil {\n")

	// Return zero value based on type
	zeroValue := qg.getZeroValue(returnType)
	sb.WriteString(fmt.Sprintf("\t\treturn %s, err\n", zeroValue))
	sb.WriteString("\t}\n\n")

	sb.WriteString(fmt.Sprintf("\treturn response.Data.%s, nil\n", ToPascalCase(field.Name)))
	sb.WriteString("}\n\n")

	// Also generate a builder version for complex queries
	if qg.isConnectionType(field.Type) {
		sb.WriteString(qg.generateQueryBuilder(field, baseType))
	}

	return sb.String()
}

// buildQueryString builds a GraphQL query string for a field
func (qg *QueryGenerator) buildQueryString(field *ast.FieldDefinition) string {
	var sb strings.Builder

	// Query declaration with variables
	sb.WriteString("query ")
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
	sb.WriteString(qg.generateFieldSelection(field.Type, 2))
	sb.WriteString("\n}")

	return sb.String()
}

// generateFieldSelection generates the field selection for a type
func (qg *QueryGenerator) generateFieldSelection(t *ast.Type, depth int) string {
	if depth > 4 {
		return ""
	}

	typeName := getBaseTypeName(t)
	typeDef := qg.schema.Types[typeName]

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

	// Select appropriate fields based on type
	if qg.isConnectionType(t) {
		// For connection types, select common pagination fields
		sb.WriteString(indent + "totalCount\n")
		sb.WriteString(indent + "pageInfo {\n")
		sb.WriteString(indent + "\thasNextPage\n")
		sb.WriteString(indent + "\thasPreviousPage\n")
		sb.WriteString(indent + "\tstartCursor\n")
		sb.WriteString(indent + "\tendCursor\n")
		sb.WriteString(indent + "}\n")

		// Check for nodes or edges field
		for _, f := range typeDef.Fields {
			if f.Name == "nodes" || f.Name == "edges" {
				sb.WriteString(indent + f.Name)
				sb.WriteString(" ")
				sb.WriteString(qg.generateFieldSelection(f.Type, depth+1))
				sb.WriteString("\n")
			}
		}
	} else {
		for _, f := range typeDef.Fields {
			// Skip connection/relation fields to prevent infinite recursion
			if qg.isConnectionType(f.Type) || depth > 2 && isObjectType(qg.schema, f.Type) {
				continue
			}

			sb.WriteString(indent + f.Name)

			// Recursively add nested fields for object types
			if nested := qg.generateFieldSelection(f.Type, depth+1); nested != "" {
				sb.WriteString(" " + nested)
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString(strings.Repeat("\t", depth-1) + "}")

	return sb.String()
}

// generateQueryBuilder generates a fluent query builder for connection types
func (qg *QueryGenerator) generateQueryBuilder(field *ast.FieldDefinition, baseType string) string {
	var sb strings.Builder

	builderName := ToPascalCase(field.Name) + "QueryBuilder"
	methodName := ToPascalCase(field.Name) + "Builder"

	sb.WriteString(fmt.Sprintf("// %s provides a fluent interface for building %s queries\n", builderName, field.Name))
	sb.WriteString(fmt.Sprintf("type %s struct {\n", builderName))
	sb.WriteString("\tclient *Client\n")

	// Add common connection arguments
	for _, arg := range field.Arguments {
		argName := ToCamelCase(arg.Name)
		argType := qg.typeMapper.GraphQLToGoType(arg.Type)
		sb.WriteString(fmt.Sprintf("\t%s %s\n", argName, argType))
	}

	sb.WriteString("\tfields []string\n")
	sb.WriteString("}\n\n")

	// Constructor
	sb.WriteString(fmt.Sprintf("// %s creates a new query builder for %s\n", methodName, field.Name))
	sb.WriteString(fmt.Sprintf("func (c *Client) %s() *%s {\n", methodName, builderName))
	sb.WriteString(fmt.Sprintf("\treturn &%s{client: c}\n", builderName))
	sb.WriteString("}\n\n")

	// Builder methods for each argument
	for _, arg := range field.Arguments {
		argMethodName := ToPascalCase(arg.Name)
		argName := ToCamelCase(arg.Name)
		argType := qg.typeMapper.GraphQLToGoType(arg.Type)

		// Make non-pointer version for builder method parameter
		paramType := strings.TrimPrefix(argType, "*")

		sb.WriteString(fmt.Sprintf("// %s sets the %s argument\n", argMethodName, arg.Name))
		sb.WriteString(fmt.Sprintf("func (b *%s) %s(v %s) *%s {\n", builderName, argMethodName, paramType, builderName))

		if strings.HasPrefix(argType, "*") {
			sb.WriteString(fmt.Sprintf("\tb.%s = &v\n", argName))
		} else {
			sb.WriteString(fmt.Sprintf("\tb.%s = v\n", argName))
		}

		sb.WriteString("\treturn b\n")
		sb.WriteString("}\n\n")
	}

	// Execute method
	returnType := qg.typeMapper.GraphQLToGoType(field.Type)
	sb.WriteString(fmt.Sprintf("// Execute executes the query and returns the result\n"))
	sb.WriteString(fmt.Sprintf("func (b *%s) Execute(ctx context.Context) (%s, error) {\n", builderName, returnType))
	sb.WriteString(fmt.Sprintf("\treturn b.client.%s(ctx", ToPascalCase(field.Name)))

	for _, arg := range field.Arguments {
		argName := ToCamelCase(arg.Name)
		sb.WriteString(fmt.Sprintf(", b.%s", argName))
	}

	sb.WriteString(")\n")
	sb.WriteString("}\n\n")

	return sb.String()
}

// isConnectionType checks if a type is a Relay-style connection
func (qg *QueryGenerator) isConnectionType(t *ast.Type) bool {
	typeName := getBaseTypeName(t)
	return strings.HasSuffix(typeName, "Connection")
}

// getZeroValue returns the zero value for a Go type
func (qg *QueryGenerator) getZeroValue(goType string) string {
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

// getBaseTypeName extracts the base type name from a GraphQL type
func getBaseTypeName(t *ast.Type) string {
	if t.Elem != nil {
		return getBaseTypeName(t.Elem)
	}
	return t.NamedType
}

// formatGraphQLType formats a GraphQL type for use in query strings
func formatGraphQLType(t *ast.Type) string {
	var result string

	if t.Elem != nil {
		result = "[" + formatGraphQLType(t.Elem) + "]"
	} else {
		result = t.NamedType
	}

	if t.NonNull {
		result += "!"
	}

	return result
}

// isObjectType checks if a type is an object type (not scalar/enum)
func isObjectType(schema *ast.Schema, t *ast.Type) bool {
	typeName := getBaseTypeName(t)
	typeDef := schema.Types[typeName]
	if typeDef == nil {
		return false
	}
	return typeDef.Kind == ast.Object || typeDef.Kind == ast.Interface
}
