package schema

import (
	"reflect"
	"slices"
	"strings"
)

// Reflector generates JSON Schema from Go types using reflection.
type Reflector struct {
	// AllowAdditionalProperties allows additional properties in objects.
	AllowAdditionalProperties bool

	// definitions stores type definitions to avoid duplication.
	definitions map[string]*Schema

	// seen tracks types being processed to avoid infinite recursion.
	seen map[reflect.Type]bool
}

// NewReflector creates a new Reflector.
func NewReflector() *Reflector {
	return &Reflector{
		definitions: make(map[string]*Schema),
		seen:        make(map[reflect.Type]bool),
	}
}

// Reflect generates a JSON Schema from a Go value.
func (r *Reflector) Reflect(v interface{}) *Schema {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := r.typeToSchema(t)

	// Move collected definitions to root schema
	if len(r.definitions) > 0 {
		schema.Definitions = r.definitions
	}

	return schema
}

// typeToSchema converts a reflect.Type to a Schema.
func (r *Reflector) typeToSchema(t reflect.Type) *Schema {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		return r.typeToSchema(t.Elem())
	}

	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}

	case reflect.Bool:
		return &Schema{Type: "boolean"}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer"}

	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}

	case reflect.Slice, reflect.Array:
		return r.sliceToSchema(t)

	case reflect.Map:
		return r.mapToSchema(t)

	case reflect.Struct:
		return r.structToSchema(t)

	case reflect.Interface:
		// interface{} / any - return empty schema (accepts anything)
		return &Schema{}

	default:
		return &Schema{}
	}
}

// sliceToSchema handles slice and array types.
func (r *Reflector) sliceToSchema(t reflect.Type) *Schema {
	itemSchema := r.typeToSchema(t.Elem())
	return &Schema{
		Type:  "array",
		Items: itemSchema,
	}
}

// mapToSchema handles map types.
func (r *Reflector) mapToSchema(t reflect.Type) *Schema {
	valueSchema := r.typeToSchema(t.Elem())
	return &Schema{
		Type:                 "object",
		AdditionalProperties: valueSchema,
	}
}

// structToSchema handles struct types.
func (r *Reflector) structToSchema(t reflect.Type) *Schema {
	// Check for recursion
	if r.seen[t] {
		// Return a reference instead
		defName := r.getDefinitionName(t)
		return &Schema{Ref: "#/$defs/" + defName}
	}
	r.seen[t] = true
	defer func() { r.seen[t] = false }()

	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
	}

	if r.AllowAdditionalProperties {
		schema.AdditionalProperties = true
	}

	// First pass: handle embedded structs (so their fields come first)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Handle embedded structs
		if field.Anonymous {
			embeddedType := field.Type
			if embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}
			if embeddedType.Kind() == reflect.Struct {
				embeddedSchema := r.structToSchema(embeddedType)
				// Merge embedded properties
				for name, prop := range embeddedSchema.Properties {
					if _, exists := schema.Properties[name]; !exists {
						schema.Properties[name] = prop
					}
				}
				// Merge required
				for _, req := range embeddedSchema.Required {
					schema.AddRequired(req)
				}
			}
		}
	}

	// Second pass: handle regular fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Skip embedded structs (already handled)
		if field.Anonymous {
			continue
		}

		// Parse json tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			// Skip fields with json:"-"
			continue
		}

		name, opts := parseJSONTag(jsonTag)
		if name == "" {
			name = field.Name
		}

		// Generate schema for field
		fieldSchema := r.typeToSchema(field.Type)

		// Check for omitempty
		if !containsOption(opts, "omitempty") {
			schema.AddRequired(name)
		}

		schema.Properties[name] = fieldSchema
	}

	return schema
}

// getDefinitionName returns a normalized name for a type definition.
func (r *Reflector) getDefinitionName(t reflect.Type) string {
	name := t.Name()
	if name == "" {
		name = t.String()
	}
	// Normalize: remove special characters
	name = strings.ReplaceAll(name, "[", "")
	name = strings.ReplaceAll(name, "]", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, ".", "")
	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, "*", "")
	return name
}

// ReflectType generates schema for a reflect.Type directly.
func (r *Reflector) ReflectType(t reflect.Type) *Schema {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return r.typeToSchema(t)
}

// AddDefinition adds a named definition.
func (r *Reflector) AddDefinition(name string, schema *Schema) {
	r.definitions[name] = schema
}

// HasDefinition checks if a definition exists.
func (r *Reflector) HasDefinition(name string) bool {
	_, ok := r.definitions[name]
	return ok
}

// GetDefinitions returns all definitions.
func (r *Reflector) GetDefinitions() map[string]*Schema {
	return r.definitions
}

// parseJSONTag parses a json struct tag.
func parseJSONTag(tag string) (name string, opts []string) {
	if tag == "" {
		return "", nil
	}
	parts := strings.Split(tag, ",")
	name = parts[0]
	if len(parts) > 1 {
		opts = parts[1:]
	}
	return
}

// containsOption checks if an option is in the list.
func containsOption(opts []string, option string) bool {
	return slices.Contains(opts, option)
}
