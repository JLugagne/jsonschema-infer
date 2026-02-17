# Architecture Documentation

This document describes the internal architecture and design decisions of the jsonschema-infer library.

## Table of Contents

- [Overview](#overview)
- [Design Philosophy](#design-philosophy)
- [Core Components](#core-components)
- [Data Flow](#data-flow)
- [Algorithm Details](#algorithm-details)
- [Design Decisions](#design-decisions)
- [Performance Characteristics](#performance-characteristics)
- [Extension Points](#extension-points)

## Overview

The jsonschema-infer library uses a **tree-based recursive architecture** to build JSON schemas incrementally from observed samples. The core principle is that complex structures (arrays, objects) are decomposed into simpler primitives, with each node in the tree handling only primitive type inference and delegating complex types to child nodes.

### Key Architectural Principles

1. **Incremental Updates**: Schema evolves after each sample
2. **Tree-Based Recursion**: Hierarchical node structure mirrors JSON structure
3. **Primitive Focus**: Each node handles only simple types, delegates complexity
4. **Sample Counting**: Track observation frequency for required field detection
5. **Separation of Concerns**: Clear boundaries between parsing, observation, and schema generation

## Design Philosophy

### Simplicity Over Optimization

The library prioritizes code clarity and maintainability over performance optimization. Each `AddSample()` call triggers a full schema rebuild, which is simple to understand and debug but may not be optimal for very large sample sets.

### Recursive Decomposition

Rather than handling complex nested structures with iterative algorithms, the library uses recursion to naturally mirror the hierarchical nature of JSON data.

### Explicit Over Implicit

Type inference rules are explicit and conservative. The library avoids "magic" heuristics in favor of straightforward observation-based inference.

## Core Components

### 1. Generator (`jsonschema.go`)

The `Generator` is the main entry point and orchestrates the schema inference process.

```go
type Generator struct {
    rootNode      *SchemaNode          // Root of the observation tree
    predefined    map[string]PredefinedType  // Field-specific type overrides
    customFormats []CustomFormat        // Format detectors (built-in + custom)
    sampleCount   int                   // Total samples observed
    maxSamples    int                   // Maximum samples to process (0 = unlimited)
    currentSchema *Schema               // Cached current schema
    schemaVersion SchemaVersion         // JSON Schema draft version (Draft06 or Draft07)
}
```

**Responsibilities:**
- Accept JSON samples and parse them
- Maintain global sample count
- Apply predefined type configurations
- Trigger schema rebuilds after each sample
- Serialize schemas to JSON

**Key Methods:**
- `New()` - Create generator with optional configuration
- `AddSample()` - Parse and observe a JSON sample
- `Generate()` - Serialize current schema to JSON
- `GetCurrentSchema()` - Return schema as object
- `Load()` - Reconstruct tree from existing schema

### 2. SchemaNode (`node.go`)

The `SchemaNode` is the core data structure representing a point in the JSON tree.

```go
type SchemaNode struct {
    observedTypes    map[string]int      // Type name -> count
    sampleCount      int                  // Observations of this node
    stringValues     []string             // String samples for pattern detection
    arrayItemNode    *SchemaNode          // Single child for array items
    objectProperties map[string]*SchemaNode  // Children for object properties
    predefinedType   *PredefinedType      // Type override
}
```

**Responsibilities:**
- Observe and categorize primitive values
- Maintain child nodes for complex types
- Track observation frequency
- Generate schema representation
- Apply pattern detection (datetime)

**Key Methods:**
- `NewSchemaNode()` - Create empty node
- `ObserveValue()` - Process a JSON value
- `ToSchema()` - Convert observations to Schema object
- `getPrimaryType()` - Determine most common type
- `applyStringPatterns()` - Detect datetime patterns
- `applyPredefinedType()` - Override with configured type

### 3. Schema (`schema.go`)

The `Schema` struct represents a JSON Schema (draft-07) definition.

```go
type Schema struct {
    Schema     string              `json:"$schema,omitempty"`
    Type       interface{}         `json:"type,omitempty"`  // string or []string
    Properties map[string]*Schema  `json:"properties,omitempty"`
    Items      *Schema             `json:"items,omitempty"`
    Required   []string            `json:"required,omitempty"`
    Format     string              `json:"format,omitempty"`
}
```

**Responsibilities:**
- Provide JSON-serializable schema representation
- Support both single and multiple types
- Handle standard JSON Schema properties

### 4. Options (`options.go`)

Functional options pattern for configuration.

```go
type Option func(*Generator)
type PredefinedType int

const (
    DateTime PredefinedType = iota
    String
    Boolean
    Number
    Integer
    Array
    Object
)
```

**Responsibilities:**
- Configure predefined field types
- Extensible configuration mechanism

**Key Functions:**
- `WithPredefined()` - Set field-specific type

## Data Flow

### Adding a Sample

```
User Input (JSON string)
    ↓
Generator.AddSample()
    ↓
json.Unmarshal() → interface{}
    ↓
rootNode.ObserveValue(interface{})
    ↓
[Recursive observation of value tree]
    ↓
Increment sampleCount
    ↓
applyPredefinedTypes()
    ↓
buildCurrentSchema()
    ↓
currentSchema cached
```

### Observing a Value (Recursive)

```
SchemaNode.ObserveValue(value)
    ↓
Increment sampleCount
    ↓
Determine primitive type
    ↓
observedTypes[type]++
    ↓
Switch on type:
    ├─ string: append to stringValues
    ├─ array:
    │   ├─ Create/reuse arrayItemNode
    │   └─ For each item: arrayItemNode.ObserveValue(item)
    └─ object:
        └─ For each property:
            ├─ Create/reuse objectProperties[key]
            └─ objectProperties[key].ObserveValue(value)
```

### Generating a Schema (Recursive)

```
Generator.Generate()
    ↓
Check sampleCount > 0
    ↓
Use/build currentSchema
    ↓
json.Marshal(currentSchema)
    ↓
Return JSON string

SchemaNode.ToSchema()
    ↓
Check predefinedType (override)
    ↓
Determine primary type
    ↓
Handle multiple types
    ↓
Switch on primary type:
    ├─ string: applyStringPatterns() (datetime)
    ├─ array:
    │   └─ schema.Items = arrayItemNode.ToSchema()
    └─ object:
        ├─ For each property:
        │   └─ schema.Properties[key] = childNode.ToSchema()
        └─ Determine required fields:
            └─ required if childNode.sampleCount == node.sampleCount
```

### Loading a Schema

```
Generator.Load(schemaJSON)
    ↓
json.Unmarshal() → Schema object
    ↓
Validate type == "object"
    ↓
Reset rootNode
    ↓
loadSchemaIntoNode(rootNode, schema, 1)
    ↓
[Recursive reconstruction of tree]
    ↓
Set sampleCount = 1
    ↓
Cache loaded schema
```

### Reconstructing Node from Schema (Recursive)

```
loadSchemaIntoNode(node, schema, parentSampleCount)
    ↓
Determine type (handle string or array)
    ↓
Initialize observedTypes[type] = parentSampleCount
    ↓
Set node.sampleCount = parentSampleCount
    ↓
Switch on type:
    ├─ array:
    │   └─ Create arrayItemNode
    │       └─ loadSchemaIntoNode(arrayItemNode, schema.Items, parentSampleCount)
    ├─ object:
    │   └─ For each property:
    │       ├─ Create objectProperties[key]
    │       ├─ Check if property is required
    │       ├─ childSampleCount = required ? parentSampleCount : parentSampleCount - 1
    │       └─ loadSchemaIntoNode(child, propSchema, childSampleCount)
    └─ string with date-time format:
        └─ Set stringValues = ["2023-01-01T00:00:00Z"]
```

## Algorithm Details

### Type Inference

**Primitive Type Detection:**

```go
func getPrimitiveType(value interface{}) string {
    switch v := value.(type) {
    case bool:
        return "boolean"
    case float64:
        if v == float64(int64(v)) {
            return "integer"
        }
        return "number"
    case string:
        return "string"
    case []interface{}:
        return "array"
    case map[string]interface{}:
        return "object"
    case nil:
        return "null"
    default:
        return "string"
    }
}
```

**Integer vs Number:**
- JSON numbers are parsed as `float64`
- Checked for integer equivalence: `v == float64(int64(v))`
- If equal, classified as integer

**Primary Type Selection:**
- When multiple types observed, select most frequent
- Used for determining schema structure (array items, object properties)
- All types included in schema output

### Required Field Detection

A field is **required** if and only if:
```
childNode.sampleCount == parentNode.sampleCount
```

**Example:**
```
Root (sampleCount=3)
  ├─ "name" (sampleCount=3)    → Required (3 == 3)
  ├─ "age" (sampleCount=2)     → Optional (2 != 3)
  └─ "email" (sampleCount=1)   → Optional (1 != 3)
```

### Array Item Merging

All array items across all samples merge into **single child node**:

```
Sample 1: {"arr": [1, 2, 3]}
  → arrayItemNode observes: 1, 2, 3

Sample 2: {"arr": [10, 20]}
  → arrayItemNode observes: 10, 20

Result: arrayItemNode.sampleCount = 5, observedTypes["integer"] = 5
```

For arrays of objects, each property has its own sample count:

```
Sample 1: {"users": [
  {"id": 1, "name": "A", "email": "a@x.com"},
  {"id": 2, "name": "B"}
]}
  → arrayItemNode.objectProperties["id"].sampleCount = 2
  → arrayItemNode.objectProperties["name"].sampleCount = 2
  → arrayItemNode.objectProperties["email"].sampleCount = 1

Sample 2: {"users": [
  {"id": 3, "name": "C", "email": "c@x.com"}
]}
  → arrayItemNode.objectProperties["id"].sampleCount = 3
  → arrayItemNode.objectProperties["name"].sampleCount = 3
  → arrayItemNode.objectProperties["email"].sampleCount = 2

Result: id and name required (3 == 3), email optional (2 != 3)
```

### Format Detection (Unified Mechanism)

The library uses a unified mechanism for all format detection (built-in and custom).

**Core Architecture:**
- All formats use `FormatDetector` functions: `func(string) bool`
- Built-in formats pre-registered at initialization
- Custom formats added via `WithCustomFormat()`
- Detection order: built-in first, then custom

**Algorithm:**
1. Collect all string values observed for a field
2. Iterate through format detectors in order
3. For each detector, check if ALL values match
4. First detector with 100% match wins
5. Set `format: <name>` in schema

**Example (DateTime):**
```
Sample 1: {"created_at": "2023-01-15T10:30:00Z"}
Sample 2: {"created_at": "2023-02-20T14:45:00Z"}
→ All match isDateTime() detector
→ created_at.format = "date-time"

Sample 1: {"created_at": "2023-01-15T10:30:00Z"}
Sample 2: {"created_at": "not a date"}
→ Not all match isDateTime() detector
→ created_at.format = undefined
```

**Built-in Detectors:**
- `isDateTime`: ISO 8601 (regex + time.Parse validation)
- `isEmail`: RFC 5322 simplified
- `isUUID`: UUIDs v1-v5
- `isIPv6`: IPv6 addresses
- `isIPv4`: IPv4 addresses
- `isURL`: HTTP/HTTPS/FTP/FTPS URLs

**Applied When:**
- Primary type is "string"
- No predefined type override
- All observed string values match a detector

### Schema Loading

**Sample Count Inference:**

When loading a schema, we don't know the original sample counts, so we infer:
- Root node: `sampleCount = 1`
- Required properties: `sampleCount = parentSampleCount`
- Optional properties: `sampleCount = parentSampleCount - 1`

This ensures:
- Required fields remain required after loading
- Optional fields remain optional after loading
- New samples can correctly update required status

**Limitation:** Loaded schemas lose exact observation counts, affecting precision of required field detection for subsequent samples.

## Design Decisions

### 1. Why Rebuild Schema After Each Sample?

**Decision:** Call `buildCurrentSchema()` after every `AddSample()`.

**Rationale:**
- Simplifies implementation (single code path)
- Makes schema always current (no stale state)
- Enables `GetCurrentSchema()` at any time
- Easier to debug and test

**Trade-off:** Performance cost for large sample sets.

**Alternative Considered:** Lazy evaluation (rebuild only on `Generate()`).
- **Rejected:** Complicates state management, less transparent behavior.

### 2. Why Tree Structure Instead of Flat Map?

**Decision:** Use recursive `SchemaNode` tree.

**Rationale:**
- Naturally mirrors JSON structure
- Isolates complexity (each node handles only primitives)
- Simplifies nested object/array handling
- Enables clean recursive algorithms

**Alternative Considered:** Flat map with path-based keys (e.g., `"user.address.city"`).
- **Rejected:** Complex path parsing, harder to handle arrays, less maintainable.

### 3. Why Single Child for Array Items?

**Decision:** Merge all array items into one `arrayItemNode`.

**Rationale:**
- JSON Schema typically defines uniform array item schemas
- Simplifies algorithm (no item position tracking)
- Matches most real-world use cases

**Trade-off:** Cannot detect tuple patterns (fixed-length arrays with different types per position).

**Alternative Considered:** Track each array position separately.
- **Rejected:** Rare use case, significant complexity increase.

### 4. Why encoding/json?

**Decision:** Use standard `encoding/json` package.

**Rationale:**
- Standard library API (familiar, well-documented)
- Simple `Unmarshal()` sufficient for semantic processing
- No need for low-level token manipulation
- No experimental flags required

**Initial Approach:** `encoding/json/jsontext` with manual token parsing.
- **Revised:** Overengineered for this use case, reverted to simple `Unmarshal()`.

**Requirement:** Go 1.25+.

### 5. Why Not Support Multiple Schema Versions?

**Decision:** Output only JSON Schema draft-07.

**Rationale:**
- Widely supported, stable specification
- Avoids complexity of version-specific logic
- Easy to add later if needed

**Future Extension:** Add `WithSchemaVersion()` option.

### 6. Why Functional Options Pattern?

**Decision:** Use `Option` type with `WithPredefined()` etc.

**Rationale:**
- Extensible (easy to add new options)
- Backward compatible (new options don't break existing code)
- Self-documenting (option names describe purpose)
- Idiomatic Go pattern

**Alternative Considered:** Config struct.
- **Rejected:** Requires versioning for breaking changes, less flexible.

## Performance Characteristics

### Time Complexity

**AddSample():**
- Parsing: O(n) where n = JSON size
- Observation: O(m) where m = number of fields
- Schema rebuild: O(m × d) where d = max depth
- **Total: O(n + m × d)**

**Generate():**
- Marshaling current schema: O(m)
- **Total: O(m)**

**Load():**
- Parsing: O(s) where s = schema size
- Reconstruction: O(m × d)
- **Total: O(s + m × d)**

### Space Complexity

**SchemaNode Tree:**
- Nodes: O(m) for m unique field paths
- String values: O(n × s) for n samples, s strings
- **Total: O(m + n × s)**

**Optimization Opportunity:** String value storage grows unbounded. Consider:
- Limit stored samples
- Store only for pattern detection
- Clear after pattern confirmed

### Scaling Characteristics

**Works Well:**
- Moderate sample counts (100s-1000s)
- Moderate field counts (10s-100s)
- Moderate nesting depth (≤ 5 levels)

**May Struggle:**
- Very large sample counts (100,000+) due to string storage
- Very high field counts (1000+) due to rebuild cost
- Extreme nesting depth (10+ levels) due to recursive overhead

**Future Optimizations:**
- Batch mode (rebuild only on demand)
- String value sampling (limit stored samples)
- Lazy child node creation

## Extension Points

### Adding Custom Format Detection

The unified format detection mechanism makes adding new formats simple - no need to modify internal code.

**Using WithCustomFormat:**

```go
// Define your detector function
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

// Register it
generator := jsonschema.New(
    jsonschema.WithCustomFormat("hex-color", isHexColor),
)

// Use it
generator.AddSample(`{"color": "#FF5733"}`)
// Result: color has format "hex-color"
```

**Adding Built-In Format (Internal):**

If you want to add a new built-in format to the library itself:

```go
// In node.go - add detector function
func isPhoneNumber(s string) bool {
    // E.164 format validation
    return len(s) > 10 && s[0] == '+' && /* ... */
}

// In jsonschema.go - add to getBuiltInFormats()
func getBuiltInFormats() []CustomFormat {
    return []CustomFormat{
        // ... existing formats ...
        {Name: "phone", Detector: isPhoneNumber},
    }
}
```

**Disabling Built-In Formats:**

For complete control over format detection:

```go
generator := jsonschema.New(
    jsonschema.WithoutBuiltInFormats(),
    jsonschema.WithCustomFormat("my-date", myDateDetector),
)
```

### Adding New Predefined Types

**Example:** Email type

```go
// In options.go
const (
    // ... existing types ...
    Email PredefinedType = iota
)

// In node.go
func (n *SchemaNode) applyPredefinedType() *Schema {
    schema := &Schema{}
    switch *n.predefinedType {
    // ... existing cases ...
    case Email:
        schema.Type = "string"
        schema.Format = "email"
    }
    return schema
}
```

### Adding Schema Validation Constraints

**Example:** Min/max for numbers

```go
// In schema.go
type Schema struct {
    // ... existing fields ...
    Minimum *float64 `json:"minimum,omitempty"`
    Maximum *float64 `json:"maximum,omitempty"`
}

// In node.go
type SchemaNode struct {
    // ... existing fields ...
    numberMin float64
    numberMax float64
}

func (n *SchemaNode) ObserveValue(value interface{}) {
    // ... existing logic ...
    if typeName == "number" || typeName == "integer" {
        if num, ok := value.(float64); ok {
            if n.sampleCount == 1 {
                n.numberMin = num
                n.numberMax = num
            } else {
                if num < n.numberMin {
                    n.numberMin = num
                }
                if num > n.numberMax {
                    n.numberMax = num
                }
            }
        }
    }
}

func (n *SchemaNode) ToSchema() *Schema {
    // ... existing logic ...
    if primaryType == "number" || primaryType == "integer" {
        schema.Minimum = &n.numberMin
        schema.Maximum = &n.numberMax
    }
}
```

### Schema Version Support

The library supports multiple JSON Schema draft versions. Users can choose between Draft 06 and Draft 07.

**Implementation:**

```go
// In options.go - SchemaVersion type and constants
type SchemaVersion string

const (
    Draft06 SchemaVersion = "http://json-schema.org/draft-06/schema#"
    Draft07 SchemaVersion = "http://json-schema.org/draft-07/schema#"
)

// Option to set schema version
func WithSchemaVersion(version SchemaVersion) Option {
    return func(g *Generator) {
        g.schemaVersion = version
    }
}

// In jsonschema.go - Generator defaults to Draft07
func New(opts ...Option) *Generator {
    g := &Generator{
        // ... other fields ...
        schemaVersion: Draft07, // Default to Draft 07
    }
    // Apply options...
    return g
}

// Schema version applied in buildCurrentSchema
func (g *Generator) buildCurrentSchema() *Schema {
    schema := g.rootNode.ToSchema(g.customFormats)
    if schema.Schema == "" {
        schema.Schema = string(g.schemaVersion)
    }
    return schema
}
```

**Usage:**

```go
// Draft 06 for schema generation
gen06 := jsonschema.New(jsonschema.WithSchemaVersion(jsonschema.Draft06))

// Draft 07 (default) for schema generation
gen07 := jsonschema.New() // or WithSchemaVersion(Draft07)

// Create empty schema with specific version (schema.go)
emptyDraft06 := jsonschema.NewSchemaWithVersion(jsonschema.Draft06)
emptyDraft07 := jsonschema.NewSchemaWithVersion(jsonschema.Draft07)

// Deprecated function (uses Draft07 by default)
oldSchema := jsonschema.NewSchema() // internally calls NewSchemaWithVersion(Draft07)
```

**Design Note:** For the features used by this library (types, arrays, objects, format, required), Draft 06 and Draft 07 are functionally equivalent. The main difference is the `$schema` URL.

**Schema Creation Functions:**
- `NewSchemaWithVersion(version SchemaVersion) *Schema` - Creates empty schema with specific version
- `NewSchema() *Schema` - **Deprecated** - Creates empty schema with Draft07 (calls NewSchemaWithVersion internally)
- For schema inference, always use `Generator` with `New()` and `AddSample()`

## Testing Strategy

### Unit Tests

Each component tested independently:
- `node.go`: ObserveValue, ToSchema, type detection, pattern matching
- `jsonschema.go`: AddSample, Generate, Load, sample counting
- `options.go`: Predefined types

### Integration Tests

End-to-end scenarios in `jsonschema_test.go`:
- Basic type inference
- Optional fields
- Arrays (simple and objects)
- DateTime detection
- Nested structures
- Load/resume functionality

### Test Coverage

**Uncovered Areas:**
- Error paths (invalid JSON, malformed schemas)
- Edge cases (empty arrays, null values)
- Performance tests (large sample sets)

## Future Enhancements

### Planned
1. String format detection (email, uri, uuid)
2. Basic constraints (min/max, length)
3. Batch mode for performance

### Under Consideration
1. Enum detection for limited value sets
2. Schema merging/combining
3. Alternative export formats (TypeScript, Go structs)
4. Streaming mode for very large datasets

### Not Planned
1. Tuple support (complex, rare use case)
2. Complex validation logic (out of scope)
3. Schema migration tools (separate concern)

---

**Document Version:** 1.0
**Last Updated:** 2025-10-07
**Maintainer:** jsonschema-infer team
