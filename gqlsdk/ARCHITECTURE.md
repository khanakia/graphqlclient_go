# gqlsdk - GraphQL Client SDK Generator

## Overview

`gqlsdk` is a CLI tool that generates type-safe Go client SDKs from GraphQL SDL schema files. It creates a builder-pattern API with field selection capabilities.

## Project Structure

```
gqlsdk/
├── main.go                 # CLI entry point
├── cmd/
│   ├── root.go            # Cobra root command
│   └── generate.go        # Generate command with flags
├── generator/
│   ├── generator.go       # Main orchestrator
│   ├── config.go          # Configuration struct
│   ├── typemapper.go      # GraphQL to Go type mapping
│   ├── types.go           # Type/enum/input generation
│   ├── client.go          # HTTP client generation
│   ├── builder.go         # Query builder generation (main logic)
│   ├── writer.go          # File writing utilities
│   ├── errors.go          # Error types
│   └── templates/         # Go templates for code generation
│       ├── builder.go.tmpl
│       ├── field_selector.go.tmpl
│       ├── operation_builder.go.tmpl
│       └── root_methods.go.tmpl
└── ARCHITECTURE.md        # This file
```

## Generated SDK Structure

The SDK is organized into separate packages to avoid circular dependencies and improve code navigation:

```
output/
├── go.mod                 # Module definition
├── client.go              # HTTP client with Execute method
├── scalars.go             # Custom scalar type definitions
├── types.go               # Go structs from GraphQL types
├── enums.go               # Go enum types
├── inputs.go              # Input types for mutations
├── builder.go             # Base builder infrastructure (FieldSelection, BaseBuilder)
│
├── fields/                # Field selector types (package: fields)
│   ├── field_chatbot.go   # ChatbotFields, ChatbotConnectionFields, ChatbotEdgeFields
│   ├── field_user.go      # UserFields, UserConnectionFields, UserEdgeFields
│   └── ...                # One file per entity (grouped by base name)
│
├── queries/               # Query builders (package: queries)
│   ├── root.go            # QueryRoot type with factory methods
│   ├── query_chatbots.go  # ChatbotsBuilder
│   ├── query_users.go     # UsersBuilder
│   └── ...                # One file per query operation
│
└── mutations/             # Mutation builders (package: mutations)
    ├── root.go            # MutationRoot type with factory methods
    ├── mutation_create_chatbot.go  # CreateChatbotMutationBuilder
    └── ...                # One file per mutation operation
```

### Package Dependencies

```
┌─────────────────────────────────────────────────────────────┐
│                    User Code (main.go)                       │
└─────────────────────────────────────────────────────────────┘
                              │
           ┌──────────────────┼──────────────────┐
           ▼                  ▼                  ▼
    ┌────────────┐     ┌───────────┐     ┌─────────────┐
    │   api      │     │  queries  │     │  mutations  │
    │ (root pkg) │     │           │     │             │
    └────────────┘     └───────────┘     └─────────────┘
           │                  │                  │
           │                  └────────┬─────────┘
           │                           │
           │                           ▼
           │                    ┌───────────┐
           └───────────────────►│  fields   │
                                └───────────┘
```

- **api (root package)**: Contains `Client`, `FieldSelection`, `BaseBuilder`, types, inputs, enums
- **fields**: Contains field selector types, imports only `api`
- **queries**: Contains `QueryRoot` and query builders, imports `api` and `fields`
- **mutations**: Contains `MutationRoot` and mutation builders, imports `api` and `fields`

This structure avoids circular dependencies because:
- `fields` only depends on `api` (for `FieldSelection`)
- `queries`/`mutations` depend on `api` (for `BaseBuilder`, `Client`) and `fields` (for selectors)
- `api` does not depend on subpackages

## Core Components

### 1. Generator (`generator/generator.go`)

Main orchestrator that coordinates the generation process.

