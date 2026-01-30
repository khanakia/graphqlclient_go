package clientgen

import (
	"gqlcustom/pkg/typegql"

	"github.com/vektah/gqlparser/v2/ast"
)

type SchemaScalar struct {
	Name        string
	Description string
	// GoType       string
	IsBuiltIn bool
	GoType    string
	// TypeMapEntry typegql.TypeMapEntry
}

type SchemaScalarMap map[string]SchemaScalar

var builtInScalars = map[string]bool{
	"String":  true,
	"Int":     true,
	"Float":   true,
	"Boolean": true,
	"ID":      true,
}

type ScalarData struct {
	SchemaScalarMap SchemaScalarMap
	Imports         []string
}

func buildSchemaScalarMap(schema *ast.Schema, bindings typegql.TypeMap) ScalarData {
	schemaScalarMap := make(SchemaScalarMap)
	imports := make([]string, 0)

	for _, scalar := range schema.Types {
		if scalar.Kind != ast.Scalar {
			continue
		}
		typeMapEntry, ok := bindings[scalar.Name]
		if !ok {
			typeMapEntry = typegql.AnyType()
		}

		schemaScalarMap[scalar.Name] = SchemaScalar{
			Name:        scalar.Name,
			Description: scalar.Description,
			GoType:      typeMapEntry.GoType,
			IsBuiltIn:   builtInScalars[scalar.Name],
		}

		if typeMapEntry.GoImport != "" {
			imports = append(imports, typeMapEntry.GoImport)
		}
	}
	return ScalarData{
		SchemaScalarMap: schemaScalarMap,
		Imports:         imports,
	}
}
