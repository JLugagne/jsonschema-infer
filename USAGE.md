# Usage Guide

This guide provides detailed examples and best practices for using the jsonschema-infer library.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [Common Patterns](#common-patterns)
- [Advanced Usage](#advanced-usage)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Installation

```bash
go get github.com/JLugagne/jsonschema-infer
```

**Requirements:**
- Go 1.25 or higher

## Quick Start

### Generate a Simple Schema

```go
package main

import (
    "fmt"
    "log"
    "github.com/JLugagne/jsonschema-infer"
)

func main() {
    // Create a generator
    generator := jsonschema.New()

    // Add JSON samples
    err := generator.AddSample(`{"name": "John", "age": 30}`)
    if err != nil {
        log.Fatal(err)
    }

    err = generator.AddSample(`{"name": "Jane", "age": 25}`)
    if err != nil {
        log.Fatal(err)
    }

    // Generate the schema
    schema, err := generator.Generate()
    if err != nil {
        log.Fatal(err)
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
    "name": {"type": "string"},
    "age": {"type": "integer"}
  },
  "required": ["name", "age"]
}
```

## Core Concepts

### 1. Type Inference

The library automatically infers types from the JSON values:

```go
generator := jsonschema.New()

// String type
generator.AddSample(`{"field": "text value"}`)

// Integer type
generator.AddSample(`{"count": 42}`)

// Number type (floating-point)
generator.AddSample(`{"price": 19.99}`)

// Boolean type
generator.AddSample(`{"active": true}`)

// Array type
generator.AddSample(`{"tags": ["go", "json"]}`)

// Object type
generator.AddSample(`{"user": {"name": "John"}}`)
```

### 2. Required vs Optional Fields

Fields appearing in **all** samples are marked as required:

```go
generator := jsonschema.New()

generator.AddSample(`{"id": 1, "name": "John", "email": "john@example.com"}`)
generator.AddSample(`{"id": 2, "name": "Jane", "email": "jane@example.com"}`)
generator.AddSample(`{"id": 3, "name": "Bob"}`)  // no email

schema, _ := generator.Generate()
```

**Result:**
```json
{
  "type": "object",
  "properties": {
    "id": {"type": "integer"},
    "name": {"type": "string"},
    "email": {"type": "string"}
  },
  "required": ["id", "name"]
}
```

`email` is defined but **not required** (appears in 2/3 samples).

### 3. Array Handling

All array items are merged into a single schema:

```go
generator := jsonschema.New()

generator.AddSample(`{"numbers": [1, 2, 3]}`)
generator.AddSample(`{"numbers": [10, 20]}`)

schema, _ := generator.Generate()
```

**Result:**
```json
{
  "type": "object",
  "properties": {
    "numbers": {
      "type": "array",
      "items": {"type": "integer"}
    }
  },
  "required": ["numbers"]
}
```

### 4. Arrays of Objects

For arrays containing objects, the library tracks which fields are required:

```go
generator := jsonschema.New()

generator.AddSample(`{
  "products": [
    {"id": 1, "name": "Product A", "price": 10.50},
    {"id": 2, "name": "Product B"}
  ]
}`)

generator.AddSample(`{
  "products": [
    {"id": 3, "name": "Product C", "price": 20.00}
  ]
}`)

schema, _ := generator.Generate()
```

**Result:** `id` and `name` are required in array items, `price` is optional.

### 5. Pattern Detection

The library automatically detects common string patterns and applies appropriate formats:

#### DateTime Detection (ISO 8601)

```go
generator := jsonschema.New()

generator.AddSample(`{"created_at": "2023-01-15T10:30:00Z"}`)
generator.AddSample(`{"created_at": "2023-02-20T14:45:00Z"}`)

schema, _ := generator.Generate()
```

**Result:**
```json
{
  "type": "object",
  "properties": {
    "created_at": {
      "type": "string",
      "format": "date-time"
    }
  },
  "required": ["created_at"]
}
```

#### Email Detection

```go
generator := jsonschema.New()
generator.AddSample(`{"email": "user@example.com"}`)
generator.AddSample(`{"email": "admin@test.org"}`)
schema, _ := generator.Generate()
// Result: email field has format "email"
```

#### UUID Detection

```go
generator := jsonschema.New()
generator.AddSample(`{"id": "550e8400-e29b-41d4-a716-446655440000"}`)
generator.AddSample(`{"id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8"}`)
schema, _ := generator.Generate()
// Result: id field has format "uuid"
```

#### IP Address Detection

```go
// IPv4
generator := jsonschema.New()
generator.AddSample(`{"ip": "192.168.1.1"}`)
generator.AddSample(`{"ip": "10.0.0.1"}`)
schema, _ := generator.Generate()
// Result: ip field has format "ipv4"

// IPv6
generator := jsonschema.New()
generator.AddSample(`{"ip": "2001:0db8:85a3:0000:0000:8a2e:0370:7334"}`)
generator.AddSample(`{"ip": "fe80::1"}`)
schema, _ := generator.Generate()
// Result: ip field has format "ipv6"
```

#### URL Detection

```go
generator := jsonschema.New()
generator.AddSample(`{"website": "https://example.com"}`)
generator.AddSample(`{"website": "http://test.org/path"}`)
generator.AddSample(`{"website": "ftp://files.example.com/data"}`)
schema, _ := generator.Generate()
// Result: website field has format "uri"
```

**Supported URL schemes:** HTTP, HTTPS, FTP, FTPS

**Pattern Priority:** Patterns are checked in order of specificity:
1. datetime (ISO 8601)
2. email
3. uuid
4. ipv6
5. ipv4
6. uri (URL)

All string values in a field must match the pattern for it to be applied.

## Common Patterns

### Pattern 1: Processing Multiple Files

```go
func generateSchemaFromFiles(filePaths []string) (string, error) {
    generator := jsonschema.New()

    for _, path := range filePaths {
        data, err := os.ReadFile(path)
        if err != nil {
            return "", fmt.Errorf("reading %s: %w", path, err)
        }

        if err := generator.AddSample(string(data)); err != nil {
            return "", fmt.Errorf("processing %s: %w", path, err)
        }
    }

    return generator.Generate()
}
```

### Pattern 2: API Response Analysis

```go
func analyzeAPIResponses(endpoint string) (string, error) {
    generator := jsonschema.New()

    // Fetch multiple responses
    for i := 0; i < 10; i++ {
        resp, err := http.Get(endpoint)
        if err != nil {
            return "", err
        }
        defer resp.Body.Close()

        body, err := io.ReadAll(resp.Body)
        if err != nil {
            return "", err
        }

        if err := generator.AddSample(string(body)); err != nil {
            return "", err
        }
    }

    return generator.Generate()
}
```

### Pattern 3: Streaming JSON Lines

```go
func processJSONLines(reader io.Reader) (string, error) {
    generator := jsonschema.New()
    scanner := bufio.NewScanner(reader)

    for scanner.Scan() {
        line := scanner.Text()
        if err := generator.AddSample(line); err != nil {
            return "", err
        }
    }

    if err := scanner.Err(); err != nil {
        return "", err
    }

    return generator.Generate()
}
```

### Pattern 4: Incremental Schema Inspection

```go
func inspectSchemaEvolution(samples []string) {
    generator := jsonschema.New()

    for i, sample := range samples {
        generator.AddSample(sample)

        // Inspect schema after each sample
        schema := generator.GetCurrentSchema()
        fmt.Printf("After sample %d:\n", i+1)
        fmt.Printf("  Properties: %d\n", len(schema.Properties))
        fmt.Printf("  Required: %d\n", len(schema.Required))
    }
}
```

## Advanced Usage

### Predefined Types

Configure specific fields to have predetermined types:

```go
generator := jsonschema.New(
    jsonschema.WithPredefined("created_at", jsonschema.DateTime),
    jsonschema.WithPredefined("updated_at", jsonschema.DateTime),
    jsonschema.WithPredefined("user_id", jsonschema.Integer),
)

generator.AddSample(`{
    "created_at": "2023-01-15T10:30:00Z",
    "updated_at": "2023-01-15T10:30:00Z",
    "user_id": 123
}`)

schema, _ := generator.Generate()
```

**Available Predefined Types:**
- `jsonschema.DateTime` - string with date-time format
- `jsonschema.String` - string type
- `jsonschema.Boolean` - boolean type
- `jsonschema.Number` - number type
- `jsonschema.Integer` - integer type
- `jsonschema.Array` - array type
- `jsonschema.Object` - object type

### Schema Versions

Choose which JSON Schema draft version to generate:

```go
// Generate Draft 06 schema
generator := jsonschema.New(
    jsonschema.WithSchemaVersion(jsonschema.Draft06),
)

generator.AddSample(`{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
}`)

schema, _ := generator.Generate()
// Output: {"$schema":"http://json-schema.org/draft-06/schema#", ...}
```

**Available Schema Versions:**
- `jsonschema.Draft06` - JSON Schema Draft 06 (`http://json-schema.org/draft-06/schema#`)
- `jsonschema.Draft07` - JSON Schema Draft 07 (`http://json-schema.org/draft-07/schema#`) - **Default**

**Default Behavior:** If you don't specify a schema version, Draft 07 is used:

```go
// These are equivalent:
generator1 := jsonschema.New()
generator2 := jsonschema.New(jsonschema.WithSchemaVersion(jsonschema.Draft07))
```

**Why Use Draft 06?**
- Compatibility with older systems or validators
- Some tools or libraries may only support Draft 06

**Note:** For the features used by this library (basic types, arrays, objects, format, required), there's no functional difference between Draft 06 and Draft 07. The main difference is the `$schema` URL in the output.

**Creating Empty Schemas:**

If you need an empty schema structure with a specific version (without using the Generator), use `NewSchemaWithVersion`:

```go
// Create empty Draft 06 schema
schema := jsonschema.NewSchemaWithVersion(jsonschema.Draft06)
// schema.Schema is "http://json-schema.org/draft-06/schema#"

// Create empty Draft 07 schema
schema := jsonschema.NewSchemaWithVersion(jsonschema.Draft07)
// schema.Schema is "http://json-schema.org/draft-07/schema#"
```

**Note:** For schema inference from samples, always use `Generator` with `New()` and `AddSample()`. The `NewSchemaWithVersion` function is only for creating empty schema structures.

### Custom Format Detectors

Register user-defined format detection functions:

```go
// Define a custom format detector for hex colors
isHexColor := func(s string) bool {
    if len(s) != 7 || s[0] != '#' {
        return false
    }
    for i := 1; i < 7; i++ {
        c := s[i]
        if !((c >= '0' && c <= '9') ||
             (c >= 'a' && c <= 'f') ||
             (c >= 'A' && c <= 'F')) {
            return false
        }
    }
    return true
}

generator := jsonschema.New(
    jsonschema.WithCustomFormat("hex-color", isHexColor),
)

generator.AddSample(`{"color": "#FF5733"}`)
generator.AddSample(`{"color": "#00FF00"}`)

schema, _ := generator.Generate()
// Result: color field has type "string" and format "hex-color"
```

**Multiple Custom Formats:**

```go
isPhoneNumber := func(s string) bool {
    return len(s) > 10 && s[0] == '+'
}

generator := jsonschema.New(
    jsonschema.WithCustomFormat("hex-color", isHexColor),
    jsonschema.WithCustomFormat("phone", isPhoneNumber),
)
```

**Priority:** Custom formats are checked **after** built-in formats (date-time, email, uuid, ipv6, ipv4, uri). All string values must match for the format to be applied.

**Disabling Built-In Formats:**

If you want complete control over format detection, you can disable all built-in formats:

```go
generator := jsonschema.New(
    jsonschema.WithoutBuiltInFormats(),
    jsonschema.WithCustomFormat("my-date", myDateDetector),
)
```

This is useful when:
- You want to use different format names
- You want to implement your own validation logic
- Built-in formats are too strict/lenient for your use case

### Load and Resume

Save and load schemas to continue evolving them:

```go
// Day 1: Generate initial schema
generator1 := jsonschema.New()
generator1.AddSample(`{"id": 1, "name": "John"}`)
schemaJSON, _ := generator1.Generate()

// Save to file
os.WriteFile("schema.json", []byte(schemaJSON), 0644)

// Day 2: Load and continue
schemaData, _ := os.ReadFile("schema.json")
generator2 := jsonschema.New()
err := generator2.Load(string(schemaData))
if err != nil {
    log.Fatal(err)
}

// Add new samples with additional fields
generator2.AddSample(`{"id": 2, "name": "Jane", "email": "jane@example.com"}`)
updatedSchema, _ := generator2.Generate()

// Save updated schema
os.WriteFile("schema.json", []byte(updatedSchema), 0644)
```

### Deeply Nested Structures

The library handles arbitrary nesting depth:

```go
generator := jsonschema.New()

generator.AddSample(`{
    "organization": {
        "name": "Acme Corp",
        "departments": [
            {
                "name": "Engineering",
                "teams": [
                    {
                        "name": "Backend",
                        "members": [
                            {"name": "Alice", "role": "Senior Engineer"},
                            {"name": "Bob", "role": "Engineer"}
                        ]
                    }
                ]
            }
        ]
    }
}`)

schema, _ := generator.Generate()
```

### Multiple Types (Union Types)

When a field has different types across samples:

```go
generator := jsonschema.New()

generator.AddSample(`{"value": "text"}`)
generator.AddSample(`{"value": 42}`)
generator.AddSample(`{"value": true}`)

schema, _ := generator.Generate()
```

**Result:**
```json
{
  "type": "object",
  "properties": {
    "value": {
      "type": ["boolean", "integer", "string"]
    }
  },
  "required": ["value"]
}
```

### Array as Root

The library supports arrays at the root level:

```go
generator := jsonschema.New()

generator.AddSample(`[{"id": 1, "name": "John"}, {"id": 2, "name": "Jane"}]`)
generator.AddSample(`[{"id": 3, "name": "Bob"}]`)

schema, _ := generator.Generate()
```

**Result:**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "array",
  "items": {
    "type": "object",
    "properties": {
      "id": {"type": "integer"},
      "name": {"type": "string"}
    },
    "required": ["id", "name"]
  }
}
```

### Primitives as Root

Primitive types are also supported:

```go
generator := jsonschema.New()

generator.AddSample(`"hello"`)
generator.AddSample(`"world"`)

schema, _ := generator.Generate()
// Result: schema with type "string"
```

## Best Practices

### 1. Provide Representative Samples

Use samples that cover all variations:

```go
// Good: diverse samples
generator.AddSample(`{"status": "active", "count": 10}`)
generator.AddSample(`{"status": "inactive", "count": 0}`)
generator.AddSample(`{"status": "pending"}`)  // count is optional

// Avoid: identical samples (no useful information)
generator.AddSample(`{"status": "active"}`)
generator.AddSample(`{"status": "active"}`)
generator.AddSample(`{"status": "active"}`)
```

### 2. Handle Errors Properly

Always check for errors:

```go
// Good
if err := generator.AddSample(jsonData); err != nil {
    log.Printf("Failed to add sample: %v", err)
    continue  // or handle appropriately
}

// Bad
generator.AddSample(jsonData)  // ignoring error
```

### 3. Use Predefined Types for Known Fields

If you know certain fields should have specific types:

```go
generator := jsonschema.New(
    jsonschema.WithPredefined("id", jsonschema.Integer),
    jsonschema.WithPredefined("created_at", jsonschema.DateTime),
)
```

### 4. Validate Input Data

Ensure JSON is valid before adding:

```go
func addSampleSafely(generator *jsonschema.Generator, data string) error {
    // Pre-validate
    var temp interface{}
    if err := json.Unmarshal([]byte(data), &temp); err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }

    return generator.AddSample(data)
}
```

### 5. Save Schemas Periodically

For long-running processes:

```go
generator := jsonschema.New()

