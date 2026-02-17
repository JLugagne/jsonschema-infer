package jsonschema

import (
	"encoding/json"
	"testing"
)

func TestWithExamples(t *testing.T) {
	t.Run("ExamplesDisabledByDefaultForNew", func(t *testing.T) {
		generator := New()
		generator.AddSample(`{"name": "John"}`)
		
		schemaJSON, _ := generator.Generate()
		var schema Schema
		json.Unmarshal([]byte(schemaJSON), &schema)
		
		if schema.Properties["name"].Example != nil {
			t.Errorf("Expected nil example, got %v", schema.Properties["name"].Example)
		}
	})

	t.Run("ExamplesExplicitlyEnabled", func(t *testing.T) {
		generator := New(WithExamples())
		generator.AddSample(`{"name": "John"}`)
		
		schemaJSON, _ := generator.Generate()
		var schema Schema
		json.Unmarshal([]byte(schemaJSON), &schema)
		
		if schema.Properties["name"].Example != "John" {
			t.Errorf("Expected example 'John', got %v", schema.Properties["name"].Example)
		}
	})
}