```go
type Generator struct {
    config     *Config
    schema     *ast.Schema      // Parsed GraphQL schema
    typeMapper *TypeMapper
    writer     *Writer
}

func (g *Generator) Generate() error {
    // 1. Generate go.mod
    // 2. Generate scalars.go
    // 3. Generate types.go
    // 4. Generate enums.go
    // 5. Generate inputs.go
    // 6. Generate client.go
    // 7. Generate builder files:
    //    - builder.go (base infrastructure)
    //    - fields/*.go (field selectors)
    //    - queries/*.go (query builders + QueryRoot)
    //    - mutations/*.go (mutation builders + MutationRoot)
}
```

### 2. TypeMapper (`generator/typemapper.go`)

Maps GraphQL types to Go types.

```go
// Key mappings:
// String  -> string
// Int     -> int
// Float   -> float64
// Boolean -> bool
// ID      -> string
// Time    -> time.Time
// JSON    -> json.RawMessage
//
// Nullable types become pointers: String -> *string
// Lists become slices: [String!]! -> []string
```

Key functions:
- `GraphQLToGoType(t *ast.Type) string` - Convert GraphQL type to Go type
- `ToPascalCase(s string) string` - Convert to PascalCase for Go exports
- `ToCamelCase(s string) string` - Convert to camelCase for JSON tags

### 3. TypeGenerator (`generator/types.go`)

Generates Go structs, enums, and inputs from GraphQL schema.

```go
type TypeGenerator struct {
    schema     *ast.Schema
    typeMapper *TypeMapper
}

// Key methods:
func (tg *TypeGenerator) GenerateTypes(packageName string) string    // Object types
func (tg *TypeGenerator) GenerateEnums(packageName string) string    // Enum types
func (tg *TypeGenerator) GenerateInputs(packageName string) string   // Input types
func (tg *TypeGenerator) GenerateScalars(packageName string) string  // Custom scalars
```

#### Conditional Imports in GenerateScalars

The `GenerateScalars` function generates type aliases for custom GraphQL scalars. It uses a **two-pass approach** to avoid generating unused imports (which would cause Go compilation errors):

**Pass 1 - Scan schema for required imports:**
```go
needsTime := false
needsJSON := false

for _, name := range typeNames {
    def := tg.schema.Types[name]
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
```

**Pass 2 - Generate imports only if needed:**
```go
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
```

**Known Scalar Mappings:**
| GraphQL Scalar | Go Type | Import Required |
|----------------|---------|-----------------|
| Time, DateTime, Date | `time.Time` | `time` |
| JSON | `json.RawMessage` | `encoding/json` |
| Cursor, UUID, Password, Upload | `string` | none |
| Map | `map[string]interface{}` | none |
| Any | `interface{}` | none |
| Unknown scalars | `string` | none |

This ensures the generated `scalars.go` file only imports packages that are actually used by the scalars in the schema.

### 4. BuilderGenerator (`generator/builder.go`)

Generates the query builder pattern code. This is the most complex component.

```go
type BuilderGenerator struct {
    schema     *ast.Schema
    typeMapper *TypeMapper
    templates  *template.Template
}

// Key methods:
func (bg *BuilderGenerator) GenerateBuilderFiles(packageName string) []GeneratedFile
func (bg *BuilderGenerator) generateFieldSelectorFiles(packageName string) []GeneratedFile
func (bg *BuilderGenerator) generateEntityFieldsFile(packageName, entity string, defs []*ast.Definition) GeneratedFile
func (bg *BuilderGenerator) generateQueryRoot(packageName string) GeneratedFile
func (bg *BuilderGenerator) generateMutationRoot(packageName string) GeneratedFile
func (bg *BuilderGenerator) generateOperationFile(packageName, opType string, field *ast.FieldDefinition) GeneratedFile
```

#### Key Data Structures