for i, sample := range samples {
    generator.AddSample(sample)

    // Save every 100 samples
    if (i+1) % 100 == 0 {
        schema, _ := generator.Generate()
        os.WriteFile("schema-checkpoint.json", []byte(schema), 0644)
    }
}
```

## Troubleshooting

### Issue: "no samples added" Error

**Problem:** Calling `Generate()` before adding any samples.

**Solution:**
```go
generator := jsonschema.New()
// Must add at least one sample
if err := generator.AddSample(`{"field": "value"}`); err != nil {
    log.Fatal(err)
}
schema, err := generator.Generate()
```

### Issue: Invalid JSON Error

**Problem:** Malformed JSON input.

**Solution:**
```go
jsonData := `{"invalid json}`
err := generator.AddSample(jsonData)
if err != nil {
    // Handle: failed to parse JSON: ...
    log.Printf("Invalid JSON: %v", err)
}
```

### Issue: All Fields Marked as Optional

**Problem:** Each sample has different fields.

**Solution:** Ensure samples have consistent structure, or this is expected behavior:
```go
// Sample 1: {a, b}
generator.AddSample(`{"a": 1, "b": 2}`)
// Sample 2: {a, c}
generator.AddSample(`{"a": 1, "c": 3}`)
// Sample 3: {b, c}
generator.AddSample(`{"b": 2, "c": 3}`)

