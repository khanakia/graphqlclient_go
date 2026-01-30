# GraphQL Client SDK Builder - Implementation Plan

## Overview
Build a CLI tool that generates type-safe Go client SDKs from GraphQL SDL files. The generated SDK will be publishable as a standalone Go module.

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────────┐
│   SDL File      │────▶│   gqlsdk CLI     │────▶│  Generated Go SDK   │
│ (schema.graphql)│     │   (generator)    │     │   (publishable)     │
└─────────────────┘     └──────────────────┘     └─────────────────────┘
```

## Generated SDK Structure

```
generated-sdk/
├── go.mod                 # Standalone module
├── client.go              # GraphQL HTTP client
├── types.go               # GraphQL object types → Go structs
├── enums.go               # Enum types
├── inputs.go              # Input types for mutations
├── scalars.go             # Custom scalar mappings
├── queries.go             # Type-safe query functions
├── mutations.go           # Type-safe mutation functions
└── builder/
    ├── query_builder.go   # Fluent query builder
    └── field_selector.go  # Field selection helpers
```

## Implementation Phases

### Phase 1: Core Generator Infrastructure
**Files to create/modify:** `gqlsdk/`

1. **Config & CLI** (`gqlsdk/cmd/generate.go`)
   - Add `generate` command with flags:
     - `--schema` (SDL file path)
     - `--output` (output directory)
     - `--package` (Go package name)
     - `--module` (Go module path for publishable SDK)

2. **Generator Core** (`gqlsdk/generator/generator.go`)
   - Parse schema using gqlparser
   - Coordinate code generation
   - Write output files

### Phase 2: Type Generation
**Generate Go types from GraphQL types**

1. **Type Mapper** (`gqlsdk/generator/typemapper.go`)
   - Map GraphQL scalars → Go types
   - Handle nullability (NON_NULL → value, nullable → pointer)
   - Support custom scalar overrides

2. **Struct Generator** (`gqlsdk/generator/types.go`)
   - Generate Go structs for OBJECT types
   - Generate enums with constants
   - Generate input structs for INPUT_OBJECT
   - Add JSON tags for serialization

### Phase 3: Operation Generation
**Generate query/mutation functions**

1. **Query Generator** (`gqlsdk/generator/queries.go`)
   - Parse Query type fields
   - Generate function per query
   - Auto-generate GraphQL query strings
   - Create response types

2. **Mutation Generator** (`gqlsdk/generator/mutations.go`)
   - Parse Mutation type fields
   - Generate function per mutation
   - Handle input variables

### Phase 4: Client & Builder
**HTTP client and fluent builders**

1. **Client Template** (`gqlsdk/templates/client.go.tmpl`)
   - HTTP transport with headers support
   - Error handling
   - Request/response handling

2. **Query Builder** (`gqlsdk/templates/builder.go.tmpl`)
   - Fluent API for field selection
   - Variable binding
   - Query string generation

## Type Mapping

| GraphQL Type | Go Type |
|--------------|---------|
| String | string / *string |
| Int | int / *int |
| Float | float64 / *float64 |
| Boolean | bool / *bool |
| ID | string / *string |
| [T] | []T |
| T! | T (non-pointer) |
| T | *T (pointer) |
| Custom Scalar | Configurable |

## Example Output

For this GraphQL:
```graphql
type User {
  id: ID!
  name: String!
  email: String
}

type Query {
  user(id: ID!): User
  users: [User!]!
}
```

Generated Go:
```go
// types.go
type User struct {
    ID    string  `json:"id"`
    Name  string  `json:"name"`
    Email *string `json:"email,omitempty"`
}

