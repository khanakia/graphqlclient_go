# gqlsdl

A CLI tool to fetch GraphQL schema from a server via introspection and save it as SDL (Schema Definition Language) format.

## Installation

```bash
# From workspace root
task gqlsdl:build

# Or directly
cd gqlsdl && go build -o gqlsdl .
```

## Usage

```bash
gqlsdl -url <graphql-endpoint> [options]
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-url` | GraphQL endpoint URL (required) | - |
| `-output` | Output file path | `schema.graphql` |
| `-auth` | Authorization header value | - |
| `-referer` | Referer header value | - |
| `-origin` | Origin header value | - |

## Examples

### Basic Usage

```bash
go run . -url "https://api.example.com/graphql"
```

### With Custom Output File

```bash
go run . -url "https://api.example.com/graphql" -output my-schema.graphql
```

### With Authorization Header

```bash
go run . -url "https://api.example.com/graphql" -auth "Bearer your-token"
```

### With Referer and Origin Headers

```bash
go run . -url "http://localhost:2310/sa/query_playground" \
  -referer "http://localhost:2310/sa/gql?pkey=1234" \
  -origin "http://localhost:2310"
```

### Full Example

```bash
go run . \
  -url "http://localhost:2310/sa/query_playground" \
  -referer "http://localhost:2310/sa/gql?pkey=1234" \
  -origin "http://localhost:2310" \
  -output schema.graphql
```

## Output

The tool generates a `schema.graphql` file (or custom filename) containing the complete GraphQL schema in SDL format:

```graphql
type Query {
  users: [User!]!
  user(id: ID!): User
}

type User {
  id: ID!
  name: String!
  email: String!
}
```

## Package API

The `schema` package can also be used programmatically:

```go
import "gqlsdl/schema"

// Fetch schema from endpoint
opts := &schema.FetchOptions{
    Headers: map[string]string{
        "Authorization": "Bearer token",
        "Referer":       "http://example.com",
    },
}

introspectionSchema, err := schema.FetchSchema("https://api.example.com/graphql", opts)
if err != nil {
    log.Fatal(err)
}

// Convert to SDL
sdl := schema.ConvertToSDL(introspectionSchema)

// Save to file
err = schema.SaveToFile(sdl, "schema.graphql")
```
