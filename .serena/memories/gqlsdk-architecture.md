# gqlsdk - Quick Reference

## What is it?
CLI tool that generates type-safe Go client SDKs from GraphQL schemas with builder pattern and field selection.

## Key Files

### Generator Core
- `generator/generator.go` - Main orchestrator, `Generate()` method
- `generator/builder.go` - Query builder generation (most complex)
- `generator/typemapper.go` - GraphQL → Go type conversion
- `generator/types.go` - Struct/enum/input generation
- `generator/writer.go` - File writing with formatting

### Templates (`generator/templates/`)
- `builder.go.tmpl` - Base infrastructure (FieldSelection, BaseBuilder, QueryRoot, MutationRoot)
- `field_selector.go.tmpl` - Field selector types (e.g., ChatbotFields)
- `operation_builder.go.tmpl` - Query/mutation builders
- `root_methods.go.tmpl` - Entry point methods

## Generated File Structure
```
output/
├── builder.go           # Base + root methods
├── fields_<entity>.go   # Field selectors grouped (Chatbot + ChatbotConnection + ChatbotEdge)
├── query_<name>.go      # One per query
├── mutation_<name>.go   # One per mutation
├── client.go, types.go, enums.go, inputs.go, scalars.go
```

## Key Functions

### builder.go
- `GenerateBuilderFiles(packageName)` → `[]GeneratedFile` - Main entry
- `generateFieldSelectorFiles()` - Groups types by entity
- `generateOperationFile()` - Creates query/mutation file
- `getBaseEntityName(typeName)` - "UserConnection" → "User"
- `toSnakeCase(s)` - "HTMLToMarkdown" → "html_to_markdown"

### Naming
- Query: `ChatbotsBuilder` in `query_chatbots.go`
- Mutation: `CreateChatbotMutationBuilder` in `mutation_create_chatbot.go` (MutationBuilder suffix avoids collision)
- Fields: `ChatbotFields`, `ChatbotConnectionFields`, `ChatbotEdgeFields` all in `fields_chatbot.go`

## Data Structures
```go
type GeneratedFile struct {
    Filename string
    Content  string
}

type FieldSelectorData struct {
    PackageName, TypeName, SelectorName string
    Fields []FieldData
}

type OperationBuilderData struct {
    PackageName, BuilderName, OpType, FieldName string
    Arguments []ArgumentData
    HasSelect bool
    SelectorName, ReturnType, ZeroValue string
}
```

## Full documentation
See `/Volumes/D/khanakia/Downloads/graphqlclient_go/gqlsdk/ARCHITECTURE.md`