```go
// GeneratedFile represents a file to be written
type GeneratedFile struct {
    Filename string
    Content  string
}

// FieldSelectorData - template data for field selectors
type FieldSelectorData struct {
    PackageName  string
    TypeName     string
    SelectorName string      // e.g., "ChatbotFields"
    Fields       []FieldData
    RootPkg      string      // Package alias for root package (e.g., "api")
}

// FieldData - data for a single field
type FieldData struct {
    FieldName      string    // GraphQL field name
    MethodName     string    // Go method name (PascalCase)
    IsObject       bool      // Whether field returns an object type
    NestedSelector string    // e.g., "UserFields" for nested selection
}

// OperationBuilderData - template data for query/mutation builders
type OperationBuilderData struct {
    PackageName  string
    BuilderName  string      // e.g., "ChatbotsBuilder" or "CreateChatbotMutationBuilder"
    OpType       string      // "query" or "mutation"
    FieldName    string      // GraphQL field name
    Arguments    []ArgumentData
    HasSelect    bool        // Whether return type is an object
    SelectorName string      // e.g., "ChatbotConnectionFields"
    ReturnType   string      // Go return type (prefixed with package, e.g., "api.Chatbot")
    ZeroValue    string      // Zero value for error returns
    RootPkg      string      // Package alias for root package (e.g., "api")
    FieldsPkg    string      // Package alias for fields package (e.g., "fields")
}

// ArgumentData - data for operation arguments
type ArgumentData struct {
    ArgName     string    // GraphQL argument name
    MethodName  string    // Go method name
    GoType      string    // Go type (prefixed with package for custom types)
    GraphQLType string    // GraphQL type string for query building
}
```

### 5. Templates (`generator/templates/`)

#### `builder.go.tmpl`
Base infrastructure (in root package):
- `FieldSelection` - Tracks selected fields
- `NewFieldSelection()` - Factory function
- `BaseBuilder` - Base for all operation builders
- `NewBaseBuilder()` - Factory function
- `GetClient()`, `GetSelection()`, `BuildQuery()`, etc.

**Note:** `QueryRoot` and `MutationRoot` are NOT in this template - they are generated separately in their respective subpackages.

#### `field_selector.go.tmpl`
Generates field selector types (in `fields/` package):
```go
package fields

import "testsdk/api"

type ChatbotFields struct {
    selection *api.FieldSelection
}

func NewChatbotFields(selection *api.FieldSelection) *ChatbotFields {
    return &ChatbotFields{selection: selection}
}

func (f *ChatbotFields) ID() *ChatbotFields {
    f.selection.AddField("id")
    return f
}

func (f *ChatbotFields) Users(selector func(*UserFields)) *ChatbotFields {
    child := api.NewFieldSelection()
    selector(NewUserFields(child))
    f.selection.AddChild("users", child)
    return f
}
```

#### `operation_builder.go.tmpl`
Generates operation builders (in `queries/` or `mutations/` package):
```go
package queries

import (
    "context"
    "testsdk/api"
    "testsdk/api/fields"
)

type ChatbotsBuilder struct {
    *api.BaseBuilder
}

func (b *ChatbotsBuilder) First(v *int) *ChatbotsBuilder {
    b.SetArg("first", v, "Int")
    return b
}

func (b *ChatbotsBuilder) Select(selector func(*fields.ChatbotConnectionFields)) *ChatbotsBuilder {
    selector(fields.NewChatbotConnectionFields(b.GetSelection()))
    return b
}

func (b *ChatbotsBuilder) Execute(ctx context.Context) (api.ChatbotConnection, error) {
    // Build and execute query
}
```

#### `root_methods.go.tmpl`
Used for backward compatibility (single-file mode). Not used in subdirectory mode.

### 6. Writer (`generator/writer.go`)

Handles file writing with Go formatting.

#### Writer

The main file writer that handles disk operations:

```go
type Writer struct {
    outputDir string
}

// Key methods:
func (w *Writer) WriteFileWithHeader(filename, content string) error  // Adds "// Code generated" header
func (w *Writer) WriteFile(filename, content string) error            // Writes and formats Go code
func (w *Writer) WriteGoMod(modulePath, packageName string) error     // Generates go.mod
func (w *Writer) EnsureDir() error                                    // Creates output directory
func (w *Writer) Clean() error                                        // Removes all files in output dir
```

**Features:**
- Automatically creates subdirectories when filename contains `/` (e.g., `fields/field_chatbot.go`)
- Formats Go code using `go/format` before writing
- Falls back to unformatted content if formatting fails (for debugging)
- Adds "Code generated" header to prevent manual edits

#### BufferedWriter

