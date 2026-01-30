package schema

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// ConvertToSDL converts an introspection schema to SDL format
func ConvertToSDL(schema *IntrospectionSchema) string {
	var sb strings.Builder

	// Write schema definition if custom root types exist
	writeSchemaDefinition(&sb, schema)

	// Sort types by name for consistent output
	types := make([]FullType, len(schema.Types))
	copy(types, schema.Types)
	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})

	// Write types
	for _, t := range types {
		// Skip built-in types
		if strings.HasPrefix(t.Name, "__") {
			continue
		}

		writeType(&sb, t)
	}

	// Write directives (skip built-in ones)
	for _, d := range schema.Directives {
		if isBuiltInDirective(d.Name) {
			continue
		}
		writeDirective(&sb, d)
	}

	return sb.String()
}

func writeSchemaDefinition(sb *strings.Builder, schema *IntrospectionSchema) {
	hasCustomRootTypes := false

	if schema.QueryType != nil && schema.QueryType.Name != "Query" {
		hasCustomRootTypes = true
	}
	if schema.MutationType != nil && schema.MutationType.Name != "Mutation" {
		hasCustomRootTypes = true
	}
	if schema.SubscriptionType != nil && schema.SubscriptionType.Name != "Subscription" {
		hasCustomRootTypes = true
	}

	if hasCustomRootTypes {
		sb.WriteString("schema {\n")
		if schema.QueryType != nil {
			sb.WriteString(fmt.Sprintf("  query: %s\n", schema.QueryType.Name))
		}
		if schema.MutationType != nil {
			sb.WriteString(fmt.Sprintf("  mutation: %s\n", schema.MutationType.Name))
		}
		if schema.SubscriptionType != nil {
			sb.WriteString(fmt.Sprintf("  subscription: %s\n", schema.SubscriptionType.Name))
		}
		sb.WriteString("}\n\n")
	}
}

func writeType(sb *strings.Builder, t FullType) {
	switch t.Kind {
	case "SCALAR":
		writeScalar(sb, t)
	case "OBJECT":
		writeObject(sb, t)
	case "INTERFACE":
		writeInterface(sb, t)
	case "UNION":
		writeUnion(sb, t)
	case "ENUM":
		writeEnum(sb, t)
	case "INPUT_OBJECT":
		writeInputObject(sb, t)
	}
}

func writeScalar(sb *strings.Builder, t FullType) {
	// Skip built-in scalars
	if isBuiltInScalar(t.Name) {
		return
	}

	writeDescription(sb, t.Description, "")
	sb.WriteString(fmt.Sprintf("scalar %s\n\n", t.Name))
}

func writeObject(sb *strings.Builder, t FullType) {
	writeDescription(sb, t.Description, "")
	sb.WriteString(fmt.Sprintf("type %s", t.Name))

	if len(t.Interfaces) > 0 {
		interfaces := make([]string, len(t.Interfaces))
		for i, iface := range t.Interfaces {
			if iface.Name != nil {
				interfaces[i] = *iface.Name
			}
		}
		sb.WriteString(fmt.Sprintf(" implements %s", strings.Join(interfaces, " & ")))
	}

	sb.WriteString(" {\n")
	writeFields(sb, t.Fields)
	sb.WriteString("}\n\n")
}

func writeInterface(sb *strings.Builder, t FullType) {
	writeDescription(sb, t.Description, "")
	sb.WriteString(fmt.Sprintf("interface %s", t.Name))

	if len(t.Interfaces) > 0 {
		interfaces := make([]string, len(t.Interfaces))
		for i, iface := range t.Interfaces {
			if iface.Name != nil {
				interfaces[i] = *iface.Name
			}
		}
		sb.WriteString(fmt.Sprintf(" implements %s", strings.Join(interfaces, " & ")))
	}

	sb.WriteString(" {\n")
	writeFields(sb, t.Fields)
	sb.WriteString("}\n\n")
}

func writeUnion(sb *strings.Builder, t FullType) {
	writeDescription(sb, t.Description, "")

	types := make([]string, len(t.PossibleTypes))
	for i, pt := range t.PossibleTypes {
		if pt.Name != nil {
			types[i] = *pt.Name
		}
	}

	sb.WriteString(fmt.Sprintf("union %s = %s\n\n", t.Name, strings.Join(types, " | ")))
}

