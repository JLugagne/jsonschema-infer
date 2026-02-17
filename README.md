# jsonschema-infer

A Go library for inferring JSON Schema from JSON samples. This library analyzes multiple JSON documents and automatically generates a JSON Schema that describes their structure, types, and patterns.

## Features

- ✅ **Infer basic types**: string, boolean, number, integer
- ✅ **Detect optional fields**: tracks which fields appear in all samples vs. some samples
- ✅ **Handle arrays**: treats all array items as the same type and infers their schema
- ✅ **Nested objects**: full support for deeply nested object structures
- ✅ **Arrays of objects**: infers schemas for complex array items with optional fields
- ✅ **Unified format detection**: all formats detected using the same mechanism (FormatDetector functions)
- ✅ **Built-in formats**: datetime (ISO 8601), email, UUID, IPv4, IPv6, and URL (HTTP/HTTPS/FTP/FTPS)
- ✅ **Custom format detectors**: register user-defined format detection functions
- ✅ **Configurable**: disable built-in formats for full control
- ✅ **Predefined types**: configure specific field types (e.g., `created_at` as DateTime)
- ✅ **Flexible root types**: supports objects, arrays, and primitives at root level
- ✅ **Incremental updates**: schema evolves after each sample is added
- ✅ **Load/Resume**: load previously generated schemas and continue adding samples
- ✅ **Schema versions**: support for Draft 06 and Draft 07 (default)
- ✅ **Examples**: optional example capturing (disabled by default)
- ✅ **Tree-based architecture**: clean recursive structure for maintainability
- ✅ **Max samples limit**: optionally limit the number of samples to process

## Requirements

- Go 1.25 or higher

## Installation

```bash
go get github.com/JLugagne/jsonschema-infer
```

## Documentation

- **[Usage Guide](USAGE.md)** - Detailed examples and best practices
- **[API Documentation](https://pkg.go.dev/github.com/JLugagne/jsonschema-infer)** - Complete API reference
- **[Architecture](ARCHITECTURE.md)** - Internal design and algorithms
- **[Examples](examples/)** - Runnable example programs

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/JLugagne/jsonschema-infer"
)

func main() {
    // Create a new generator
    generator := jsonschema.New()

    // Add JSON samples
    generator.AddSample(`{"name": "John", "age": 30, "active": true}`)
    generator.AddSample(`{"name": "Jane", "age": 25, "active": false}`)
    generator.AddSample(`{"name": "Bob", "age": 35}`)

    // Generate the schema
    schema, err := generator.Generate()
    if err != nil {
        panic(err)
    }

    fmt.Println(schema)
}
```

**Output:**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "name": {
      "type": "string"
    },
    "age": {
      "type": "integer"
    },
    "active": {
      "type": "boolean"
    }
  },
  "required": ["name", "age"]
}
```

Note: `active` is not in `required` because it doesn't appear in all samples.

### Predefined Types

Configure specific fields to have predefined types:

```go
generator := jsonschema.New(
    jsonschema.WithPredefined("created_at", jsonschema.DateTime),
    jsonschema.WithPredefined("updated_at", jsonschema.DateTime),
)

generator.AddSample(`{"id": 1, "created_at": "2023-01-15T10:30:00Z"}`)
generator.AddSample(`{"id": 2, "created_at": "2023-02-20T14:45:00Z"}`)

schema, _ := generator.Generate()
```

Available predefined types:
- `DateTime` - string with date-time format
- `String` - string type
- `Boolean` - boolean type
- `Number` - number type
- `Integer` - integer type
- `Array` - array type
- `Object` - object type

### Examples

The library can capture the first observed value as an `example` for each field:

```go
// Enable example capturing
generator := jsonschema.New(jsonschema.WithExamples())
```

By default, example capturing is disabled to save memory and keep schemas compact.

### Arrays of Objects

The library handles arrays of objects and detects optional fields within array items:

```go
generator := jsonschema.New()

generator.AddSample(`{
    "users": [
        {"id": 1, "name": "John", "email": "john@example.com"},
        {"id": 2, "name": "Jane"}
    ]
}`)

generator.AddSample(`{
    "users": [
        {"id": 3, "name": "Bob", "email": "bob@example.com"}
    ]
}`)

schema, _ := generator.Generate()
```

The resulting schema will show that `email` is optional in the array items since it doesn't appear in all objects.

### Load and Resume

Load a previously generated schema and continue adding samples:

```go
// Generate initial schema
generator1 := jsonschema.New()
generator1.AddSample(`{"name": "John", "age": 30}`)
schemaJSON, _ := generator1.Generate()

// Later, load the schema and add more samples
generator2 := jsonschema.New()
err := generator2.Load(schemaJSON)
if err != nil {
    panic(err)
}

// Add new samples with additional fields
generator2.AddSample(`{"name": "Jane", "age": 25, "email": "jane@example.com"}`)

// Generate updated schema
updatedSchema, _ := generator2.Generate()
```

### Get Current Schema

Retrieve the current schema as a `Schema` object after any sample:

```go
generator := jsonschema.New()
generator.AddSample(`{"name": "John"}`)

// Get the current schema as an object (not JSON string)
schema := generator.GetCurrentSchema()

// Access properties
fmt.Println(schema.Type) // "object"
fmt.Println(schema.Properties["name"].Type) // "string"
```

## Building and Testing

### Build

```bash
go build
```

Or use the Makefile:

```bash
make build
```

### Test

```bash
go test -v
```

Or use the Makefile:

```bash
make test
```

### Test with Coverage

```bash
make test-coverage
```

This generates `coverage.html` which you can open in a browser.

## Architecture

The library uses a tree-based recursive architecture:

- **`SchemaNode`**: Each node represents a part of the JSON structure
  - Handles only primitives (string, number, boolean, null)
  - Delegates to child nodes for complex types (arrays, objects)
  - Accumulates observations across all samples

- **Incremental Updates**: Schema is rebuilt after each `AddSample()` call
  - No need to wait until all samples are collected
  - Can inspect schema evolution at any point

- **Optional Field Detection**: Tracks how many times each field appears
  - Fields appearing in all samples → required
  - Fields appearing in some samples → optional

## More Examples

See the [examples/](examples/) directory for runnable examples:

- **[basic](examples/basic/)** - Basic type inference and optional fields
- **[arrays](examples/arrays/)** - Arrays of objects with optional fields
- **[datetime](examples/datetime/)** - Automatic datetime detection
- **[predefined](examples/predefined/)** - Configuring predefined types
- **[load_resume](examples/load_resume/)** - Loading and resuming schemas
- **[nested](examples/nested/)** - Deeply nested structures
- **[incremental](examples/incremental/)** - Watching schema evolution

Run all examples:

```bash
cd examples
./run-examples.sh
```

## Comparison with Other Libraries

This library is unique in the Go ecosystem for sample-based JSON schema inference. Similar functionality exists in other languages:

- **Python**: [genson](https://github.com/wolverdude/genson) - similar approach
- **JavaScript**: [@jsonhero/schema-infer](https://www.npmjs.com/package/@jsonhero/schema-infer)
- **Online**: [jsonschema.net](https://jsonschema.net) - web-based tool

Key advantages of jsonschema-infer:
- ✅ Pure Go implementation
- ✅ Incremental schema updates
- ✅ Load/resume capability
- ✅ Tree-based recursive architecture
- ✅ Optional field frequency tracking

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

[Specify your license here]

## Notes

- The library uses Go's standard `encoding/json` package for JSON parsing
- All array items are treated as having the same schema (merged together)
- Multiple type detection is supported (e.g., a field that's sometimes string, sometimes number)
