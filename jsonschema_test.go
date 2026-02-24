package jsonschema

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestBasicTypeInference(t *testing.T) {
	generator := New()

	json1 := `{"name": "John", "age": 30, "active": true}`
	json2 := `{"name": "Jane", "age": 25, "active": false}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Check that all fields are required
	if len(schema.Required) != 3 {
		t.Errorf("Expected 3 required fields, got %d", len(schema.Required))
	}

	// Check field types
	if schema.Properties["name"].Type != "string" {
		t.Errorf("Expected name to be string, got %v", schema.Properties["name"].Type)
	}

	if schema.Properties["age"].Type != "integer" {
		t.Errorf("Expected age to be integer, got %v", schema.Properties["age"].Type)
	}

	if schema.Properties["active"].Type != "boolean" {
		t.Errorf("Expected active to be boolean, got %v", schema.Properties["active"].Type)
	}
}

func TestOptionalFields(t *testing.T) {
	generator := New()

	json1 := `{"name": "John", "age": 30}`
	json2 := `{"name": "Jane"}`
	json3 := `{"name": "Bob", "age": 40}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json3)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Only "name" should be required
	if len(schema.Required) != 1 || schema.Required[0] != "name" {
		t.Errorf("Expected only 'name' to be required, got %v", schema.Required)
	}

	// Age should still be defined but not required
	if schema.Properties["age"] == nil {
		t.Error("Expected 'age' property to be defined")
	}
}

func TestArrayHandling(t *testing.T) {
	generator := New()

	json1 := `{"tags": ["go", "json", "schema"]}`
	json2 := `{"tags": ["test", "unit"]}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if schema.Properties["tags"].Type != "array" {
		t.Errorf("Expected tags to be array, got %v", schema.Properties["tags"].Type)
	}

	if schema.Properties["tags"].Items == nil {
		t.Error("Expected array items to be defined")
	}

	if schema.Properties["tags"].Items.Type != "string" {
		t.Errorf("Expected array items to be string, got %v", schema.Properties["tags"].Items.Type)
	}
}

func TestDateTimeDetection(t *testing.T) {
	generator := New()

	json1 := `{"created_at": "2023-01-15T10:30:00Z"}`
	json2 := `{"created_at": "2023-02-20T14:45:00Z"}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if schema.Properties["created_at"].Type != "string" {
		t.Errorf("Expected created_at to be string, got %v", schema.Properties["created_at"].Type)
	}

	if schema.Properties["created_at"].Format != "date-time" {
		t.Errorf("Expected created_at format to be date-time, got %v", schema.Properties["created_at"].Format)
	}
}

func TestEmailDetection(t *testing.T) {
	generator := New()

	json1 := `{"email": "user@example.com"}`
	json2 := `{"email": "admin@test.org"}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if schema.Properties["email"].Type != "string" {
		t.Errorf("Expected email to be string, got %v", schema.Properties["email"].Type)
	}

	if schema.Properties["email"].Format != "email" {
		t.Errorf("Expected email format to be email, got %v", schema.Properties["email"].Format)
	}
}

func TestUUIDDetection(t *testing.T) {
	generator := New()

	json1 := `{"id": "550e8400-e29b-41d4-a716-446655440000"}`
	json2 := `{"id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8"}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if schema.Properties["id"].Type != "string" {
		t.Errorf("Expected id to be string, got %v", schema.Properties["id"].Type)
	}

	if schema.Properties["id"].Format != "uuid" {
		t.Errorf("Expected id format to be uuid, got %v", schema.Properties["id"].Format)
	}
}

func TestIPv4Detection(t *testing.T) {
	generator := New()

	json1 := `{"ip": "192.168.1.1"}`
	json2 := `{"ip": "10.0.0.1"}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if schema.Properties["ip"].Type != "string" {
		t.Errorf("Expected ip to be string, got %v", schema.Properties["ip"].Type)
	}

	if schema.Properties["ip"].Format != "ipv4" {
		t.Errorf("Expected ip format to be ipv4, got %v", schema.Properties["ip"].Format)
	}
}

func TestIPv6Detection(t *testing.T) {
	generator := New()

	json1 := `{"ip": "2001:0db8:85a3:0000:0000:8a2e:0370:7334"}`
	json2 := `{"ip": "fe80::1"}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if schema.Properties["ip"].Type != "string" {
		t.Errorf("Expected ip to be string, got %v", schema.Properties["ip"].Type)
	}

	if schema.Properties["ip"].Format != "ipv6" {
		t.Errorf("Expected ip format to be ipv6, got %v", schema.Properties["ip"].Format)
	}
}

func TestURLDetection(t *testing.T) {
	generator := New()

	json1 := `{"website": "https://example.com"}`
	json2 := `{"website": "http://test.org/path"}`
	json3 := `{"website": "ftp://files.example.com/data"}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json3)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if schema.Properties["website"].Type != "string" {
		t.Errorf("Expected website to be string, got %v", schema.Properties["website"].Type)
	}

	if schema.Properties["website"].Format != "uri" {
		t.Errorf("Expected website format to be uri, got %v", schema.Properties["website"].Format)
	}
}

func TestPredefinedTypes(t *testing.T) {
	generator := New(
		WithPredefined("created_at", DateTime),
		WithPredefined("updated_at", DateTime),
	)

	json1 := `{"created_at": "2023-01-15T10:30:00Z", "updated_at": "2023-01-15T10:30:00Z"}`
	json2 := `{"created_at": "2023-02-20T14:45:00Z", "updated_at": "2023-02-20T14:45:00Z"}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if schema.Properties["created_at"].Format != "date-time" {
		t.Errorf("Expected created_at format to be date-time, got %v", schema.Properties["created_at"].Format)
	}

	if schema.Properties["updated_at"].Format != "date-time" {
		t.Errorf("Expected updated_at format to be date-time, got %v", schema.Properties["updated_at"].Format)
	}
}