// Result: no required fields (none appear in all 3 samples)
```

### Issue: Pattern Not Detected

**Problem:** String doesn't match expected pattern format.

**Solution:** Ensure strings follow the correct format:
```go
// ✓ Detected: ISO 8601 datetime
generator.AddSample(`{"time": "2023-01-15T10:30:00Z"}`)

// ✗ Not detected: custom datetime format
generator.AddSample(`{"time": "01/15/2023 10:30 AM"}`)

// ✓ Detected: valid email
generator.AddSample(`{"contact": "user@example.com"}`)

// ✗ Not detected: missing @ symbol
generator.AddSample(`{"contact": "user.example.com"}`)

// ✓ Detected: valid UUID
generator.AddSample(`{"id": "550e8400-e29b-41d4-a716-446655440000"}`)

// ✗ Not detected: invalid UUID format
generator.AddSample(`{"id": "550e8400-e29b-41d4"}`)
```

**Note:** All values in a field must match the pattern for format detection to apply. If even one value doesn't match, no format will be set.

### Issue: Unexpected Multiple Types

**Problem:** Field has inconsistent types across samples.

**Solution:** Use predefined types or ensure consistent data:
```go
// Problem: inconsistent types
generator.AddSample(`{"id": "123"}`)    // string
generator.AddSample(`{"id": 456}`)      // integer

// Solution 1: Use predefined type
generator := jsonschema.New(
    jsonschema.WithPredefined("id", jsonschema.Integer),
)

// Solution 2: Ensure consistent data types
generator.AddSample(`{"id": 123}`)      // integer
generator.AddSample(`{"id": 456}`)      // integer
```

### Issue: Build Errors

**Problem:** Build or test failures.

**Solution:**
```bash
# Ensure you have Go 1.25 or higher
go version

# Build and test
go build
go test -v
```

## Further Reading

- [Architecture Documentation](ARCHITECTURE.md) - Internal design details
- [API Documentation](https://pkg.go.dev/github.com/JLugagne/jsonschema-infer) - Full API reference
- [README](README.md) - Quick start and overview
