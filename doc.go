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
// # Format Detection
//
// The library uses a unified format detection mechanism. All formats (built-in and custom)
// are detected using FormatDetector functions: func(string) bool
//
// Built-in formats are automatically registered:
//   - date-time: ISO 8601 datetime strings (e.g., "2023-01-15T10:30:00Z")
//   - email: Email addresses (e.g., "user@example.com")
//   - uuid: UUIDs v1-v5 (e.g., "550e8400-e29b-41d4-a716-446655440000")
//   - ipv4: IPv4 addresses (e.g., "192.168.1.1")
//   - ipv6: IPv6 addresses (e.g., "2001:0db8:85a3::8a2e:0370:7334")
//   - uri: URLs with HTTP, HTTPS, FTP, FTPS schemes (e.g., "https://example.com")
//
// Custom formats can be registered using WithCustomFormat:
//
//	isHexColor := func(s string) bool {
//	    return len(s) == 7 && s[0] == '#' && /* ... */
//	}
//	generator := jsonschema.New(
//	    jsonschema.WithCustomFormat("hex-color", isHexColor),
//	)
//	generator.AddSample(`{"color": "#FF5733"}`)
//	schema, _ := generator.Generate()
//	// Result: color has type "string" and format "hex-color"
//
// Built-in formats can be disabled if you want full control:
//
//	generator := jsonschema.New(jsonschema.WithoutBuiltInFormats())
//
// All string values in a field must match for a format to be applied.
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
// # Schema Versions
//
// You can choose which JSON Schema draft version to generate. The library supports
// Draft 06 and Draft 07 (default):
//
//	// Generate Draft 06 schema
//	generator := jsonschema.New(jsonschema.WithSchemaVersion(jsonschema.Draft06))
//	generator.AddSample(`{"name": "John", "age": 30}`)
//	schema, _ := generator.Generate()
//	// Result: $schema is "http://json-schema.org/draft-06/schema#"
//
//	// Generate Draft 07 schema (default)
//	generator2 := jsonschema.New()  // or WithSchemaVersion(Draft07)
//	generator2.AddSample(`{"name": "Jane", "age": 25}`)
//	schema2, _ := generator2.Generate()
//	// Result: $schema is "http://json-schema.org/draft-07/schema#"
//
// For the features used by this library, Draft 06 and Draft 07 are functionally equivalent.
// The main difference is the $schema URL in the output.
//
// If you need to create an empty schema with a specific version (without using the Generator),
// use NewSchemaWithVersion:
//
//	schema := jsonschema.NewSchemaWithVersion(jsonschema.Draft06)
//	// schema.Schema is "http://json-schema.org/draft-06/schema#"
//
// Note: For schema inference from samples, always use Generator with New() and AddSample().
// The NewSchemaWithVersion function is only for creating empty schema structures.
//
// # Examples
//
// The library can capture the first observed value for each field and include it as an
// example in the generated schema. By default, example capturing is disabled.
// To enable it, use the WithExamples option:
//
//	generator := jsonschema.New(jsonschema.WithExamples())
//	generator.AddSample(`{"name": "John"}`)
//	// Result: name field will have example: "John"
//
// # Lazy Schema Building
//
// The schema is built on demand when Generate() or GetCurrentSchema() is called, not
// after every AddSample(). This avoids expensive O(N) tree traversals during bulk
// ingestion. The result is cached and reused until the next sample invalidates it.
//
// If you parse JSON yourself (e.g. with json.Decoder), use AddParsedSample to skip
// re-parsing inside the library:
//
//	dec := json.NewDecoder(reader)
//	for dec.More() {
//	    var v interface{}
//	    dec.Decode(&v)
//	    generator.AddParsedSample(v)  // no double-parse
//	}
//	schema, _ := generator.Generate()
//
// You can still inspect the evolving schema at any point via GetCurrentSchema():
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
//   - Schema is built lazily (on demand) and cached between calls
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
//   - No validation constraints (min/max, length)
//   - Only JSON Schema draft-07 output format
//   - No enum detection for fields with limited value sets
//   - Sample count tracking is approximate after loading schemas
//
// # Performance Considerations
//
//   - Schema is built lazily — no overhead during sample ingestion.
//   - Use AddParsedSample when you already hold a decoded interface{} value to
//     avoid a second json.Unmarshal call.
//   - For very large sample sets, set a limit with WithMaxSamples to cap
//     the number of samples processed.
//
// # Thread Safety
//
// The Generator is thread-safe. AddSample, AddParsedSample, Generate, and
// GetCurrentSchema can all be called concurrently from multiple goroutines.
package jsonschema
