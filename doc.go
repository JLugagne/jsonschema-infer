// Package jsonschema provides automatic JSON Schema generation from JSON samples.
//
// This library analyzes multiple JSON documents and automatically generates a JSON Schema
// that describes their structure, types, and patterns. It uses an incremental, tree-based
// approach to build schemas that evolve as more samples are observed.
//
// # Overview
//
// The jsonschema package allows you to:
//   - Infer JSON schemas from one or more JSON samples
//   - Detect optional vs required fields based on sample frequency
//   - Automatically identify datetime patterns in strings
//   - Configure predefined types for specific fields
//   - Load and resume from previously generated schemas
//   - Handle complex nested structures with arrays and objects
//
// # Basic Usage
//
// The simplest way to use the library is to create a Generator, add JSON samples,
// and generate the resulting schema:
//
//	generator := jsonschema.New()
//	generator.AddSample(`{"name": "John", "age": 30}`)
//	generator.AddSample(`{"name": "Jane", "age": 25}`)
//	schema, err := generator.Generate()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(schema)
//
// # Type Inference
//
// The library automatically infers the following JSON Schema types:
//   - string: for text values
//   - integer: for whole numbers
//   - number: for floating-point numbers
//   - boolean: for true/false values
//   - array: for arrays (with item schema inference)
//   - object: for objects (with recursive property inference)
//   - null: for null values
//
// When a field has multiple types across samples, the schema will include all
// observed types as an array.
//
// # Optional vs Required Fields
//
// Fields are marked as required only if they appear in ALL samples. If a field
// appears in some but not all samples, it will be defined in the schema but not
// listed in the "required" array:
//
//	generator := jsonschema.New()
//	generator.AddSample(`{"name": "John", "age": 30}`)
//	generator.AddSample(`{"name": "Jane"}`)  // no age
//	schema, _ := generator.Generate()
//	// Result: only "name" is required, "age" is optional
//
// # Arrays
//
// Arrays are handled by merging all observed items into a single schema. The library
// treats all array items as having the same type and generates a unified schema:
//
//	generator := jsonschema.New()
//	generator.AddSample(`{"tags": ["go", "json"]}`)
//	generator.AddSample(`{"tags": ["schema", "inference"]}`)
//	schema, _ := generator.Generate()
//	// Result: tags is array of strings
//
// For arrays of objects, the library detects which fields are required across all
// observed array items:
//
//	generator := jsonschema.New()
//	generator.AddSample(`{"users": [{"id": 1, "name": "John", "email": "john@example.com"}]}`)
//	generator.AddSample(`{"users": [{"id": 2, "name": "Jane"}]}`)  // no email
//	schema, _ := generator.Generate()
//	// Result: id and name are required in array items, email is optional
//
// # DateTime Detection
//
// The library automatically detects ISO 8601 datetime strings and marks them with
// the "date-time" format:
//
//	generator := jsonschema.New()
//	generator.AddSample(`{"created_at": "2023-01-15T10:30:00Z"}`)
//	generator.AddSample(`{"created_at": "2023-02-20T14:45:00Z"}`)
//	schema, _ := generator.Generate()
//	// Result: created_at has type "string" and format "date-time"
//
// # Predefined Types
//
// You can configure specific fields to have predefined types, which override
// automatic type inference:
//
//	generator := jsonschema.New(
//	    jsonschema.WithPredefined("created_at", jsonschema.DateTime),
//	    jsonschema.WithPredefined("updated_at", jsonschema.DateTime),
//	)
//	generator.AddSample(`{"created_at": "2023-01-15T10:30:00Z"}`)
//	schema, _ := generator.Generate()
//	// Result: created_at is guaranteed to be string with date-time format
//
// Available predefined types:
//   - DateTime: string with date-time format
//   - String: string type
//   - Boolean: boolean type
//   - Number: number type
//   - Integer: integer type
//   - Array: array type
//   - Object: object type
//
// # Incremental Schema Updates
//
// The schema is updated incrementally after each sample is added. You can inspect
// the current schema at any time using GetCurrentSchema():
//
//	generator := jsonschema.New()
//	generator.AddSample(`{"name": "John"}`)
//	schema1 := generator.GetCurrentSchema()
//	fmt.Printf("After 1 sample: %+v\n", schema1)
//
//	generator.AddSample(`{"name": "Jane", "age": 25}`)
//	schema2 := generator.GetCurrentSchema()
//	fmt.Printf("After 2 samples: %+v\n", schema2)
//
// # Loading and Resuming
//
// You can load a previously generated schema and continue adding samples to it.
// This is useful for evolving schemas over time:
//
//	// Initial schema generation
//	generator1 := jsonschema.New()
//	generator1.AddSample(`{"name": "John", "age": 30}`)
//	schemaJSON, _ := generator1.Generate()
//
//	// Later, load and continue
//	generator2 := jsonschema.New()
//	err := generator2.Load(schemaJSON)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	generator2.AddSample(`{"name": "Jane", "age": 25, "email": "jane@example.com"}`)
//	updatedSchema, _ := generator2.Generate()
//	// Result: schema now includes optional "email" field
//
// # Nested Structures
//
// The library fully supports nested objects and arrays at any depth:
//
//	generator := jsonschema.New()
//	generator.AddSample(`{
//	    "company": {
//	        "name": "Acme Corp",
//	        "employees": [
//	            {"name": "Alice", "role": "Engineer"},
//	            {"name": "Bob", "role": "Designer"}
//	        ]
//	    }
//	}`)
//	schema, _ := generator.Generate()
//	// Result: fully nested schema with company.employees as array of objects
//
// # Architecture
//
// The library uses a tree-based recursive architecture:
//   - Each SchemaNode represents a part of the JSON structure
//   - Nodes handle only primitives (string, number, boolean, null)
//   - Complex types (arrays, objects) delegate to child nodes
//   - All observations are accumulated across samples
//   - Schema is rebuilt after each AddSample() call
//
// This design keeps the code maintainable by limiting complexity to simple
// primitives within well-defined scopes.
//
// # Requirements
//
// Go 1.25 or higher is required:
//
//	go build
//	go test
//
// The library uses Go's standard encoding/json package for JSON parsing.
//
// # Error Handling
//
// Methods that can fail return errors:
//   - AddSample() returns an error if the JSON is invalid
//   - Generate() returns an error if no samples have been added
//   - Load() returns an error if the schema JSON is invalid or not an object schema
//
// Always check errors to ensure reliable schema generation.
//
// # Limitations
//
// Current limitations:
//   - All array items are treated as having the same schema (no tuple support)
//   - No validation constraints (min/max, length, patterns beyond datetime)
//   - Only JSON Schema draft-07 output format
//   - No enum detection for fields with limited value sets
//   - Sample count tracking is approximate after loading schemas
//
// # Performance Considerations
//
// The library rebuilds the entire schema after each AddSample() call. For large
// numbers of samples, this may have performance implications. Future versions
// may add batch processing options.
//
// # Thread Safety
//
// The Generator is not thread-safe. If you need to add samples from multiple
// goroutines, use appropriate synchronization (e.g., sync.Mutex).
package jsonschema