A utility for building file content incrementally in memory:

```go
type BufferedWriter struct {
    buffer bytes.Buffer
}

// Key methods:
func (bw *BufferedWriter) Write(p []byte) (n int, err error)      // Write bytes
func (bw *BufferedWriter) WriteString(s string) (n int, err error) // Write string
func (bw *BufferedWriter) String() string                          // Get accumulated content
func (bw *BufferedWriter) Reset()                                  // Clear buffer
```

**Why BufferedWriter is used:**

1. **Incremental Content Building**: When generating complex files (like types.go or builder.go), content is built piece by piece. BufferedWriter collects all fragments before the final write.

2. **Memory Efficiency**: Instead of concatenating strings (which creates many intermediate allocations), BufferedWriter uses a single growing buffer.

3. **Template Integration**: Works well with `text/template` - templates can write directly to the buffer:
   ```go
   bw := NewBufferedWriter()
   template.Execute(bw, data)  // Template writes to buffer
   content := bw.String()       // Get final content
   ```

4. **Implements io.Writer**: Can be used anywhere an `io.Writer` is expected, making it compatible with standard library functions.

5. **Separation of Concerns**: Separates content generation (BufferedWriter) from file I/O (Writer), making code easier to test and maintain.

**Usage Pattern:**
```go
// Build content in memory
bw := NewBufferedWriter()
bw.WriteString("package api\n\n")
bw.WriteString("type Foo struct {\n")
bw.WriteString("    Bar string\n")
bw.WriteString("}\n")

// Write to disk with formatting
writer.WriteFile("types.go", bw.String())
```

## Code Generation Flow

```
1. Parse CLI flags (--schema, --output, --package, --module)
                    ↓
2. Parse GraphQL schema using gqlparser
                    ↓
3. Generate static files:
   - go.mod
   - client.go (HTTP client)
   - scalars.go (custom scalars discovered from schema)
                    ↓
4. Generate type files using TypeGenerator:
   - types.go (object types → Go structs)
   - enums.go (enum types → Go string constants)
   - inputs.go (input types → Go structs)
                    ↓
5. Generate builder files using BuilderGenerator:
   a. builder.go - Base infrastructure (FieldSelection, BaseBuilder)

   b. fields/*.go - Field selectors grouped by entity
      - Groups types by base name (User, UserConnection, UserEdge → "User")
      - Uses getBaseEntityName() to strip suffixes
      - Package: fields
      - Imports: root package for FieldSelection

   c. queries/root.go - QueryRoot type
      - Package: queries
      - Imports: root package for Client, BaseBuilder

   d. queries/query_*.go - One file per query
      - Package: queries
      - Imports: root package + fields package (if HasSelect)

   e. mutations/root.go - MutationRoot type
      - Package: mutations
      - Imports: root package for Client, BaseBuilder

   f. mutations/mutation_*.go - One file per mutation
      - Package: mutations (with "MutationBuilder" suffix to avoid collisions)
      - Imports: root package + fields package (if HasSelect)
                    ↓
6. Write all files using Writer (with go/format)
```

## Key Algorithms

### Entity Grouping (`getBaseEntityName`)
Groups related types by stripping suffixes:
```go
func getBaseEntityName(typeName string) string {
    suffixes := []string{"Connection", "Edge", "Payload", "Input", "Response"}
    for _, suffix := range suffixes {
        if strings.HasSuffix(typeName, suffix) {
            return strings.TrimSuffix(typeName, suffix)
        }
    }
    return typeName
}
// "UserConnection" → "User"
// "UserEdge" → "User"
// "User" → "User"
```

### Type Sorting (`getTypeSortOrder`)
Orders types within an entity file:
```go
// Main entity = 0, Connection = 1, Edge = 2, Payload/Response = 3
```

### Snake Case Conversion (`toSnakeCase`)
For file names:
```go
// "HTMLToMarkdown" → "html_to_markdown"
// "Chatbots" → "chatbots"
```