func TestNestedObjects(t *testing.T) {
	generator := New()

	json1 := `{"user": {"name": "John", "email": "john@example.com"}}`
	json2 := `{"user": {"name": "Jane", "email": "jane@example.com"}}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if schema.Properties["user"].Type != "object" {
		t.Errorf("Expected user to be object, got %v", schema.Properties["user"].Type)
	}

	if schema.Properties["user"].Properties == nil {
		t.Fatal("Expected user to have properties")
	}

	if schema.Properties["user"].Properties["name"].Type != "string" {
		t.Errorf("Expected user.name to be string, got %v", schema.Properties["user"].Properties["name"].Type)
	}

	if schema.Properties["user"].Properties["email"].Type != "string" {
		t.Errorf("Expected user.email to be string, got %v", schema.Properties["user"].Properties["email"].Type)
	}
}

func TestNumberVsInteger(t *testing.T) {
	generator := New()

	json1 := `{"count": 10, "price": 19.99}`
	json2 := `{"count": 20, "price": 29.99}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if schema.Properties["count"].Type != "integer" {
		t.Errorf("Expected count to be integer, got %v", schema.Properties["count"].Type)
	}

	if schema.Properties["price"].Type != "number" {
		t.Errorf("Expected price to be number, got %v", schema.Properties["price"].Type)
	}
}

func TestEmptySamples(t *testing.T) {
	generator := New()

	_, err := generator.Generate()
	if err == nil {
		t.Error("Expected error when generating schema with no samples")
	}
}

func TestInvalidJSON(t *testing.T) {
	generator := New()

	err := generator.AddSample(`{invalid json}`)
	if err == nil {
		t.Error("Expected error when adding invalid JSON")
	}
}

func TestArrayOfObjects(t *testing.T) {
	generator := New()

	json1 := `{"users": [{"id": 1, "name": "John"}, {"id": 2, "name": "Jane"}]}`
	json2 := `{"users": [{"id": 3, "name": "Bob"}]}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Check that users is an array
	if schema.Properties["users"].Type != "array" {
		t.Errorf("Expected users to be array, got %v", schema.Properties["users"].Type)
	}

	// Check that array items are objects
	if schema.Properties["users"].Items == nil {
		t.Fatal("Expected users to have items defined")
	}

	if schema.Properties["users"].Items.Type != "object" {
		t.Errorf("Expected users items to be object, got %v", schema.Properties["users"].Items.Type)
	}

	// Check that object properties are defined
	if schema.Properties["users"].Items.Properties == nil {
		t.Fatal("Expected users items to have properties")
	}

	if schema.Properties["users"].Items.Properties["id"].Type != "integer" {
		t.Errorf("Expected id to be integer, got %v", schema.Properties["users"].Items.Properties["id"].Type)
	}

	if schema.Properties["users"].Items.Properties["name"].Type != "string" {
		t.Errorf("Expected name to be string, got %v", schema.Properties["users"].Items.Properties["name"].Type)
	}

	// Check that all properties are required (they appear in all array items)
	if len(schema.Properties["users"].Items.Required) != 2 {
		t.Errorf("Expected 2 required fields in array items, got %d", len(schema.Properties["users"].Items.Required))
	}
}

func TestArrayOfObjectsWithOptionalFields(t *testing.T) {
	generator := New()

	json1 := `{"products": [{"id": 1, "name": "Product A", "price": 10.5}, {"id": 2, "name": "Product B"}]}`
	json2 := `{"products": [{"id": 3, "name": "Product C", "price": 20.0}]}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Check array items structure
	items := schema.Properties["products"].Items
	if items == nil {
		t.Fatal("Expected products to have items")
	}

	// Check that id and name are required (in all items)
	requiredCount := len(items.Required)
	if requiredCount != 2 {
		t.Errorf("Expected 2 required fields (id, name), got %d: %v", requiredCount, items.Required)
	}

	// Check that price exists but is optional (not in all items)
	if items.Properties["price"] == nil {
		t.Error("Expected price property to be defined")
	}

	// Verify price is not in required list
	priceRequired := false
	for _, req := range items.Required {
		if req == "price" {
			priceRequired = true
			break
		}
	}
	if priceRequired {
		t.Error("Expected price to be optional, but it was required")
	}
}

