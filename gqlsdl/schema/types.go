package schema

// IntrospectionResponse represents the root response from a GraphQL introspection query
type IntrospectionResponse struct {
	Data IntrospectionData `json:"data"`
}

// IntrospectionData contains the __schema field
type IntrospectionData struct {
	Schema IntrospectionSchema `json:"__schema"`
}

// IntrospectionSchema represents the GraphQL schema structure
type IntrospectionSchema struct {
	QueryType        *TypeRef    `json:"queryType"`
	MutationType     *TypeRef    `json:"mutationType"`
	SubscriptionType *TypeRef    `json:"subscriptionType"`
	Types            []FullType  `json:"types"`
	Directives       []Directive `json:"directives"`
}

// TypeRef is a reference to a type by name
type TypeRef struct {
	Name string `json:"name"`
}

// FullType represents a complete GraphQL type definition
type FullType struct {
	Kind          string       `json:"kind"`
	Name          string       `json:"name"`
	Description   *string      `json:"description"`
	Fields        []Field      `json:"fields"`
	InputFields   []InputValue `json:"inputFields"`
	Interfaces    []TypeInfo   `json:"interfaces"`
	EnumValues    []EnumValue  `json:"enumValues"`
	PossibleTypes []TypeInfo   `json:"possibleTypes"`
}

// Field represents a field in an object or interface type
type Field struct {
	Name              string       `json:"name"`
	Description       *string      `json:"description"`
	Args              []InputValue `json:"args"`
	Type              TypeInfo     `json:"type"`
	IsDeprecated      bool         `json:"isDeprecated"`
	DeprecationReason *string      `json:"deprecationReason"`
}

// InputValue represents an input field or argument
type InputValue struct {
	Name         string   `json:"name"`
	Description  *string  `json:"description"`
	Type         TypeInfo `json:"type"`
	DefaultValue *string  `json:"defaultValue"`
}

// TypeInfo represents type information with possible nesting
type TypeInfo struct {
	Kind   string    `json:"kind"`
	Name   *string   `json:"name"`
	OfType *TypeInfo `json:"ofType"`
}

// EnumValue represents a value in an enum type
type EnumValue struct {
	Name              string  `json:"name"`
	Description       *string `json:"description"`
	IsDeprecated      bool    `json:"isDeprecated"`
	DeprecationReason *string `json:"deprecationReason"`
}

// Directive represents a GraphQL directive
type Directive struct {
	Name        string       `json:"name"`
	Description *string      `json:"description"`
	Locations   []string     `json:"locations"`
	Args        []InputValue `json:"args"`
}
