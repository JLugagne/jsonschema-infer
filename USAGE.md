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

### 5. DateTime Detection

ISO 8601 datetime strings are automatically detected:

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

### Issue: DateTime Not Detected

**Problem:** String doesn't match ISO 8601 format.

**Solution:** Ensure datetime strings follow ISO 8601:
```go
// ✓ Detected: ISO 8601 format
generator.AddSample(`{"time": "2023-01-15T10:30:00Z"}`)

// ✗ Not detected: custom format
generator.AddSample(`{"time": "01/15/2023 10:30 AM"}`)
```

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