// queries.go
func (c *Client) User(ctx context.Context, id string) (*User, error)
func (c *Client) Users(ctx context.Context) ([]User, error)
```

## Key Files to Create

| File | Purpose |
|------|---------|
| `gqlsdk/cmd/root.go` | CLI entry point with cobra |
| `gqlsdk/cmd/generate.go` | Generate command |
| `gqlsdk/generator/generator.go` | Main generator orchestrator |
| `gqlsdk/generator/typemapper.go` | GraphQL → Go type mapping |
| `gqlsdk/generator/types.go` | Generate struct definitions |
| `gqlsdk/generator/queries.go` | Generate query functions |
| `gqlsdk/generator/mutations.go` | Generate mutation functions |
| `gqlsdk/generator/writer.go` | File output handling |
| `gqlsdk/templates/*.tmpl` | Go code templates |

## Verification Plan

1. **Unit Tests**
   - Test type mapping logic
   - Test struct generation
   - Test query string generation

2. **Integration Test**
   - Generate SDK from `schema.graphql`
   - Compile generated code
   - Verify types match schema

3. **Manual Testing**
   ```bash
   # Generate SDK
   go run ./gqlsdk generate --schema schema.graphql --output ./sdk --package myapi

   # Verify compilation
   cd sdk && go build ./...
   ```

## Dependencies to Add

```go
// go.mod additions
github.com/spf13/cobra        // CLI framework
github.com/vektah/gqlparser/v2 // Already present
```

---

## Detailed Builder Pattern Design

### Field Selection Pattern

```go
// Generated code example - Type-safe field selectors
type UserFields struct {
    builder *queryBuilder
    path    string
}

func (f *UserFields) ID() *UserFields {
    f.builder.addField(f.path, "id")
    return f
}

func (f *UserFields) Email() *UserFields {
    f.builder.addField(f.path, "email")
    return f
}

// Nested relations
func (f *UserFields) Chatbots(selector func(*ChatbotFields)) *UserFields {
    nested := &ChatbotFields{builder: f.builder, path: f.path + ".chatbots"}
    selector(nested)
    return f
}
```

### Query Builder Pattern

```go
// Generated query builder
type UsersQueryBuilder struct {
    client  *Client
    first   *int
    after   *string
    where   *UserWhereInput
    orderBy []UserOrder
    fields  *UserConnectionFields
}

func (c *Client) Users() *UsersQueryBuilder {
    return &UsersQueryBuilder{client: c, fields: &UserConnectionFields{...}}
}

func (q *UsersQueryBuilder) First(n int) *UsersQueryBuilder {
    q.first = &n
    return q
}

func (q *UsersQueryBuilder) Where(w *UserWhereInput) *UsersQueryBuilder {
    q.where = w
    return q
}

func (q *UsersQueryBuilder) Select(selector func(*UserConnectionFields)) *UsersQueryBuilder {
    selector(q.fields)
    return q
}

func (q *UsersQueryBuilder) Execute(ctx context.Context) (*UserConnection, error) {
    // Build and execute query
}
```

### Usage Examples

```go
// Basic query
result, err := client.Users().
    First(10).
    Select(func(conn *UserConnectionFields) {
        conn.Nodes(func(u *UserFields) {
            u.ID().FirstName().Email()
        })
    }).
    Execute(ctx)

// With filtering and nested relations
result, err := client.Users().
    First(10).
    Where(client.UserWhere().EmailContains("@example.com")).
    Select(func(conn *UserConnectionFields) {
        conn.
            PageInfo(func(p *PageInfoFields) {
                p.HasNextPage().EndCursor()
            }).
            Nodes(func(u *UserFields) {
                u.ID().FirstName().
                    Chatbots(func(c *ChatbotFields) {
                        c.ID().Name().Active()
                    })
            })
    }).
    Execute(ctx)
```

### Connection/Pagination Support

```go
// Relay-style pagination
type UserConnectionFields struct {
    builder *queryBuilder
}

func (f *UserConnectionFields) TotalCount() *UserConnectionFields
func (f *UserConnectionFields) PageInfo(selector func(*PageInfoFields)) *UserConnectionFields
func (f *UserConnectionFields) Edges(selector func(*UserEdgeFields)) *UserConnectionFields
func (f *UserConnectionFields) Nodes(selector func(*UserFields)) *UserConnectionFields // convenience
```

---

## Generated File Structure

```
generated-sdk/
├── go.mod
├── client.go              # HTTP client with execute()
├── types.go               # User, Chatbot, etc.
├── enums.go               # OrderDirection, UserType, etc.
├── inputs.go              # UserWhereInput, UserOrder, etc.
├── scalars.go             # Time, Cursor custom scalars
├── queries/
│   ├── users.go           # UsersQueryBuilder, UserFields
│   ├── chatbots.go        # ChatbotsQueryBuilder
│   └── ...
├── mutations/
│   ├── create_user.go
│   └── ...
└── builder/
    └── query_builder.go   # Core query string building
```
