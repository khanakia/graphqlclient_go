package generator

import (
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

// TypeMapper handles mapping GraphQL types to Go types
type TypeMapper struct {
	// CustomScalars maps custom scalar names to Go types
	CustomScalars map[string]string
}

// NewTypeMapper creates a new TypeMapper with default scalar mappings
func NewTypeMapper() *TypeMapper {
	return &TypeMapper{
		CustomScalars: map[string]string{
			"Time":     "time.Time",
			"DateTime": "time.Time",
			"Date":     "time.Time",
			"Cursor":   "string",
			"UUID":     "string",
			"JSON":     "json.RawMessage",
			"Map":      "map[string]interface{}",
			"Any":      "interface{}",
			"Password": "string",
			"Upload":   "io.Reader",
		},
	}
}

// GraphQLToGoType converts a GraphQL type to its Go equivalent
func (tm *TypeMapper) GraphQLToGoType(t *ast.Type) string {
	if t == nil {
		return "interface{}"
	}

	goType := tm.resolveType(t)

	// If nullable (not NonNull), make it a pointer (except for slices)
	if !t.NonNull && !strings.HasPrefix(goType, "[]") {
		goType = "*" + goType
	}

	return goType
}

// resolveType resolves the base Go type from a GraphQL type
func (tm *TypeMapper) resolveType(t *ast.Type) string {
	if t.Elem != nil {
		// It's a list type
		elemType := tm.GraphQLToGoType(t.Elem)
		return "[]" + elemType
	}

	// Named type
	return tm.namedTypeToGo(t.NamedType)
}

// namedTypeToGo converts a named GraphQL type to Go
func (tm *TypeMapper) namedTypeToGo(name string) string {
	// Built-in scalars
	switch name {
	case "String":
		return "string"
	case "Int":
		return "int"
	case "Float":
		return "float64"
	case "Boolean":
		return "bool"
	case "ID":
		return "string"
	}

	// Custom scalars
	if goType, ok := tm.CustomScalars[name]; ok {
		return goType
	}

	// User-defined types (keep the name)
	return name
}

// GoTypeForField returns the Go type for a GraphQL field, handling nullability
func (tm *TypeMapper) GoTypeForField(field *ast.FieldDefinition) string {
	return tm.GraphQLToGoType(field.Type)
}

// GoTypeForArg returns the Go type for a GraphQL argument
func (tm *TypeMapper) GoTypeForArg(arg *ast.ArgumentDefinition) string {
	return tm.GraphQLToGoType(arg.Type)
}

// GoTypeForInputField returns the Go type for a GraphQL input field
func (tm *TypeMapper) GoTypeForInputField(field *ast.FieldDefinition) string {
	return tm.GraphQLToGoType(field.Type)
}

// NeedsTimeImport checks if the type requires time package import
func (tm *TypeMapper) NeedsTimeImport(typeName string) bool {
	goType := tm.namedTypeToGo(typeName)
	return strings.Contains(goType, "time.Time")
}

// NeedsJSONImport checks if the type requires encoding/json package import
func (tm *TypeMapper) NeedsJSONImport(typeName string) bool {
	goType := tm.namedTypeToGo(typeName)
	return strings.Contains(goType, "json.RawMessage")
}

// FormatGoType formats a Go type name following Go conventions
func FormatGoType(name string) string {
	// Convert to PascalCase
	return ToPascalCase(name)
}

// ToPascalCase converts a string to PascalCase
func ToPascalCase(s string) string {
	if s == "" {
		return s
	}

	// Handle common acronyms
	acronyms := map[string]string{
		"id":   "ID",
		"url":  "URL",
		"api":  "API",
		"http": "HTTP",
		"json": "JSON",
		"xml":  "XML",
		"sql":  "SQL",
		"html": "HTML",
		"css":  "CSS",
		"uri":  "URI",
		"uuid": "UUID",
	}

	words := splitIntoWords(s)
	var result strings.Builder

	for _, word := range words {
		lower := strings.ToLower(word)
		if acronym, ok := acronyms[lower]; ok {
			result.WriteString(acronym)
		} else {
			result.WriteString(strings.Title(strings.ToLower(word)))
		}
	}

	return result.String()
}

// ToCamelCase converts a string to camelCase
func ToCamelCase(s string) string {
	pascal := ToPascalCase(s)
	if pascal == "" {
		return pascal
	}

	// Find the first lowercase letter position
	for i, r := range pascal {
		if r >= 'a' && r <= 'z' {
			if i == 0 {
				return pascal
			}
			if i == 1 {
				return strings.ToLower(string(pascal[0])) + pascal[1:]
			}
			// Handle acronyms like "ID" -> "id", "URL" -> "url"
			return strings.ToLower(pascal[:i-1]) + pascal[i-1:]
		}
	}

	// All uppercase (acronym)
	return strings.ToLower(pascal)
}

// splitIntoWords splits a string into words based on common delimiters and case changes
func splitIntoWords(s string) []string {
	var words []string
	var currentWord strings.Builder

	for i, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
			continue
		}

		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := rune(s[i-1])
			if prev >= 'a' && prev <= 'z' {
				if currentWord.Len() > 0 {
					words = append(words, currentWord.String())
					currentWord.Reset()
				}
			}
		}

		currentWord.WriteRune(r)
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

// JSONTag returns the JSON tag for a field
func JSONTag(fieldName string, omitempty bool) string {
	jsonName := ToJSONName(fieldName)
	if omitempty {
		return fmt.Sprintf("`json:\"%s,omitempty\"`", jsonName)
	}
	return fmt.Sprintf("`json:\"%s\"`", jsonName)
}

// ToJSONName converts a field name to its JSON representation
func ToJSONName(name string) string {
	return ToCamelCase(name)
}
