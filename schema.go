package jsonschema

import "encoding/json"

// Schema represents a JSON Schema
type Schema struct {
	Schema               string             `json:"$schema,omitempty"`
	Type                 any                `json:"type,omitempty"` // can be string or []string
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Format               string             `json:"format,omitempty"`
	Example              any                `json:"example,omitempty"`
	AdditionalProperties *bool              `json:"additionalProperties,omitempty"`
}

// NewSchema creates a new Schema with default Draft 07 values.
func NewSchema() *Schema {
	return NewSchemaWithVersion(Draft07)
}

// NewSchemaWithVersion creates a new Schema with the specified JSON Schema version.
// This is useful when you need an empty schema structure with a specific version set.
//
// For schema inference from samples, use Generator instead:
//
//	gen := jsonschema.New(jsonschema.WithSchemaVersion(Draft06))
//	gen.AddSample(`{"your": "data"}`)
//	schemaJSON, _ := gen.Generate()
func NewSchemaWithVersion(version SchemaVersion) *Schema {
	return &Schema{
		Schema: string(version),
	}
}

// MarshalJSON customizes JSON marshaling for Schema
func (s *Schema) MarshalJSON() ([]byte, error) {
	type Alias Schema
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	})
}