func TestNestedArraysOfObjects(t *testing.T) {
	generator := New()

	json1 := `{"company": {"name": "Acme", "employees": [{"name": "Alice", "role": "Engineer"}]}}`
	json2 := `{"company": {"name": "TechCo", "employees": [{"name": "Bob", "role": "Designer"}, {"name": "Charlie", "role": "Manager"}]}}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Navigate to nested structure
	company := schema.Properties["company"]
	if company.Type != "object" {
		t.Errorf("Expected company to be object, got %v", company.Type)
	}

	employees := company.Properties["employees"]
	if employees.Type != "array" {
		t.Errorf("Expected employees to be array, got %v", employees.Type)
	}

	// Check array items
	if employees.Items == nil {
		t.Fatal("Expected employees to have items")
	}

	if employees.Items.Type != "object" {
		t.Errorf("Expected employees items to be object, got %v", employees.Items.Type)
	}

	// Check employee properties
	if employees.Items.Properties["name"].Type != "string" {
		t.Errorf("Expected employee name to be string, got %v", employees.Items.Properties["name"].Type)
	}

	if employees.Items.Properties["role"].Type != "string" {
		t.Errorf("Expected employee role to be string, got %v", employees.Items.Properties["role"].Type)
	}
}

func TestLoadSchema(t *testing.T) {
	// First, generate a schema
	generator1 := New()
	err := generator1.AddSample(`{"name": "John", "age": 30}`)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator1.AddSample(`{"name": "Jane", "age": 25}`)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator1.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// Now load it into a new generator
	generator2 := New()
	err = generator2.Load(schemaJSON)
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	// Generate schema from loaded generator
	loadedSchemaJSON, err := generator2.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema from loaded generator: %v", err)
	}

	// Parse both schemas
	var originalSchema, loadedSchema Schema
	err = json.Unmarshal([]byte(schemaJSON), &originalSchema)
	if err != nil {
		t.Fatalf("Failed to unmarshal original schema: %v", err)
	}
	err = json.Unmarshal([]byte(loadedSchemaJSON), &loadedSchema)
	if err != nil {
		t.Fatalf("Failed to unmarshal loaded schema: %v", err)
	}

	// Verify they have the same structure
	if loadedSchema.Type != originalSchema.Type {
		t.Errorf("Expected type %v, got %v", originalSchema.Type, loadedSchema.Type)
	}

	if len(loadedSchema.Properties) != len(originalSchema.Properties) {
		t.Errorf("Expected %d properties, got %d", len(originalSchema.Properties), len(loadedSchema.Properties))
	}

	// Check specific properties
	if loadedSchema.Properties["name"].Type != "string" {
		t.Errorf("Expected name to be string, got %v", loadedSchema.Properties["name"].Type)
	}

	if loadedSchema.Properties["age"].Type != "integer" {
		t.Errorf("Expected age to be integer, got %v", loadedSchema.Properties["age"].Type)
	}
}

func TestLoadSchemaAndAddSamples(t *testing.T) {
	// Create initial schema with 2 fields
	generator1 := New()
	err := generator1.AddSample(`{"name": "John", "age": 30}`)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	schemaJSON, err := generator1.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// Load schema and add new sample with additional field
	generator2 := New()
	err = generator2.Load(schemaJSON)
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	// Add a new sample with an additional field
	err = generator2.AddSample(`{"name": "Bob", "age": 40, "email": "bob@example.com"}`)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	newSchemaJSON, err := generator2.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var newSchema Schema
	err = json.Unmarshal([]byte(newSchemaJSON), &newSchema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Should now have 3 properties
	if len(newSchema.Properties) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(newSchema.Properties))
	}

	// Email should exist but not be required (only in 1 out of 2 samples)
	if newSchema.Properties["email"] == nil {
		t.Error("Expected email property to exist")
	}

	emailRequired := false
	for _, req := range newSchema.Required {
		if req == "email" {
			emailRequired = true
			break
		}
	}
	if emailRequired {
		t.Error("Expected email to be optional")
	}
}

func TestConcurrentAddSample(t *testing.T) {
	generator := New()

	samples := []string{
		`{"name": "John", "age": 30, "active": true}`,
		`{"name": "Jane", "age": 25, "active": false}`,
		`{"name": "Bob", "age": 35, "active": true}`,
		`{"name": "Alice", "age": 28, "active": false}`,
		`{"name": "Charlie", "age": 32, "active": true}`,
		`{"name": "Diana", "age": 27, "active": false}`,
		`{"name": "Eve", "age": 29, "active": true}`,
		`{"name": "Frank", "age": 31, "active": false}`,
	}

	// Add samples concurrently from multiple goroutines
	done := make(chan bool)
	for _, sample := range samples {
		go func(s string) {
			err := generator.AddSample(s)
			if err != nil {
				t.Errorf("Failed to add sample: %v", err)
			}
			done <- true
		}(sample)
	}

	// Wait for all goroutines to complete
	for range samples {
		<-done
	}

	// Generate schema
	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Verify all fields are required (appeared in all 8 samples)
	if len(schema.Required) != 3 {
		t.Errorf("Expected 3 required fields, got %d", len(schema.Required))
	}

	// Verify field types
	if schema.Properties["name"].Type != "string" {
		t.Errorf("Expected name to be string, got %v", schema.Properties["name"].Type)
	}

	if schema.Properties["age"].Type != "integer" {
		t.Errorf("Expected age to be integer, got %v", schema.Properties["age"].Type)
	}

	if schema.Properties["active"].Type != "boolean" {
		t.Errorf("Expected active to be boolean, got %v", schema.Properties["active"].Type)
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	generator := New()

	// Initial samples
	err := generator.AddSample(`{"name": "John", "age": 30}`)
	if err != nil {
		t.Fatalf("Failed to add initial sample: %v", err)
	}

	// Concurrently add samples and generate schemas
	done := make(chan bool)

	// Writer goroutines
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				sample := `{"name": "User", "age": 25}`
				_ = generator.AddSample(sample)
			}
			done <- true
		}(i)
	}

	// Reader goroutines
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				_, _ = generator.Generate()
				_ = generator.GetCurrentSchema()
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Final verification
	schema := generator.GetCurrentSchema()
	if schema == nil {
		t.Fatal("Expected schema to be generated")
	}

	if len(schema.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(schema.Properties))
	}
}

func TestCustomFormatDetection(t *testing.T) {
	// Define a custom format detector for hex colors
	isHexColor := func(s string) bool {
		if len(s) != 7 {
			return false
		}
		if s[0] != '#' {
			return false
		}
		for i := 1; i < 7; i++ {
			c := s[i]
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
		return true
	}

	generator := New(WithCustomFormat("hex-color", isHexColor))

	json1 := `{"color": "#FF5733"}`
	json2 := `{"color": "#00FF00"}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if schema.Properties["color"].Type != "string" {
		t.Errorf("Expected color to be string, got %v", schema.Properties["color"].Type)
	}

	if schema.Properties["color"].Format != "hex-color" {
		t.Errorf("Expected color format to be hex-color, got %v", schema.Properties["color"].Format)
	}
}

