package jsonschema

import "encoding/json"

// Schema represents a JSON Schema
type Schema struct {
	Schema               string             `json:"$schema,omitempty"`
	Type                 interface{}        `json:"type,omitempty"` // can be string or []string
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Format               string             `json:"format,omitempty"`
	AdditionalProperties *bool              `json:"additionalProperties,omitempty"`
}

// NewSchema creates a new Schema with default values
func NewSchema() *Schema {
	return &Schema{
		Schema: "http://json-schema.org/draft-07/schema#",
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