### Type Prefixing (`prefixCustomType`)
Adds package prefix to custom types for cross-package usage:
```go
func prefixCustomType(goType, pkgAlias string) string
// "AiModel" → "api.AiModel"
// "*AiModel" → "*api.AiModel"
// "[]AiModel" → "[]api.AiModel"
// "string" → "string" (primitives unchanged)
// `""` → `""` (zero values unchanged)
```

## Naming Conventions

| GraphQL | Go Type | Package | File |
|---------|---------|---------|------|
| `Query.chatbots` | `ChatbotsBuilder` | queries | `queries/query_chatbots.go` |
| `Mutation.createChatbot` | `CreateChatbotMutationBuilder` | mutations | `mutations/mutation_create_chatbot.go` |
| `Chatbot` type | `ChatbotFields` | fields | `fields/field_chatbot.go` |
| `ChatbotConnection` | `ChatbotConnectionFields` | fields | `fields/field_chatbot.go` |
| `ChatbotEdge` | `ChatbotEdgeFields` | fields | `fields/field_chatbot.go` |

**Note:** Mutation builders have "MutationBuilder" suffix to avoid name collisions with query builders (e.g., both `Query.ping` and `Mutation.ping` exist).

## CLI Usage

```bash
gqlsdk generate \
  --schema path/to/schema.graphql \
  --output path/to/output \
  --package mymodule/api \
  --module github.com/example/myapi
```

The `--package` flag specifies the import path for the generated package (e.g., `testsdk/api`). The local package name is extracted from the last segment (e.g., `api`).

## Generated SDK Usage

```go
import (
    "context"
    "testsdk/api"
    "testsdk/api/fields"
    "testsdk/api/queries"
    "testsdk/api/mutations"
)

func main() {
    // Create client
    client := api.NewClient("http://localhost:8080/graphql",
        api.WithAuthToken("token"),
    )

    // Create query and mutation roots
    qr := queries.NewQueryRoot(client)
    mr := mutations.NewMutationRoot(client)

    ctx := context.Background()

    // Query with field selection
    result, err := qr.Chatbots().
        First(intPtr(10)).
        Select(func(conn *fields.ChatbotConnectionFields) {
            conn.TotalCount()
            conn.Edges(func(e *fields.ChatbotEdgeFields) {
                e.Cursor()
                e.Node(func(c *fields.ChatbotFields) {
                    c.ID().Name().CreatedAt()
                    // Nested selection (3 levels deep)
                    c.AiModel(func(ai *fields.AiModelFields) {
                        ai.ID().ModelID()
                    })
                })
            })
        }).
        Execute(ctx)

    // Mutation
    result, err := mr.CreateChatbot().
        Input(api.CreateChatbotInput{Name: "My Bot"}).
        Select(func(c *fields.CreateChatbotFields) {
            c.ID().Name()
        }).
        Execute(ctx)
}
```

## Common Modifications

### Adding a new scalar type
1. Edit `generator/typemapper.go` - add to `scalarMappings` map
2. Scalars are also auto-discovered from schema

### Changing file organization
1. Edit `generator/builder.go`:
   - `GenerateBuilderFiles()` - controls which files are generated
   - `generateFieldSelectorFiles()` - groups field types by entity
   - `generateEntityFieldsFile()` - generates field selector file (in fields/ package)
   - `generateQueryRoot()` - generates queries/root.go
   - `generateMutationRoot()` - generates mutations/root.go
   - `generateOperationFile()` - generates query/mutation files

### Modifying generated code structure
1. Edit templates in `generator/templates/`:
   - `builder.go.tmpl` - base infrastructure (FieldSelection, BaseBuilder)
   - `field_selector.go.tmpl` - field selector types
   - `operation_builder.go.tmpl` - operation builders
   - `root_methods.go.tmpl` - (legacy, for single-file mode)

### Adding new features to generated client
1. Edit `generator/client.go` - `GenerateClient()` method

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/vektah/gqlparser/v2` - GraphQL schema parsing
- `go/format` - Go code formatting

## Testing

```bash
cd gqlsdk
go test ./... -v
```

Key test files:
- `generator/generator_test.go` - Integration tests
- `generator/typemapper_test.go` - Type mapping tests
- `generator/types_test.go` - Type generation tests