func TestArrayAsRoot(t *testing.T) {
	generator := New()

	json1 := `[{"id": 1, "name": "John"}, {"id": 2, "name": "Jane"}]`
	json2 := `[{"id": 3, "name": "Bob"}]`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if schema.Type != "array" {
		t.Errorf("Expected root type to be array, got %v", schema.Type)
	}

	if schema.Items == nil {
		t.Fatal("Expected items schema to be defined")
	}

	if schema.Items.Type != "object" {
		t.Errorf("Expected items type to be object, got %v", schema.Items.Type)
	}

	if len(schema.Items.Properties) != 2 {
		t.Errorf("Expected 2 properties in items, got %d", len(schema.Items.Properties))
	}

	if schema.Items.Properties["id"].Type != "integer" {
		t.Errorf("Expected id to be integer, got %v", schema.Items.Properties["id"].Type)
	}

	if schema.Items.Properties["name"].Type != "string" {
		t.Errorf("Expected name to be string, got %v", schema.Items.Properties["name"].Type)
	}

	if len(schema.Items.Required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(schema.Items.Required))
	}
}

func TestPrimitiveAsRoot(t *testing.T) {
	generator := New()

	json1 := `"hello"`
	json2 := `"world"`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if schema.Type != "string" {
		t.Errorf("Expected root type to be string, got %v", schema.Type)
	}
}

