# Examples

This directory contains runnable examples demonstrating various features of the jsonschema-infer library.

## Running Examples

All examples require Go 1.25+:

```bash
cd examples/basic
go run main.go
```

Or use the provided run script:

```bash
./run-examples.sh
```

## Available Examples

### 1. Basic Type Inference (`basic/`)

Demonstrates basic type inference and optional field detection.

**Features shown:**
- Creating a generator
- Adding JSON samples
- Generating a schema
- Required vs optional fields

```bash
cd basic && go run main.go
```

### 2. Arrays of Objects (`arrays/`)

Shows how arrays of objects are handled, with optional field detection within array items.

**Features shown:**
- Arrays of objects
- Optional fields in array items
- Type inference for array item properties

```bash
cd arrays && go run main.go
```

### 3. DateTime Detection (`datetime/`)

Demonstrates automatic ISO 8601 datetime string detection.

**Features shown:**
- Automatic datetime pattern matching
- `date-time` format in schema
- ISO 8601 validation

```bash
cd datetime && go run main.go
```

### 4. Predefined Types (`predefined/`)

Shows how to configure specific fields with predefined types.

**Features shown:**
- Using `WithPredefined()` option
- Available predefined types
- Type enforcement

```bash
cd predefined && go run main.go
```

### 5. Load and Resume (`load_resume/`)

Demonstrates loading a previously generated schema and continuing to add samples.

**Features shown:**
- Generating initial schema
- Loading schema with `Load()`
- Adding new samples to loaded schema
- Schema evolution over time

```bash
cd load_resume && go run main.go
```

### 6. Nested Structures (`nested/`)

Shows handling of deeply nested objects and arrays.

**Features shown:**
- Nested objects
- Arrays within objects
- Objects within arrays
- Deep nesting (multiple levels)

```bash
cd nested && go run main.go
```

### 7. Incremental Updates (`incremental/`)

Demonstrates how the schema evolves incrementally after each sample.

**Features shown:**
- `GetCurrentSchema()` method
- Schema inspection after each sample
- Field appearance tracking

```bash
cd incremental && go run main.go
```

## Example Output

Here's what you can expect from the basic example:

```
=== Basic Type Inference Example ===

Adding samples:
  Sample 1: {"name": "John", "age": 30, "active": true}
  Sample 2: {"name": "Jane", "age": 25, "active": false}
  Sample 3: {"name": "Bob", "age": 35}

Generated Schema:
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "active": {
      "type": "boolean"
    },
    "age": {
      "type": "integer"
    },
    "name": {
      "type": "string"
    }
  },
  "required": ["age", "name"]
}

Note: 'active' is not in 'required' because it doesn't appear in all samples.
```

## Building Your Own Examples

Use these examples as templates for your own use cases:

```go
package main

import (
    "fmt"
    "log"
    "github.com/JLugagne/jsonschema-infer"
)

func main() {
    generator := jsonschema.New()

    // Add your JSON samples
    err := generator.AddSample(`{"your": "data"}`)
    if err != nil {
        log.Fatal(err)
    }

    // Generate schema
    schema, err := generator.Generate()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(schema)
}
```

## Tips

1. **Start simple**: Begin with the `basic/` example to understand core concepts
2. **Use representative data**: Provide diverse samples that cover all variations
3. **Check errors**: Always handle errors from `AddSample()` and `Generate()`
4. **Inspect incrementally**: Use `GetCurrentSchema()` to see schema evolution
5. **Configure types**: Use `WithPredefined()` for fields with known types

## Further Reading

- [Usage Guide](../USAGE.md) - Detailed usage patterns and best practices
- [API Documentation](../doc.go) - Complete API reference
- [Architecture](../ARCHITECTURE.md) - Internal design and algorithms
- [README](../README.md) - Project overview