func writeEnum(sb *strings.Builder, t FullType) {
	writeDescription(sb, t.Description, "")
	sb.WriteString(fmt.Sprintf("enum %s {\n", t.Name))

	for _, ev := range t.EnumValues {
		writeDescription(sb, ev.Description, "  ")
		sb.WriteString(fmt.Sprintf("  %s", ev.Name))

		if ev.IsDeprecated {
			if ev.DeprecationReason != nil && *ev.DeprecationReason != "" {
				sb.WriteString(fmt.Sprintf(" @deprecated(reason: %q)", *ev.DeprecationReason))
			} else {
				sb.WriteString(" @deprecated")
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("}\n\n")
}

func writeInputObject(sb *strings.Builder, t FullType) {
	writeDescription(sb, t.Description, "")
	sb.WriteString(fmt.Sprintf("input %s {\n", t.Name))

	for _, f := range t.InputFields {
		writeDescription(sb, f.Description, "  ")
		sb.WriteString(fmt.Sprintf("  %s: %s", f.Name, formatType(f.Type)))

		if f.DefaultValue != nil {
			sb.WriteString(fmt.Sprintf(" = %s", *f.DefaultValue))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("}\n\n")
}

func writeFields(sb *strings.Builder, fields []Field) {
	for _, f := range fields {
		writeDescription(sb, f.Description, "  ")
		sb.WriteString(fmt.Sprintf("  %s", f.Name))

		if len(f.Args) > 0 {
			writeArguments(sb, f.Args)
		}

		sb.WriteString(fmt.Sprintf(": %s", formatType(f.Type)))

		if f.IsDeprecated {
			if f.DeprecationReason != nil && *f.DeprecationReason != "" {
				sb.WriteString(fmt.Sprintf(" @deprecated(reason: %q)", *f.DeprecationReason))
			} else {
				sb.WriteString(" @deprecated")
			}
		}
		sb.WriteString("\n")
	}
}

func writeArguments(sb *strings.Builder, args []InputValue) {
	if len(args) == 0 {
		return
	}

	// Check if any argument has a description
	hasDescription := false
	for _, arg := range args {
		if arg.Description != nil && *arg.Description != "" {
			hasDescription = true
			break
		}
	}

	if hasDescription || len(args) > 2 {
		// Multi-line format
		sb.WriteString("(\n")
		for _, arg := range args {
			writeDescription(sb, arg.Description, "    ")
			sb.WriteString(fmt.Sprintf("    %s: %s", arg.Name, formatType(arg.Type)))
			if arg.DefaultValue != nil {
				sb.WriteString(fmt.Sprintf(" = %s", *arg.DefaultValue))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("  )")
	} else {
		// Single-line format
		sb.WriteString("(")
		argStrs := make([]string, len(args))
		for i, arg := range args {
			argStr := fmt.Sprintf("%s: %s", arg.Name, formatType(arg.Type))
			if arg.DefaultValue != nil {
				argStr += fmt.Sprintf(" = %s", *arg.DefaultValue)
			}
			argStrs[i] = argStr
		}
		sb.WriteString(strings.Join(argStrs, ", "))
		sb.WriteString(")")
	}
}

func writeDirective(sb *strings.Builder, d Directive) {
	writeDescription(sb, d.Description, "")
	sb.WriteString(fmt.Sprintf("directive @%s", d.Name))

	if len(d.Args) > 0 {
		writeArguments(sb, d.Args)
	}

	if len(d.Locations) > 0 {
		sb.WriteString(" on ")
		sb.WriteString(strings.Join(d.Locations, " | "))
	}

	sb.WriteString("\n\n")
}

func writeDescription(sb *strings.Builder, description *string, indent string) {
	if description == nil || *description == "" {
		return
	}

	desc := *description
	if strings.Contains(desc, "\n") {
		sb.WriteString(fmt.Sprintf("%s\"\"\"\n", indent))
		lines := strings.Split(desc, "\n")
		for _, line := range lines {
			sb.WriteString(fmt.Sprintf("%s%s\n", indent, line))
		}
		sb.WriteString(fmt.Sprintf("%s\"\"\"\n", indent))
	} else {
		sb.WriteString(fmt.Sprintf("%s\"%s\"\n", indent, escapeString(desc)))
	}
}

func formatType(t TypeInfo) string {
	switch t.Kind {
	case "NON_NULL":
		if t.OfType != nil {
			return formatType(*t.OfType) + "!"
		}
	case "LIST":
		if t.OfType != nil {
			return "[" + formatType(*t.OfType) + "]"
		}
	default:
		if t.Name != nil {
			return *t.Name
		}
	}
	return ""
}

func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

func isBuiltInScalar(name string) bool {
	builtInScalars := map[string]bool{
		"Int":     true,
		"Float":   true,
		"String":  true,
		"Boolean": true,
		"ID":      true,
	}
	return builtInScalars[name]
}

func isBuiltInDirective(name string) bool {
	builtInDirectives := map[string]bool{
		"skip":        true,
		"include":     true,
		"deprecated":  true,
		"specifiedBy": true,
	}
	return builtInDirectives[name]
}

// SaveToFile saves the SDL schema to a file
func SaveToFile(sdl string, filepath string) error {
	return os.WriteFile(filepath, []byte(sdl), 0644)
}
