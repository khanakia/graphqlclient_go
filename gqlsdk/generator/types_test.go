package generator

import (
	"strings"
	"testing"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func parseTestSchema(schemaContent string) (*ast.Schema, error) {
	return gqlparser.LoadSchema(&ast.Source{
		Name:  "test.graphql",
		Input: schemaContent,
	})
}

func TestTypeGenerator_GenerateTypes(t *testing.T) {
	schemaContent := `
type User {
	id: ID!
	name: String!
	email: String
	age: Int
}

type Post {
	id: ID!
	title: String!
	author: User!
}
`

	schema, err := parseTestSchema(schemaContent)
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	tm := NewTypeMapper()
	tg := NewTypeGenerator(schema, tm)

	result := tg.GenerateTypes("testpkg")

	// Check package declaration
	if !strings.Contains(result, "package testpkg") {
		t.Error("Generated types should contain package declaration")
	}

	// Check User struct
	if !strings.Contains(result, "type User struct") {
		t.Error("Generated types should contain User struct")
	}

	// Check field types
	if !strings.Contains(result, "ID string") {
		t.Error("Generated types should have ID as string")
	}

	if !strings.Contains(result, "Name string") {
		t.Error("Generated types should have Name as string")
	}

	if !strings.Contains(result, "Email *string") {
		t.Error("Generated types should have Email as *string (nullable)")
	}

	if !strings.Contains(result, "Age *int") {
		t.Error("Generated types should have Age as *int (nullable)")
	}

	// Check Post struct
	if !strings.Contains(result, "type Post struct") {
		t.Error("Generated types should contain Post struct")
	}
}

func TestTypeGenerator_GenerateEnums(t *testing.T) {
	schemaContent := `
enum UserRole {
	ADMIN
	USER
	GUEST
}

enum Status {
	ACTIVE
	INACTIVE
}
`

	schema, err := parseTestSchema(schemaContent)
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	tm := NewTypeMapper()
	tg := NewTypeGenerator(schema, tm)

	result := tg.GenerateEnums("testpkg")

	// Check package declaration
	if !strings.Contains(result, "package testpkg") {
		t.Error("Generated enums should contain package declaration")
	}

	// Check UserRole enum
	if !strings.Contains(result, "type UserRole string") {
		t.Error("Generated enums should contain UserRole type")
	}

	// Check enum values
	if !strings.Contains(result, `UserRoleAdmin UserRole = "ADMIN"`) {
		t.Error("Generated enums should contain ADMIN value")
	}

	if !strings.Contains(result, `UserRoleUser UserRole = "USER"`) {
		t.Error("Generated enums should contain USER value")
	}

	// Check Status enum
	if !strings.Contains(result, "type Status string") {
		t.Error("Generated enums should contain Status type")
	}
}

func TestTypeGenerator_GenerateInputs(t *testing.T) {
	schemaContent := `
input CreateUserInput {
	name: String!
	email: String
	role: String
}

input UpdateUserInput {
	name: String
	email: String
}
`

	schema, err := parseTestSchema(schemaContent)
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	tm := NewTypeMapper()
	tg := NewTypeGenerator(schema, tm)

	result := tg.GenerateInputs("testpkg")

	// Check package declaration
	if !strings.Contains(result, "package testpkg") {
		t.Error("Generated inputs should contain package declaration")
	}

	// Check CreateUserInput struct
	if !strings.Contains(result, "type CreateUserInput struct") {
		t.Error("Generated inputs should contain CreateUserInput struct")
	}

	// Check UpdateUserInput struct
	if !strings.Contains(result, "type UpdateUserInput struct") {
		t.Error("Generated inputs should contain UpdateUserInput struct")
	}

	// Check field types - name is required in CreateUserInput
	if !strings.Contains(result, "Name string") {
		t.Error("Generated inputs should have Name as string (non-null)")
	}
}

func TestTypeGenerator_GenerateScalars(t *testing.T) {
	schemaContent := `
scalar Time
scalar UUID
scalar JSON
scalar CustomScalar
`

	schema, err := parseTestSchema(schemaContent)
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	tm := NewTypeMapper()
	tg := NewTypeGenerator(schema, tm)

	result := tg.GenerateScalars("testpkg")

	// Check package declaration
	if !strings.Contains(result, "package testpkg") {
		t.Error("Generated scalars should contain package declaration")
	}

	// Check known scalars
	if !strings.Contains(result, "Time = time.Time") {
		t.Error("Generated scalars should map Time to time.Time")
	}

	if !strings.Contains(result, "UUID = string") {
		t.Error("Generated scalars should map UUID to string")
	}

	if !strings.Contains(result, "JSON = json.RawMessage") {
		t.Error("Generated scalars should map JSON to json.RawMessage")
	}

	// Check unknown scalars default to string
	if !strings.Contains(result, "CustomScalar = string") {
		t.Error("Generated scalars should default unknown scalars to string")
	}
}

func TestFormatDescription(t *testing.T) {
	tests := []struct {
		name        string
		typeName    string
		description string
		wantLines   []string
	}{
		{
			name:        "single line",
			typeName:    "User",
			description: "A user in the system",
			wantLines:   []string{"// User A user in the system"},
		},
		{
			name:        "multi line",
			typeName:    "User",
			description: "A user in the system\nWith multiple lines",
			wantLines:   []string{"// User A user in the system", "// With multiple lines"},
		},
		{
			name:        "empty description",
			typeName:    "User",
			description: "",
			wantLines:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDescription(tt.typeName, tt.description)

			for _, want := range tt.wantLines {
				if !strings.Contains(result, want) {
					t.Errorf("formatDescription() = %q, should contain %q", result, want)
				}
			}
		})
	}
}

func TestFormatInlineComment(t *testing.T) {
	tests := []struct {
		description string
		expected    string
	}{
		{"Simple comment", " // Simple comment"},
		{"Multi\nline\ncomment", " // Multi"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := formatInlineComment(tt.description)
			if result != tt.expected {
				t.Errorf("formatInlineComment(%q) = %q, want %q", tt.description, result, tt.expected)
			}
		})
	}
}