func TestMultipleCustomFormats(t *testing.T) {
	// Define custom format detectors
	isHexColor := func(s string) bool {
		if len(s) != 7 || s[0] != '#' {
			return false
		}
		for i := 1; i < 7; i++ {
			c := s[i]
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
		return true
	}

	isPhoneNumber := func(s string) bool {
		if len(s) < 10 {
			return false
		}
		if s[0] != '+' {
			return false
		}
		for i := 1; i < len(s); i++ {
			if s[i] < '0' || s[i] > '9' {
				return false
			}
		}
		return true
	}

	generator := New(
		WithCustomFormat("hex-color", isHexColor),
		WithCustomFormat("phone", isPhoneNumber),
	)

	json1 := `{"color": "#FF5733", "phone": "+1234567890"}`
	json2 := `{"color": "#00FF00", "phone": "+9876543210"}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if schema.Properties["color"].Format != "hex-color" {
		t.Errorf("Expected color format to be hex-color, got %v", schema.Properties["color"].Format)
	}

	if schema.Properties["phone"].Format != "phone" {
		t.Errorf("Expected phone format to be phone, got %v", schema.Properties["phone"].Format)
	}
}

func TestWithoutBuiltInFormats(t *testing.T) {
	// Test that built-in formats can be disabled
	generator := New(WithoutBuiltInFormats())

	json1 := `{"time": "2023-01-15T10:30:00Z", "email": "user@example.com"}`
	json2 := `{"time": "2023-02-20T14:45:00Z", "email": "admin@test.org"}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Both should be plain strings without format
	if schema.Properties["time"].Format != "" {
		t.Errorf("Expected time format to be empty, got %v", schema.Properties["time"].Format)
	}

	if schema.Properties["email"].Format != "" {
		t.Errorf("Expected email format to be empty, got %v", schema.Properties["email"].Format)
	}
}

func TestCustomFormatOverride(t *testing.T) {
	// Test that you can define your own formats when built-ins are disabled
	myDateDetector := func(s string) bool {
		return len(s) == 10 && s[4] == '-' && s[7] == '-'
	}

	generator := New(
		WithoutBuiltInFormats(),
		WithCustomFormat("my-date", myDateDetector),
	)

	json1 := `{"date": "2023-01-15"}`
	json2 := `{"date": "2024-12-31"}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	if schema.Properties["date"].Format != "my-date" {
		t.Errorf("Expected date format to be my-date, got %v", schema.Properties["date"].Format)
	}
}

func TestExamples(t *testing.T) {
	generator := New(WithExamples())

	json1 := `{"name": "John", "age": 30, "active": true, "tags": ["go", "json"]}`
	json2 := `{"name": "Jane", "age": 25, "active": false, "tags": ["test"]}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Check that examples are captured from the first sample
	if schema.Properties["name"].Example != "John" {
		t.Errorf("Expected name example to be 'John', got %v", schema.Properties["name"].Example)
	}

	if schema.Properties["age"].Example != float64(30) {
		t.Errorf("Expected age example to be 30, got %v", schema.Properties["age"].Example)
	}

	if schema.Properties["active"].Example != true {
		t.Errorf("Expected active example to be true, got %v", schema.Properties["active"].Example)
	}

	// Check array example
	if schema.Properties["tags"].Example == nil {
		t.Error("Expected tags to have an example")
	} else {
		tagsExample, ok := schema.Properties["tags"].Example.([]any)
		if !ok {
			t.Errorf("Expected tags example to be array, got %T", schema.Properties["tags"].Example)
		} else if len(tagsExample) != 2 || tagsExample[0] != "go" || tagsExample[1] != "json" {
			t.Errorf("Expected tags example to be ['go', 'json'], got %v", tagsExample)
		}
	}
}

func TestExamplesWithNestedObjects(t *testing.T) {
	generator := New(WithExamples())

	json1 := `{"user": {"name": "John", "email": "john@example.com"}, "count": 5}`
	json2 := `{"user": {"name": "Jane", "email": "jane@example.com"}, "count": 10}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Check nested object example
	if schema.Properties["user"].Example == nil {
		t.Fatal("Expected user to have an example")
	}

	userExample, ok := schema.Properties["user"].Example.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected user example to be object, got %T", schema.Properties["user"].Example)
	}

	if userExample["name"] != "John" {
		t.Errorf("Expected user.name example to be 'John', got %v", userExample["name"])
	}

	if userExample["email"] != "john@example.com" {
		t.Errorf("Expected user.email example to be 'john@example.com', got %v", userExample["email"])
	}

	// Check that nested properties also have examples
	if schema.Properties["user"].Properties["name"].Example != "John" {
		t.Errorf("Expected nested name example to be 'John', got %v", schema.Properties["user"].Properties["name"].Example)
	}

	if schema.Properties["count"].Example != float64(5) {
		t.Errorf("Expected count example to be 5, got %v", schema.Properties["count"].Example)
	}
}

func TestSchemaVersionDefault(t *testing.T) {
	// Test that default schema version is Draft 07
	generator := New()

	json1 := `{"name": "John", "age": 30}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	expectedVersion := "http://json-schema.org/draft-07/schema#"
	if schema.Schema != expectedVersion {
		t.Errorf("Expected schema version to be %s, got %s", expectedVersion, schema.Schema)
	}
}

func TestSchemaVersionDraft07Explicit(t *testing.T) {
	// Test explicit Draft 07 configuration
	generator := New(WithSchemaVersion(Draft07))

	json1 := `{"name": "John", "age": 30}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	expectedVersion := "http://json-schema.org/draft-07/schema#"
	if schema.Schema != expectedVersion {
		t.Errorf("Expected schema version to be %s, got %s", expectedVersion, schema.Schema)
	}
}

func TestSchemaVersionDraft06(t *testing.T) {
	// Test Draft 06 configuration
	generator := New(WithSchemaVersion(Draft06))

	json1 := `{"name": "John", "age": 30}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	expectedVersion := "http://json-schema.org/draft-06/schema#"
	if schema.Schema != expectedVersion {
		t.Errorf("Expected schema version to be %s, got %s", expectedVersion, schema.Schema)
	}
}

func TestSchemaVersionWithComplexTypes(t *testing.T) {
	// Test that Draft 06 works correctly with complex types
	generator := New(WithSchemaVersion(Draft06))

	json1 := `{
		"user": {
			"name": "John",
			"email": "john@example.com",
			"created": "2023-01-15T10:30:00Z"
		},
		"tags": ["golang", "testing"],
		"count": 42
	}`

	json2 := `{
		"user": {
			"name": "Jane",
			"email": "jane@example.com",
			"created": "2023-02-20T14:45:00Z"
		},
		"tags": ["python", "devops", "docker"],
		"count": 17
	}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample 1: %v", err)
	}

	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample 2: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Verify schema version
	expectedVersion := "http://json-schema.org/draft-06/schema#"
	if schema.Schema != expectedVersion {
		t.Errorf("Expected schema version to be %s, got %s", expectedVersion, schema.Schema)
	}

	// Verify structure is correct
	if schema.Type != "object" {
		t.Errorf("Expected root type to be object, got %v", schema.Type)
	}

	// Verify user object
	if schema.Properties["user"].Type != "object" {
		t.Errorf("Expected user to be object, got %v", schema.Properties["user"].Type)
	}

	if schema.Properties["user"].Properties["email"].Format != "email" {
		t.Errorf("Expected email format, got %v", schema.Properties["user"].Properties["email"].Format)
	}

	if schema.Properties["user"].Properties["created"].Format != "date-time" {
		t.Errorf("Expected date-time format, got %v", schema.Properties["user"].Properties["created"].Format)
	}

	// Verify array
	if schema.Properties["tags"].Type != "array" {
		t.Errorf("Expected tags to be array, got %v", schema.Properties["tags"].Type)
	}

	if schema.Properties["tags"].Items.Type != "string" {
		t.Errorf("Expected tags items to be string, got %v", schema.Properties["tags"].Items.Type)
	}

	// Verify integer
	if schema.Properties["count"].Type != "integer" {
		t.Errorf("Expected count to be integer, got %v", schema.Properties["count"].Type)
	}
}

func TestSchemaVersionGetCurrentSchema(t *testing.T) {
	// Test that GetCurrentSchema returns correct version
	generator := New(WithSchemaVersion(Draft06))

	json1 := `{"name": "John"}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schema := generator.GetCurrentSchema()
	if schema == nil {
		t.Fatal("Expected non-nil schema from GetCurrentSchema")
	}

	expectedVersion := "http://json-schema.org/draft-06/schema#"
	if schema.Schema != expectedVersion {
		t.Errorf("Expected schema version to be %s, got %s", expectedVersion, schema.Schema)
	}
}

func TestSchemaVersionWithFormats(t *testing.T) {
	// Test that formats work correctly with Draft 06
	generator := New(WithSchemaVersion(Draft06))

	json1 := `{
		"email": "user@example.com",
		"uuid": "550e8400-e29b-41d4-a716-446655440000",
		"ipv4": "192.168.1.1",
		"ipv6": "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		"url": "https://example.com/path"
	}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Verify schema version
	expectedVersion := "http://json-schema.org/draft-06/schema#"
	if schema.Schema != expectedVersion {
		t.Errorf("Expected schema version to be %s, got %s", expectedVersion, schema.Schema)
	}

	// Verify all formats are detected correctly
	if schema.Properties["email"].Format != "email" {
		t.Errorf("Expected email format, got %v", schema.Properties["email"].Format)
	}

	if schema.Properties["uuid"].Format != "uuid" {
		t.Errorf("Expected uuid format, got %v", schema.Properties["uuid"].Format)
	}

	if schema.Properties["ipv4"].Format != "ipv4" {
		t.Errorf("Expected ipv4 format, got %v", schema.Properties["ipv4"].Format)
	}

	if schema.Properties["ipv6"].Format != "ipv6" {
		t.Errorf("Expected ipv6 format, got %v", schema.Properties["ipv6"].Format)
	}

	if schema.Properties["url"].Format != "uri" {
		t.Errorf("Expected uri format, got %v", schema.Properties["url"].Format)
	}
}

func TestSchemaVersionWithCustomFormats(t *testing.T) {
	// Test that custom formats work with Draft 06
	isHexColor := func(s string) bool {
		if len(s) != 7 || s[0] != '#' {
			return false
		}
		for i := 1; i < 7; i++ {
			c := s[i]
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
		return true
	}

	generator := New(
		WithSchemaVersion(Draft06),
		WithCustomFormat("hex-color", isHexColor),
	)

	json1 := `{"color": "#FF5733"}`

	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Verify schema version
	expectedVersion := "http://json-schema.org/draft-06/schema#"
	if schema.Schema != expectedVersion {
		t.Errorf("Expected schema version to be %s, got %s", expectedVersion, schema.Schema)
	}

	// Verify custom format
	if schema.Properties["color"].Format != "hex-color" {
		t.Errorf("Expected hex-color format, got %v", schema.Properties["color"].Format)
	}
}

func TestSchemaVersionIncremental(t *testing.T) {
	// Test that schema version is preserved across incremental updates
	generator := New(WithSchemaVersion(Draft06))

	json1 := `{"name": "John"}`
	json2 := `{"name": "Jane", "age": 25}`
	json3 := `{"name": "Bob", "age": 30, "email": "bob@example.com"}`

	// Add first sample
	err := generator.AddSample(json1)
	if err != nil {
		t.Fatalf("Failed to add sample 1: %v", err)
	}

	schema1 := generator.GetCurrentSchema()
	if schema1.Schema != string(Draft06) {
		t.Errorf("After sample 1: expected Draft06, got %s", schema1.Schema)
	}

	// Add second sample
	err = generator.AddSample(json2)
	if err != nil {
		t.Fatalf("Failed to add sample 2: %v", err)
	}

	schema2 := generator.GetCurrentSchema()
	if schema2.Schema != string(Draft06) {
		t.Errorf("After sample 2: expected Draft06, got %s", schema2.Schema)
	}

	// Add third sample
	err = generator.AddSample(json3)
	if err != nil {
		t.Fatalf("Failed to add sample 3: %v", err)
	}

	schema3 := generator.GetCurrentSchema()
	if schema3.Schema != string(Draft06) {
		t.Errorf("After sample 3: expected Draft06, got %s", schema3.Schema)
	}

	// Final generate
	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate final schema: %v", err)
	}

	var finalSchema Schema
	err = json.Unmarshal([]byte(schemaJSON), &finalSchema)
	if err != nil {
		t.Fatalf("Failed to unmarshal final schema: %v", err)
	}

	expectedVersion := "http://json-schema.org/draft-06/schema#"
	if finalSchema.Schema != expectedVersion {
		t.Errorf("Final schema: expected version %s, got %s", expectedVersion, finalSchema.Schema)
	}
}

func TestNewSchemaWithVersionDraft06(t *testing.T) {
	// Test creating an empty schema with Draft06
	schema := NewSchemaWithVersion(Draft06)

	if schema == nil {
		t.Fatal("NewSchemaWithVersion returned nil")
	}

	expectedVersion := "http://json-schema.org/draft-06/schema#"
	if schema.Schema != expectedVersion {
		t.Errorf("Expected schema version %s, got %s", expectedVersion, schema.Schema)
	}

	// Verify other fields are empty/nil
	if schema.Type != nil {
		t.Errorf("Expected Type to be nil, got %v", schema.Type)
	}
	if schema.Properties != nil {
		t.Errorf("Expected Properties to be nil, got %v", schema.Properties)
	}
	if schema.Items != nil {
		t.Errorf("Expected Items to be nil, got %v", schema.Items)
	}
	if schema.Required != nil {
		t.Errorf("Expected Required to be nil, got %v", schema.Required)
	}
	if schema.Format != "" {
		t.Errorf("Expected Format to be empty, got %s", schema.Format)
	}
}

func TestNewSchemaWithVersionDraft07(t *testing.T) {
	// Test creating an empty schema with Draft07
	schema := NewSchemaWithVersion(Draft07)

	if schema == nil {
		t.Fatal("NewSchemaWithVersion returned nil")
	}

	expectedVersion := "http://json-schema.org/draft-07/schema#"
	if schema.Schema != expectedVersion {
		t.Errorf("Expected schema version %s, got %s", expectedVersion, schema.Schema)
	}

	// Verify other fields are empty/nil
	if schema.Type != nil {
		t.Errorf("Expected Type to be nil, got %v", schema.Type)
	}
	if schema.Properties != nil {
		t.Errorf("Expected Properties to be nil, got %v", schema.Properties)
	}
}

func TestNewSchemaBackwardCompatibility(t *testing.T) {
	// Test that deprecated NewSchema() still works and creates Draft07
	schema := NewSchema()

	if schema == nil {
		t.Fatal("NewSchema returned nil")
	}

	expectedVersion := "http://json-schema.org/draft-07/schema#"
	if schema.Schema != expectedVersion {
		t.Errorf("Expected schema version %s (Draft07 default), got %s", expectedVersion, schema.Schema)
	}
}

func TestNewSchemaWithVersionMarshaling(t *testing.T) {
	// Test that schemas created with NewSchemaWithVersion marshal correctly
	schema := NewSchemaWithVersion(Draft06)

	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}

	var unmarshaled Schema
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	expectedVersion := "http://json-schema.org/draft-06/schema#"
	if unmarshaled.Schema != expectedVersion {
		t.Errorf("Expected schema version %s after marshal/unmarshal, got %s", expectedVersion, unmarshaled.Schema)
	}
}

// TestNullValueMakesFieldOptional verifies that a property whose value is null
// in one sample is treated as optional (not required).
func TestNullValueMakesFieldOptional(t *testing.T) {
	generator := New()

	// Sample 1: "value" is a nested object
	err := generator.AddSample(`{
		"structure": "item",
		"type": "coordinates",
		"value": {
			"latitude": {"structure": "item", "type": "float", "value": 48.61319435},
			"longitude": {"structure": "item", "type": "float", "value": 7.6940764}
		}
	}`)
	if err != nil {
		t.Fatalf("Failed to add sample 1: %v", err)
	}

	// Sample 2: "value" is null
	err = generator.AddSample(`{
		"structure": "item",
		"type": "coordinates",
		"value": null
	}`)
	if err != nil {
		t.Fatalf("Failed to add sample 2: %v", err)
	}

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	if err = json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// "structure" and "type" appear in both samples → required
	isRequired := func(field string) bool {
		for _, r := range schema.Required {
			if r == field {
				return true
			}
		}
		return false
	}
	if !isRequired("structure") {
		t.Errorf("Expected 'structure' to be required")
	}
	if !isRequired("type") {
		t.Errorf("Expected 'type' to be required")
	}
	// "value" is null in sample 2 → must be optional
	if isRequired("value") {
		t.Errorf("Expected 'value' to be optional (null in one sample)")
	}

	// "value" must still appear in properties
	if schema.Properties["value"] == nil {
		t.Errorf("Expected 'value' to appear in properties")
	}
}

// TestNullValueNoTypeContamination verifies that null does not pollute the
// inferred type of a field (the field should keep its non-null type).
func TestNullValueNoTypeContamination(t *testing.T) {
	generator := New()
	generator.AddSample(`{"name": "Alice", "score": 42}`)
	generator.AddSample(`{"name": null, "score": 99}`)

	schemaJSON, _ := generator.Generate()
	var schema Schema
	json.Unmarshal([]byte(schemaJSON), &schema)

	// "name" should be type "string", not ["null", "string"]
	if schema.Properties["name"].Type != "string" {
		t.Errorf("Expected 'name' type to be 'string', got %v", schema.Properties["name"].Type)
	}
	// "name" is null in one sample → optional
	for _, r := range schema.Required {
		if r == "name" {
			t.Errorf("Expected 'name' to be optional (was null in one sample)")
		}
	}
}

// TestConstSameValue verifies that a field with identical values across all
// samples gets a "const" annotation in the schema.
func TestConstSameValue(t *testing.T) {
	generator := New()
	generator.AddSample(`{"structure": "item", "type": "coordinates", "id": 1}`)
	generator.AddSample(`{"structure": "item", "type": "coordinates", "id": 2}`)
	generator.AddSample(`{"structure": "item", "type": "coordinates", "id": 3}`)

	schemaJSON, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schema Schema
	if err = json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// "structure" is always "item" → must have const
	structureProp := schema.Properties["structure"]
	if structureProp.Const == nil {
		t.Errorf("Expected 'structure' to have const, got nil")
	} else if structureProp.Const != "item" {
		t.Errorf("Expected 'structure' const to be 'item', got %v", structureProp.Const)
	}

	// "type" is always "coordinates" → must have const
	typeProp := schema.Properties["type"]
	if typeProp.Const == nil {
		t.Errorf("Expected 'type' to have const, got nil")
	} else if typeProp.Const != "coordinates" {
		t.Errorf("Expected 'type' const to be 'coordinates', got %v", typeProp.Const)
	}

	// "id" varies → must NOT have const
	idProp := schema.Properties["id"]
	if idProp.Const != nil {
		t.Errorf("Expected 'id' to have no const (values differ), got %v", idProp.Const)
	}
}

// TestConstDifferentValues verifies that fields with different values do not
// get a "const" annotation.
func TestConstDifferentValues(t *testing.T) {
	generator := New()
	generator.AddSample(`{"name": "Alice", "active": true}`)
	generator.AddSample(`{"name": "Bob", "active": false}`)

	schemaJSON, _ := generator.Generate()
	var schema Schema
	json.Unmarshal([]byte(schemaJSON), &schema)

	if schema.Properties["name"].Const != nil {
		t.Errorf("Expected 'name' to have no const (values differ), got %v", schema.Properties["name"].Const)
	}
	if schema.Properties["active"].Const != nil {
		t.Errorf("Expected 'active' to have no const (values differ), got %v", schema.Properties["active"].Const)
	}
}

// TestConstWithNullSamples verifies that const is still detected correctly when
// some samples have null for the field (null observations are ignored).
func TestConstWithNullSamples(t *testing.T) {
	generator := New()
	generator.AddSample(`{"kind": "sensor", "value": 10}`)
	generator.AddSample(`{"kind": null, "value": 20}`)
	generator.AddSample(`{"kind": "sensor", "value": 30}`)

	schemaJSON, _ := generator.Generate()
	var schema Schema
	json.Unmarshal([]byte(schemaJSON), &schema)

	// "kind" is "sensor" whenever non-null → const = "sensor"
	kindProp := schema.Properties["kind"]
	if kindProp.Const == nil {
		t.Errorf("Expected 'kind' to have const 'sensor', got nil")
	} else if kindProp.Const != "sensor" {
		t.Errorf("Expected 'kind' const to be 'sensor', got %v", kindProp.Const)
	}
	// "kind" is null in one sample → optional
	for _, r := range schema.Required {
		if r == "kind" {
			t.Errorf("Expected 'kind' to be optional (was null in one sample)")
		}
	}
}

func TestGenerateToWriter(t *testing.T) {
	generator := New()
	generator.AddSample(`{"name": "Alice", "age": 30}`)

	var buf bytes.Buffer
	if err := generator.GenerateTo(&buf); err != nil {
		t.Fatalf("GenerateTo failed: %v", err)
	}

	var schema Schema
	if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
		t.Fatalf("Failed to unmarshal GenerateTo output: %v", err)
	}
	if schema.Properties["name"].Type != "string" {
		t.Errorf("Expected name to be string, got %v", schema.Properties["name"].Type)
	}
}

func TestGenerateToWriterNoSamples(t *testing.T) {
	generator := New()
	var buf bytes.Buffer
	if err := generator.GenerateTo(&buf); err == nil {
		t.Errorf("Expected error when no samples added, got nil")
	}
}

func TestWithIndentGenerate(t *testing.T) {
	generator := New(WithIndent("  "))
	generator.AddSample(`{"name": "Alice"}`)

	result, err := generator.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if !strings.Contains(result, "\n") {
		t.Errorf("Expected indented output to contain newlines, got: %s", result)
	}
	if !strings.Contains(result, "  ") {
		t.Errorf("Expected indented output to contain two-space indent, got: %s", result)
	}
}

func TestWithIndentGenerateTo(t *testing.T) {
	generator := New(WithIndent("\t"))
	generator.AddSample(`{"name": "Alice"}`)

	var buf bytes.Buffer
	if err := generator.GenerateTo(&buf); err != nil {
		t.Fatalf("GenerateTo failed: %v", err)
	}
	if !strings.Contains(buf.String(), "\t") {
		t.Errorf("Expected tab-indented output, got: %s", buf.String())
	}
}

func TestGenerateCompactByDefault(t *testing.T) {
	generator := New()
	generator.AddSample(`{"name": "Alice"}`)

	result, err := generator.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if strings.Contains(result, "\n") {
		t.Errorf("Expected compact output (no newlines) by default, got: %s", result)
	}
}
