package generator

import (
	"bytes"
	"embed"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/vektah/gqlparser/v2/ast"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// BuilderGenerator generates the query builder pattern code
type BuilderGenerator struct {
	schema     *ast.Schema
	typeMapper *TypeMapper
	templates  *template.Template
}

// NewBuilderGenerator creates a new BuilderGenerator
func NewBuilderGenerator(schema *ast.Schema, tm *TypeMapper) *BuilderGenerator {
	tmpl := template.Must(template.ParseFS(templatesFS, "templates/*.tmpl"))
	return &BuilderGenerator{
		schema:     schema,
		typeMapper: tm,
		templates:  tmpl,
	}
}

// FieldData holds data for a field in template
type FieldData struct {
	FieldName      string
	MethodName     string
	IsObject       bool
	NestedSelector string
}

// FieldSelectorData holds data for field selector template
type FieldSelectorData struct {
	PackageName  string
	TypeName     string
	SelectorName string
	Fields       []FieldData
	RootPkg      string // Package alias for root package (empty if same package)
}

// ArgumentData holds data for an argument
type ArgumentData struct {
	ArgName     string
	MethodName  string
	GoType      string
	GraphQLType string
}

// OperationBuilderData holds data for operation builder template
type OperationBuilderData struct {
	PackageName  string
	BuilderName  string
	OpType       string
	FieldName    string
	Arguments    []ArgumentData
	HasSelect    bool
	SelectorName string
	ReturnType   string
	ZeroValue    string
	RootPkg      string // Package alias for root package (empty if same package)
	FieldsPkg    string // Package alias for fields package (empty if same package)
}

// RootMethodData holds data for root method
type RootMethodData struct {
	MethodName  string
	FieldName   string
	BuilderName string
	OpName      string
}

// RootMethodsData holds data for root methods template
type RootMethodsData struct {
	QueryMethods    []RootMethodData
	MutationMethods []RootMethodData
}

// GeneratedFile represents a generated file
type GeneratedFile struct {
	Filename string
	Content  string
}

// GenerateBuilderFiles generates the query builder infrastructure as separate files
// Files are organized into subdirectories: fields/, queries/, mutations/
// Each subdirectory has its own package to avoid circular dependencies
func (bg *BuilderGenerator) GenerateBuilderFiles(packageName string) []GeneratedFile {
	var files []GeneratedFile

	// Extract local package name from import path (e.g., "testsdk/api" -> "api")
	localPkgName := packageName
	if idx := strings.LastIndex(packageName, "/"); idx != -1 {
		localPkgName = packageName[idx+1:]
	}

	// Generate base builder code (builder.go) - stays in root package
	// Does NOT include QueryRoot/MutationRoot (those go in queries/mutations packages)
	var baseBuf bytes.Buffer
	bg.templates.ExecuteTemplate(&baseBuf, "builder.go.tmpl", map[string]string{
		"PackageName": localPkgName,
	})
	files = append(files, GeneratedFile{
		Filename: "builder.go",
		Content:  baseBuf.String(),
	})

	// Generate field selector files in fields/ subdirectory (package: fields)
	fieldFiles := bg.generateFieldSelectorFiles(packageName)
	files = append(files, fieldFiles...)

	// Generate query files in queries/ subdirectory (package: queries)
	if bg.schema.Query != nil {
		// Generate QueryRoot in queries/root.go
		files = append(files, bg.generateQueryRoot(packageName))

		for _, field := range bg.schema.Query.Fields {
			if strings.HasPrefix(field.Name, "__") {
				continue
			}
			file := bg.generateOperationFile(packageName, "query", field)
			files = append(files, file)
		}
	}

	// Generate mutation files in mutations/ subdirectory (package: mutations)
	if bg.schema.Mutation != nil {
		// Generate MutationRoot in mutations/root.go
		files = append(files, bg.generateMutationRoot(packageName))

		for _, field := range bg.schema.Mutation.Fields {
			if strings.HasPrefix(field.Name, "__") {
				continue
			}
			file := bg.generateOperationFile(packageName, "mutation", field)
			files = append(files, file)
		}
	}

	return files
}

// generateQueryRoot generates the QueryRoot type in queries/root.go
func (bg *BuilderGenerator) generateQueryRoot(packageName string) GeneratedFile {
	var sb strings.Builder

	// Extract root package alias (last part of import path)
	rootPkgAlias := packageName
	if idx := strings.LastIndex(packageName, "/"); idx != -1 {
		rootPkgAlias = packageName[idx+1:]
	}

	sb.WriteString("package queries\n\n")
	fmt.Fprintf(&sb, "import \"%s\"\n\n", packageName)

	sb.WriteString("// QueryRoot is the entry point for queries\n")
	sb.WriteString("type QueryRoot struct {\n")
	fmt.Fprintf(&sb, "\tclient *%s.Client\n", rootPkgAlias)
	sb.WriteString("}\n\n")

	sb.WriteString("// NewQueryRoot creates a new QueryRoot\n")
	fmt.Fprintf(&sb, "func NewQueryRoot(client *%s.Client) *QueryRoot {\n", rootPkgAlias)
	sb.WriteString("\treturn &QueryRoot{client: client}\n")
	sb.WriteString("}\n")

	// Generate methods for each query
	for _, field := range bg.schema.Query.Fields {
		if strings.HasPrefix(field.Name, "__") {
			continue
		}
		methodName := ToPascalCase(field.Name)
		builderName := methodName + "Builder"

		fmt.Fprintf(&sb, "\n// %s creates a new %s\n", methodName, builderName)
		fmt.Fprintf(&sb, "func (q *QueryRoot) %s() *%s {\n", methodName, builderName)
		fmt.Fprintf(&sb, "\treturn &%s{\n", builderName)
		fmt.Fprintf(&sb, "\t\tBaseBuilder: %s.NewBaseBuilder(q.client, \"query\", \"%s\", \"%s\"),\n", rootPkgAlias, field.Name, methodName)
		sb.WriteString("\t}\n")
		sb.WriteString("}\n")
	}

	return GeneratedFile{
		Filename: "queries/root.go",
		Content:  sb.String(),
	}
}

// generateMutationRoot generates the MutationRoot type in mutations/root.go
func (bg *BuilderGenerator) generateMutationRoot(packageName string) GeneratedFile {
	var sb strings.Builder

	// Extract root package alias (last part of import path)
	rootPkgAlias := packageName
	if idx := strings.LastIndex(packageName, "/"); idx != -1 {
		rootPkgAlias = packageName[idx+1:]
	}

	sb.WriteString("package mutations\n\n")
	fmt.Fprintf(&sb, "import \"%s\"\n\n", packageName)

	sb.WriteString("// MutationRoot is the entry point for mutations\n")
	sb.WriteString("type MutationRoot struct {\n")
	fmt.Fprintf(&sb, "\tclient *%s.Client\n", rootPkgAlias)
	sb.WriteString("}\n\n")

	sb.WriteString("// NewMutationRoot creates a new MutationRoot\n")
	fmt.Fprintf(&sb, "func NewMutationRoot(client *%s.Client) *MutationRoot {\n", rootPkgAlias)
	sb.WriteString("\treturn &MutationRoot{client: client}\n")
	sb.WriteString("}\n")

	// Generate methods for each mutation
	for _, field := range bg.schema.Mutation.Fields {
		if strings.HasPrefix(field.Name, "__") {
			continue
		}
		methodName := ToPascalCase(field.Name)
		builderName := methodName + "MutationBuilder"

		fmt.Fprintf(&sb, "\n// %s creates a new %s\n", methodName, builderName)
		fmt.Fprintf(&sb, "func (m *MutationRoot) %s() *%s {\n", methodName, builderName)
		fmt.Fprintf(&sb, "\treturn &%s{\n", builderName)
		fmt.Fprintf(&sb, "\t\tBaseBuilder: %s.NewBaseBuilder(m.client, \"mutation\", \"%s\", \"%s\"),\n", rootPkgAlias, field.Name, methodName)
		sb.WriteString("\t}\n")
		sb.WriteString("}\n")
	}

	return GeneratedFile{
		Filename: "mutations/root.go",
		Content:  sb.String(),
	}
}

// GenerateBuilder generates all builder code in a single file (for backward compatibility)
func (bg *BuilderGenerator) GenerateBuilder(packageName string) string {
	var sb strings.Builder

	// Generate base builder code from template
	var baseBuf bytes.Buffer
	bg.templates.ExecuteTemplate(&baseBuf, "builder.go.tmpl", map[string]string{
		"PackageName": packageName,
	})
	sb.WriteString(baseBuf.String())

	// Generate field selectors for each object type
	sb.WriteString(bg.generateFieldSelectors())

	// Generate query builders
	if bg.schema.Query != nil {
		sb.WriteString("\n// --- Query Builders ---\n")
		sb.WriteString(bg.generateOperationBuilders("query", bg.schema.Query))
	}

	// Generate mutation builders
	if bg.schema.Mutation != nil {
		sb.WriteString("\n// --- Mutation Builders ---\n")
		sb.WriteString(bg.generateOperationBuilders("mutation", bg.schema.Mutation))
	}

	// Generate root methods
	sb.WriteString(bg.generateRootMethods())

	return sb.String()
}

// generateFieldSelectorFiles generates separate files for field selectors grouped by entity
func (bg *BuilderGenerator) generateFieldSelectorFiles(packageName string) []GeneratedFile {
	// Group types by base entity name
	// e.g., User, UserConnection, UserEdge -> "User"
	entityGroups := make(map[string][]*ast.Definition)

	typeNames := getSortedTypeNames(bg.schema)
	for _, name := range typeNames {
		def := bg.schema.Types[name]
		if def.BuiltIn || strings.HasPrefix(name, "__") {
			continue
		}

		if def.Kind != ast.Object && def.Kind != ast.Interface {
			continue
		}

		if name == "Query" || name == "Mutation" || name == "Subscription" {
			continue
		}

		// Determine base entity name
		baseName := getBaseEntityName(name)
		entityGroups[baseName] = append(entityGroups[baseName], def)
	}

	// Generate a file for each entity group
	var files []GeneratedFile
	sortedEntities := make([]string, 0, len(entityGroups))
	for entity := range entityGroups {
		sortedEntities = append(sortedEntities, entity)
	}
	sort.Strings(sortedEntities)

	for _, entity := range sortedEntities {
		defs := entityGroups[entity]
		file := bg.generateEntityFieldsFile(packageName, entity, defs)
		files = append(files, file)
	}

	return files
}

// getBaseEntityName extracts the base entity name from a type name
// e.g., "UserConnection" -> "User", "UserEdge" -> "User", "User" -> "User"
func getBaseEntityName(typeName string) string {
	suffixes := []string{"Connection", "Edge", "Payload", "Input", "Response"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(typeName, suffix) {
			return strings.TrimSuffix(typeName, suffix)
		}
	}
	return typeName
}

// generateEntityFieldsFile generates a file containing all field selectors for an entity
// Files are placed in the fields/ subdirectory with package "fields"
func (bg *BuilderGenerator) generateEntityFieldsFile(packageName, entity string, defs []*ast.Definition) GeneratedFile {
	var sb strings.Builder

	// Extract root package alias (last part of import path)
	rootPkgAlias := packageName
	if idx := strings.LastIndex(packageName, "/"); idx != -1 {
		rootPkgAlias = packageName[idx+1:]
	}

	// Package declaration
	sb.WriteString("package fields\n\n")

	// Import root package for FieldSelection type
	sb.WriteString(fmt.Sprintf("import \"%s\"\n\n", packageName))

	// Sort definitions for consistent output (main entity first, then Connection, then Edge)
	sort.Slice(defs, func(i, j int) bool {
		return getTypeSortOrder(defs[i].Name) < getTypeSortOrder(defs[j].Name)
	})

	// Generate field selectors for each type in this entity group
	for _, def := range defs {
		data := bg.buildFieldSelectorData(def)
		data.PackageName = "fields"
		data.RootPkg = rootPkgAlias
		var buf bytes.Buffer
		bg.templates.ExecuteTemplate(&buf, "field_selector.go.tmpl", data)
		sb.WriteString(buf.String())
	}

	// Filename in fields/ subdirectory
	filename := fmt.Sprintf("fields/field_%s.go", toSnakeCase(entity))

	return GeneratedFile{
		Filename: filename,
		Content:  sb.String(),
	}
}

// getTypeSortOrder returns a sort order for type names
// Main entity = 0, Connection = 1, Edge = 2, others = 3
func getTypeSortOrder(typeName string) int {
	if strings.HasSuffix(typeName, "Connection") {
		return 1
	}
	if strings.HasSuffix(typeName, "Edge") {
		return 2
	}
	if strings.HasSuffix(typeName, "Payload") || strings.HasSuffix(typeName, "Response") {
		return 3
	}
	return 0 // Main entity first
}

// generateFieldSelectorsFile generates the types_fields.go file (kept for backward compatibility)
func (bg *BuilderGenerator) generateFieldSelectorsFile(packageName string) string {
	var sb strings.Builder

	// Package declaration (header added by WriteFileWithHeader)
	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// Generate field selectors for each object type
	typeNames := getSortedTypeNames(bg.schema)

	for _, name := range typeNames {
		def := bg.schema.Types[name]
		if def.BuiltIn || strings.HasPrefix(name, "__") {
			continue
		}

		if def.Kind != ast.Object && def.Kind != ast.Interface {
			continue
		}

		if name == "Query" || name == "Mutation" || name == "Subscription" {
			continue
		}

		data := bg.buildFieldSelectorData(def)
		data.PackageName = packageName
		var buf bytes.Buffer
		bg.templates.ExecuteTemplate(&buf, "field_selector.go.tmpl", data)
		sb.WriteString(buf.String())
	}

	return sb.String()
}

// prefixCustomType adds a package prefix to custom types (non-primitive types)
// e.g., "AiModel" -> "api.AiModel", "*AiModel" -> "*api.AiModel", "[]AiModel" -> "[]api.AiModel"
// Primitive types like string, int, bool are returned unchanged
// Zero values like "", 0, false, nil are also returned unchanged
func prefixCustomType(goType, pkgAlias string) string {
	// List of primitive types that don't need prefixing
	primitives := map[string]bool{
		"string": true, "int": true, "int32": true, "int64": true,
		"float32": true, "float64": true, "bool": true, "any": true,
		"interface{}": true,
	}

	// List of zero values that don't need prefixing
	zeroValues := map[string]bool{
		`""`: true, "0": true, "false": true, "nil": true,
	}

	// Check if it's a zero value literal
	if zeroValues[goType] {
		return goType
	}

	// Handle pointer types
	if strings.HasPrefix(goType, "*") {
		inner := goType[1:]
		if primitives[inner] {
			return goType
		}
		// Don't prefix map types
		if strings.HasPrefix(inner, "map[") {
			return goType
		}
		return "*" + pkgAlias + "." + inner
	}

	// Handle slice types
	if strings.HasPrefix(goType, "[]") {
		inner := goType[2:]
		// Handle slice of pointers
		if strings.HasPrefix(inner, "*") {
			innerType := inner[1:]
			if primitives[innerType] {
				return goType
			}
			return "[]*" + pkgAlias + "." + innerType
		}
		if primitives[inner] {
			return goType
		}
		return "[]" + pkgAlias + "." + inner
	}

	// Handle map types (simplified - assumes map[string]Type pattern)
	if strings.HasPrefix(goType, "map[") {
		return goType // Keep as-is for now, maps are usually map[string]interface{}
	}

	// Simple type
	if primitives[goType] {
		return goType
	}
	return pkgAlias + "." + goType
}

// generateOperationFile generates a single query or mutation file
// Files are placed in queries/ or mutations/ subdirectory with appropriate package
func (bg *BuilderGenerator) generateOperationFile(packageName, opType string, field *ast.FieldDefinition) GeneratedFile {
	var sb strings.Builder

	// Extract root package alias (last part of import path)
	rootPkgAlias := packageName
	if idx := strings.LastIndex(packageName, "/"); idx != -1 {
		rootPkgAlias = packageName[idx+1:]
	}

	// Determine subdirectory and package name
	var subDir, pkgName string
	if opType == "query" {
		subDir = "queries"
		pkgName = "queries"
	} else {
		subDir = "mutations"
		pkgName = "mutations"
	}

	// Generate the builder data first to determine if we need fields import
	data := bg.buildOperationBuilderData(opType, field)

	// Package declaration
	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Imports - only include fields package if operation has a Select method
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n\n")
	sb.WriteString(fmt.Sprintf("\t\"%s\"\n", packageName))
	if data.HasSelect {
		sb.WriteString(fmt.Sprintf("\t\"%s/fields\"\n", packageName))
	}
	sb.WriteString(")\n\n")
	data.PackageName = pkgName
	data.RootPkg = rootPkgAlias
	data.FieldsPkg = "fields"

	// Prefix custom types with root package alias (for subpackage usage)
	data.ReturnType = prefixCustomType(data.ReturnType, rootPkgAlias)
	data.ZeroValue = prefixCustomType(data.ZeroValue, rootPkgAlias)
	for i := range data.Arguments {
		data.Arguments[i].GoType = prefixCustomType(data.Arguments[i].GoType, rootPkgAlias)
	}

	var buf bytes.Buffer
	bg.templates.ExecuteTemplate(&buf, "operation_builder.go.tmpl", data)
	sb.WriteString(buf.String())

	// Generate filename in subdirectory
	filename := fmt.Sprintf("%s/%s_%s.go", subDir, opType, toSnakeCase(field.Name))

	return GeneratedFile{
		Filename: filename,
		Content:  sb.String(),
	}
}

// toSnakeCase converts a string to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// generateFieldSelectors generates field selector types for all object types
func (bg *BuilderGenerator) generateFieldSelectors() string {
	var sb strings.Builder

	typeNames := getSortedTypeNames(bg.schema)

	for _, name := range typeNames {
		def := bg.schema.Types[name]
		if def.BuiltIn || strings.HasPrefix(name, "__") {
			continue
		}

		if def.Kind != ast.Object && def.Kind != ast.Interface {
			continue
		}

		if name == "Query" || name == "Mutation" || name == "Subscription" {
			continue
		}

		data := bg.buildFieldSelectorData(def)
		var buf bytes.Buffer
		bg.templates.ExecuteTemplate(&buf, "field_selector.go.tmpl", data)
		sb.WriteString(buf.String())
	}

	return sb.String()
}

// buildFieldSelectorData builds template data for a field selector
func (bg *BuilderGenerator) buildFieldSelectorData(def *ast.Definition) FieldSelectorData {
	data := FieldSelectorData{
		TypeName:     def.Name,
		SelectorName: def.Name + "Fields",
		Fields:       make([]FieldData, 0),
	}

	for _, field := range def.Fields {
		if strings.HasPrefix(field.Name, "__") {
			continue
		}

		baseTypeName := getBaseTypeName(field.Type)
		isObj := isObjectType(bg.schema, field.Type) && baseTypeName != def.Name

		fieldData := FieldData{
			FieldName:  field.Name,
			MethodName: ToPascalCase(field.Name),
			IsObject:   isObj,
		}

		if isObj {
			fieldData.NestedSelector = baseTypeName + "Fields"
		}

		data.Fields = append(data.Fields, fieldData)
	}

	return data
}

// generateOperationBuilders generates builders for query or mutation fields
func (bg *BuilderGenerator) generateOperationBuilders(opType string, def *ast.Definition) string {
	var sb strings.Builder

	for _, field := range def.Fields {
		if strings.HasPrefix(field.Name, "__") {
			continue
		}

		data := bg.buildOperationBuilderData(opType, field)
		var buf bytes.Buffer
		bg.templates.ExecuteTemplate(&buf, "operation_builder.go.tmpl", data)
		sb.WriteString(buf.String())
	}

	return sb.String()
}

// buildOperationBuilderData builds template data for an operation builder
func (bg *BuilderGenerator) buildOperationBuilderData(opType string, field *ast.FieldDefinition) OperationBuilderData {
	returnTypeName := getBaseTypeName(field.Type)
	returnType := bg.typeMapper.GraphQLToGoType(field.Type)
	isObjReturn := isObjectType(bg.schema, field.Type)

	// Prefix mutation builders to avoid name collisions
	builderName := ToPascalCase(field.Name) + "Builder"
	if opType == "mutation" {
		builderName = ToPascalCase(field.Name) + "MutationBuilder"
	}

	data := OperationBuilderData{
		BuilderName: builderName,
		OpType:      opType,
		FieldName:   field.Name,
		Arguments:   make([]ArgumentData, 0),
		HasSelect:   isObjReturn,
		ReturnType:  returnType,
		ZeroValue:   getZeroValue(returnType),
	}

	if isObjReturn {
		data.SelectorName = returnTypeName + "Fields"
	}

	for _, arg := range field.Arguments {
		argData := ArgumentData{
			ArgName:     arg.Name,
			MethodName:  ToPascalCase(arg.Name),
			GoType:      bg.typeMapper.GraphQLToGoType(arg.Type),
			GraphQLType: formatGraphQLType(arg.Type),
		}
		data.Arguments = append(data.Arguments, argData)
	}

	return data
}

// generateRootMethods generates QueryRoot and MutationRoot methods
func (bg *BuilderGenerator) generateRootMethods() string {
	data := RootMethodsData{
		QueryMethods:    make([]RootMethodData, 0),
		MutationMethods: make([]RootMethodData, 0),
	}

	if bg.schema.Query != nil {
		for _, field := range bg.schema.Query.Fields {
			if strings.HasPrefix(field.Name, "__") {
				continue
			}
			methodName := ToPascalCase(field.Name)
			data.QueryMethods = append(data.QueryMethods, RootMethodData{
				MethodName:  methodName,
				FieldName:   field.Name,
				BuilderName: methodName + "Builder",
				OpName:      methodName,
			})
		}
	}

	if bg.schema.Mutation != nil {
		for _, field := range bg.schema.Mutation.Fields {
			if strings.HasPrefix(field.Name, "__") {
				continue
			}
			methodName := ToPascalCase(field.Name)
			data.MutationMethods = append(data.MutationMethods, RootMethodData{
				MethodName:  methodName,
				FieldName:   field.Name,
				BuilderName: methodName + "MutationBuilder",
				OpName:      methodName,
			})
		}
	}

	var buf bytes.Buffer
	bg.templates.ExecuteTemplate(&buf, "root_methods.go.tmpl", data)
	return buf.String()
}

// getSortedTypeNames returns sorted type names from schema
func getSortedTypeNames(schema *ast.Schema) []string {
	names := make([]string, 0, len(schema.Types))
	for name := range schema.Types {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// getZeroValue returns the zero value for a Go type
func getZeroValue(goType string) string {
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
