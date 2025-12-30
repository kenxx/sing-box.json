// Package schema provides a custom JSON Schema generator using reflection.
package schema

// Schema represents a JSON Schema (Draft-07).
type Schema struct {
	// Meta
	Schema      string `json:"$schema,omitempty"`
	ID          string `json:"$id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`

	// References
	Ref         string             `json:"$ref,omitempty"`
	Definitions map[string]*Schema `json:"$defs,omitempty"`

	// Type
	Type interface{} `json:"type,omitempty"` // string or []string

	// Object
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Required             []string           `json:"required,omitempty"`
	AdditionalProperties interface{}        `json:"additionalProperties,omitempty"`

	// Array
	Items *Schema `json:"items,omitempty"`

	// Composition
	OneOf []*Schema `json:"oneOf,omitempty"`
	AnyOf []*Schema `json:"anyOf,omitempty"`
	AllOf []*Schema `json:"allOf,omitempty"`

	// Validation
	Const   interface{}   `json:"const,omitempty"`
	Default interface{}   `json:"default,omitempty"`
	Enum    []interface{} `json:"enum,omitempty"`

	// String
	MinLength *int   `json:"minLength,omitempty"`
	MaxLength *int   `json:"maxLength,omitempty"`
	Pattern   string `json:"pattern,omitempty"`
	Format    string `json:"format,omitempty"`

	// Numeric
	Minimum          *float64 `json:"minimum,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty"`
	MultipleOf       *float64 `json:"multipleOf,omitempty"`

	// Array
	MinItems    *int `json:"minItems,omitempty"`
	MaxItems    *int `json:"maxItems,omitempty"`
	UniqueItems bool `json:"uniqueItems,omitempty"`
}

// NewSchema creates a new empty Schema.
func NewSchema() *Schema {
	return &Schema{}
}

// WithType sets the type and returns the schema for chaining.
func (s *Schema) WithType(t string) *Schema {
	s.Type = t
	return s
}

// WithRef sets the $ref and returns the schema for chaining.
func (s *Schema) WithRef(ref string) *Schema {
	s.Ref = ref
	return s
}

// WithConst sets the const value and returns the schema for chaining.
func (s *Schema) WithConst(v interface{}) *Schema {
	s.Const = v
	return s
}

// WithDefault sets the default value and returns the schema for chaining.
func (s *Schema) WithDefault(v interface{}) *Schema {
	s.Default = v
	return s
}

// AddProperty adds a property to the schema.
func (s *Schema) AddProperty(name string, prop *Schema) {
	if s.Properties == nil {
		s.Properties = make(map[string]*Schema)
	}
	s.Properties[name] = prop
}

// AddRequired adds a required field.
func (s *Schema) AddRequired(name string) {
	for _, r := range s.Required {
		if r == name {
			return
		}
	}
	s.Required = append(s.Required, name)
}

// AddDefinition adds a definition to $defs.
func (s *Schema) AddDefinition(name string, def *Schema) {
	if s.Definitions == nil {
		s.Definitions = make(map[string]*Schema)
	}
	s.Definitions[name] = def
}

// HasDefinition checks if a definition exists.
func (s *Schema) HasDefinition(name string) bool {
	if s.Definitions == nil {
		return false
	}
	_, ok := s.Definitions[name]
	return ok
}

