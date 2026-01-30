package generator

import (
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
)

func TestTypeMapper_GraphQLToGoType(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		name     string
		gqlType  *ast.Type
		expected string
	}{
		{
			name:     "non-null string",
			gqlType:  &ast.Type{NamedType: "String", NonNull: true},
			expected: "string",
		},
		{
			name:     "nullable string",
			gqlType:  &ast.Type{NamedType: "String", NonNull: false},
			expected: "*string",
		},
		{
			name:     "non-null int",
			gqlType:  &ast.Type{NamedType: "Int", NonNull: true},
			expected: "int",
		},
		{
			name:     "nullable int",
			gqlType:  &ast.Type{NamedType: "Int", NonNull: false},
			expected: "*int",
		},
		{
			name:     "non-null float",
			gqlType:  &ast.Type{NamedType: "Float", NonNull: true},
			expected: "float64",
		},
		{
			name:     "non-null boolean",
			gqlType:  &ast.Type{NamedType: "Boolean", NonNull: true},
			expected: "bool",
		},
		{
			name:     "non-null ID",
			gqlType:  &ast.Type{NamedType: "ID", NonNull: true},
			expected: "string",
		},
		{
			name: "non-null list of non-null strings",
			gqlType: &ast.Type{
				Elem:    &ast.Type{NamedType: "String", NonNull: true},
				NonNull: true,
			},
			expected: "[]string",
		},
		{
			name: "nullable list of nullable strings",
			gqlType: &ast.Type{
				Elem:    &ast.Type{NamedType: "String", NonNull: false},
				NonNull: false,
			},
			expected: "[]*string",
		},
		{
			name:     "custom type",
			gqlType:  &ast.Type{NamedType: "User", NonNull: true},
			expected: "User",
		},
		{
			name:     "nullable custom type",
			gqlType:  &ast.Type{NamedType: "User", NonNull: false},
			expected: "*User",
		},
		{
			name:     "Time scalar",
			gqlType:  &ast.Type{NamedType: "Time", NonNull: true},
			expected: "time.Time",
		},
		{
			name:     "nullable Time scalar",
			gqlType:  &ast.Type{NamedType: "Time", NonNull: false},
			expected: "*time.Time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.GraphQLToGoType(tt.gqlType)
			if result != tt.expected {
				t.Errorf("GraphQLToGoType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"hello_world", "HelloWorld"},
		{"helloWorld", "HelloWorld"},
		{"id", "ID"},
		{"user_id", "UserID"},
		{"http_url", "HTTPURL"},
		{"", ""},
		{"already_pascal", "AlreadyPascal"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello", "hello"},
		{"HelloWorld", "helloWorld"},
		{"ID", "id"},
		{"UserID", "userID"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestJSONTag(t *testing.T) {
	tests := []struct {
		fieldName string
		omitempty bool
		expected  string
	}{
		{"userName", false, "`json:\"userName\"`"},
		{"userName", true, "`json:\"userName,omitempty\"`"},
		{"ID", false, "`json:\"id\"`"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			result := JSONTag(tt.fieldName, tt.omitempty)
			if result != tt.expected {
				t.Errorf("JSONTag(%q, %v) = %q, want %q", tt.fieldName, tt.omitempty, result, tt.expected)
			}
		})
	}
}
