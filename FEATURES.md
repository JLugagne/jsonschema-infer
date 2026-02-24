# Feature Status

This document tracks implemented and planned features for jsonschema-infer.

## ✅ Implemented Features

### Core Type Inference
- ✅ String type detection
- ✅ Integer type detection
- ✅ Number type detection (floating-point)
- ✅ Boolean type detection
- ✅ Array type detection
- ✅ Object type detection
- ✅ Null type detection
- ✅ Multiple/union types (e.g., `["string", "integer"]`)

### Field Requirements
- ✅ Required field detection (appears in all samples)
- ✅ Optional field detection (appears in some samples)
- ✅ Frequency-based tracking

### Complex Structures
- ✅ Nested objects (arbitrary depth)
- ✅ Arrays with merged item schemas
- ✅ Arrays of objects with optional field detection
- ✅ Deeply nested arrays of objects

### Format Detection (Unified Mechanism)
- ✅ All formats use same detection mechanism (`FormatDetector` functions)
- ✅ Built-in formats:
  - DateTime (ISO 8601) - `format: "date-time"`
  - Email addresses - `format: "email"`
  - UUID (v1-v5) - `format: "uuid"`
  - IPv4 addresses - `format: "ipv4"`
  - IPv6 addresses - `format: "ipv6"`
  - URLs (HTTP/HTTPS/FTP/FTPS) - `format: "uri"`

### Configuration
- ✅ Predefined types (DateTime, String, Boolean, Number, Integer, Array, Object)
- ✅ Max samples limit - `WithMaxSamples(int)`
- ✅ Custom format detectors - `WithCustomFormat(name, detector)`
- ✅ Disable built-in formats - `WithoutBuiltInFormats()`
- ✅ Schema version selection - `WithSchemaVersion(Draft06)` or `WithSchemaVersion(Draft07)`
- ✅ Enable/Disable examples - `WithExamples(bool)`

### Schema Management
- ✅ Lazy schema building (build on demand, not after every sample)
- ✅ `AddParsedSample(interface{})` - skip JSON parsing for pre-decoded values
- ✅ Load existing schema - `Load(schemaJSON)`
- ✅ Resume adding samples to loaded schema
- ✅ Get current schema as object - `GetCurrentSchema()`

### Output
- ✅ JSON Schema Draft 06 and Draft 07 support (configurable)
- ✅ Pretty-printed JSON output
- ✅ Support for array as root type
- ✅ Support for primitives as root type

### Concurrency
- ✅ Thread-safe operations with mutex

### Testing
- ✅ Comprehensive test coverage (43 tests)

---

## 📋 Planned Features

### Schema Constraints

#### Numeric Constraints
- ⬜ `minimum` - Track minimum observed value
- ⬜ `maximum` - Track maximum observed value
- ⬜ `exclusiveMinimum` - For range validation
- ⬜ `exclusiveMaximum` - For range validation
- ⬜ `multipleOf` - Detect common divisors

#### String Constraints
- ⬜ `minLength` - Track shortest string observed
- ⬜ `maxLength` - Track longest string observed
- ⬜ `pattern` - Custom regex patterns (user-defined)

#### Array Constraints
- ⬜ `minItems` - Track smallest array observed
- ⬜ `maxItems` - Track largest array observed
- ⬜ `uniqueItems` - Detect if all items are unique

#### Object Constraints
- ⬜ `minProperties` - Track minimum property count
- ⬜ `maxProperties` - Track maximum property count

#### Enum Detection
- ⬜ Automatic enum generation for fields with ≤N distinct values
- ⬜ Configurable threshold - `WithEnumThreshold(int)`

### Additional Format Detection

#### Date/Time Formats
- ⬜ `date` - Date without time (e.g., "2023-01-15")
- ⬜ `time` - Time without date (e.g., "10:30:00")
- ⬜ `duration` - ISO 8601 durations (e.g., "P3Y6M4DT12H30M5S")

#### Network Formats
- ⬜ `hostname` - Domain names (e.g., "example.com")
- ⬜ `idn-hostname` - Internationalized domain names
- ⬜ `uri-reference` - Relative URIs
- ⬜ `uri-template` - URI templates (RFC 6570)
- ⬜ `iri` - Internationalized Resource Identifiers
- ⬜ `iri-reference` - Relative IRIs

#### Data Formats
- ⬜ `regex` - Valid regular expressions
- ⬜ `json-pointer` - JSON Pointer (RFC 6901)
- ⬜ `relative-json-pointer` - Relative JSON Pointer
- ⬜ `byte` - Base64-encoded data

#### Other Formats
- ⬜ `phone` - Phone numbers (E.164 format)
- ⬜ `credit-card` - Credit card numbers
- ⬜ `hex-color` - Hexadecimal color codes (e.g., "#FF5733")
- ⬜ `currency` - Currency codes (ISO 4217)
- ⬜ `country-code` - Country codes (ISO 3166-1)

### Metadata & Documentation

- ⬜ `title` - Human-readable field names
- ⬜ `description` - Field descriptions
  - ⬜ User-provided via options
  - ⬜ Auto-generated from field names
- ✅ `examples` - Capture sample values from observations
- ⬜ `default` - Default values (most common value?)
- ⬜ `deprecated` - Mark obsolete fields
- ⬜ `readOnly` / `writeOnly` - API usage hints
- ⬜ `$comment` - Internal notes

### Advanced Type Features

#### Const & Literal Types
- ⬜ `const` - Field always has same value across all samples
- ⬜ Automatic const detection when value never varies

#### Schema Composition
- ⬜ `oneOf` - Field matches exactly one of several schemas
- ⬜ `anyOf` - Field matches one or more schemas
- ⬜ `allOf` - Field matches all schemas
- ⬜ `not` - Field must not match schema

#### Object Features
- ⬜ `additionalProperties` - Control for unexpected fields
  - ⬜ `false` - Strict mode (no extra fields)
  - ⬜ Schema - Extra fields must match schema
- ⬜ `patternProperties` - Schema for fields matching regex
- ⬜ `propertyNames` - Constraints on property names
- ⬜ `dependencies` - Field dependencies (if A then B required)
- ⬜ `dependentSchemas` - Schema changes based on field presence

#### Array Features
- ⬜ `tuple` validation - Arrays with positional schemas
  - ⬜ Don't merge all items; keep position-specific schemas
  - ⬜ `prefixItems` (draft 2020-12)
- ⬜ `contains` - Array must contain item matching schema
- ⬜ `minContains` / `maxContains` - Count constraints

#### Multiple Root Types
- ✅ Support array at root (not just object)
- ✅ Support primitives at root (string, number, etc.)
- ✅ Auto-detect root type from samples

### Validation & Analysis

#### Schema Validation
- ⬜ `Validate(jsonData string) error` - Check if JSON matches current schema
- ⬜ `ValidateWithDetails(jsonData string)` - Return detailed validation errors
- ⬜ `IsValid(jsonData string) bool` - Simple boolean check

#### Statistics & Analytics
- ⬜ `GetStats()` - Return statistics object:
  - ⬜ Total samples processed
  - ⬜ Field coverage (% of samples containing each field)
  - ⬜ Type distribution per field
  - ⬜ Value cardinality (distinct values per field)
  - ⬜ Min/max/avg values for numerics
  - ⬜ Min/max/avg lengths for strings/arrays
- ⬜ `GetFieldPaths()` - List all JSON paths in schema (e.g., `user.address.city`)
- ⬜ `GetFieldFrequency(path string)` - How often field appears

#### Schema Operations
- ⬜ `SchemaDiff(other *Schema)` - Compare two schemas
  - ⬜ Detect added/removed fields
  - ⬜ Detect type changes
  - ⬜ Detect constraint changes
- ⬜ `MergeSchema(other *Generator)` - Combine schemas from multiple generators
- ⬜ `Clone()` - Deep copy of generator state
- ⬜ `Reset()` - Clear all samples and start fresh

### Export Formats

#### Type Definitions
- ⬜ **TypeScript** - Generate `.d.ts` interface files
  - ⬜ `ExportTypeScript() string`
  - ⬜ Optional/required field handling
  - ⬜ Union types
  - ⬜ Nested interfaces
- ⬜ **Go structs** - Generate Go type definitions
  - ⬜ `ExportGoStruct(packageName string) string`
  - ⬜ JSON tags
  - ⬜ Pointer types for optional fields
  - ⬜ Custom type names
- ⬜ **Protobuf** - Generate `.proto` definitions
  - ⬜ `ExportProtobuf() string`
  - ⬜ Message definitions
  - ⬜ Field numbering
- ⬜ **GraphQL** - Generate GraphQL type definitions
  - ⬜ `ExportGraphQL() string`
  - ⬜ Type/Input type generation
  - ⬜ Required/optional fields

#### Schema Formats
- ⬜ **OpenAPI 3.x** - Convert to OpenAPI schema format
  - ⬜ `ExportOpenAPI() string`
  - ⬜ Component schema generation
- ⬜ **JSON Schema versions**
  - ⬜ Draft 2019-09 support
  - ⬜ Draft 2020-12 support
  - ⬜ Configurable output version

#### Other Formats
- ⬜ **Avro** - Apache Avro schema
- ⬜ **Thrift** - Apache Thrift IDL
- ⬜ **SQL DDL** - Database table definitions
- ⬜ **XML Schema** - XSD generation

### Sampling Control

#### Field Selection
- ⬜ `WithExcludeFields([]string)` - Ignore certain field paths
  - ⬜ Support wildcards (e.g., `*.internal`)
- ⬜ `WithIncludeFields([]string)` - Only process certain fields
- ⬜ `WithExcludePattern(regex)` - Exclude fields matching regex

#### Depth & Complexity
- ⬜ `WithMaxDepth(int)` - Limit nesting depth
- ⬜ `WithMaxArrayItems(int)` - Sample only first N array items
- ⬜ `WithMaxProperties(int)` - Sample only first N object properties

#### Requirements Control
- ⬜ `WithRequiredThreshold(float64)` - Field in X% of samples → required
  - ⬜ Default: 1.0 (100%)
  - ⬜ Example: 0.8 (80% of samples)
- ⬜ `WithMinSamples(int)` - Minimum samples before field appears in schema
  - ⬜ Helps filter noise from rare fields

#### Nullable Handling
- ⬜ `WithNullableMode(mode)` - Control null type handling
  - ⬜ `NullableAsUnion` - Use `["string", "null"]`
  - ⬜ `NullableAsOptional` - Omit from required array
  - ⬜ `NullableExplicit` - Add `nullable: true` (OpenAPI style)

### Performance Optimizations

#### Batch Processing
- ✅ Lazy generation — Schema only built when `Generate()` / `GetCurrentSchema()` is called
- ⬜ `AddSamples([]string)` - Convenience batch method

#### Streaming
- ⬜ `AddSampleStream(io.Reader)` - Process streaming JSON
- ⬜ `AddJSONLines(io.Reader)` - Process JSONL format
- ⬜ Large file handling without full memory load

#### Memory Management
- ⬜ `WithMaxStringValues(int)` - Limit stored string samples per field
  - ⬜ Currently stores all strings for pattern detection
  - ⬜ Could sample or use bloom filters
- ⬜ Sampling strategies for high-volume data
- ⬜ String deduplication/interning

### Advanced Features

#### Custom Type Detectors
- ✅ `WithCustomFormat(name, detector FormatDetector)` - User-defined patterns
- ⬜ `RegisterTypeInferrer(inferrer func(interface{}) string)` - Custom type logic
- ✅ Priority/ordering for custom detectors (checked after built-in formats)

#### Schema Evolution Tracking
- ⬜ Version tracking - Track how schema changes over time
- ⬜ `GetSchemaHistory()` - Return snapshots at different sample counts
- ⬜ Breaking change detection

#### Hooks & Callbacks
- ⬜ `OnFieldDiscovered(callback)` - Trigger when new field found
- ⬜ `OnTypeConflict(callback)` - Trigger when field has multiple types
- ⬜ `OnSampleAdded(callback)` - Trigger after each sample

#### Error Handling
- ⬜ Configurable error handling modes:
  - ⬜ Strict - Fail on any invalid JSON
  - ⬜ Lenient - Skip invalid samples with warning
  - ⬜ Coerce - Attempt to fix common JSON errors
- ⬜ `GetErrors()` - Return list of all errors encountered

### Output Options

#### Formatting
- ⬜ `WithIndent(string)` - Custom indentation
- ⬜ `WithCompact(bool)` - Minified vs pretty-print
- ⬜ `WithSortKeys(bool)` - Alphabetically sort properties

#### Schema Annotations
- ⬜ `With$id(string)` - Add schema $id
- ⬜ `WithSchemaVersion(string)` - Specify JSON Schema version
- ⬜ Custom `$defs` / `definitions` section

### CLI Tool
- ⬜ Command-line interface for jsonschema-infer
  - ⬜ Read from files or stdin
  - ⬜ Output to file or stdout
  - ⬜ Support all library options as flags
  - ⬜ Multiple input file support
  - ⬜ Watch mode (monitor directory for new samples)

### Integration & Ecosystem

#### Language Bindings
- ⬜ C bindings (cgo)
- ⬜ WASM compilation for browser/Node.js use

#### Plugins
- ⬜ Plugin system for custom exporters
- ⬜ Plugin system for custom format detectors

#### Testing Utilities
- ⬜ `GenerateTestData(schema)` - Create sample JSON matching schema
- ⬜ Fuzzing support for schema validation

---

## 🎯 Priority Recommendations

### High Priority (Most Requested)
1. **Enum detection** - Very common use case
2. **Numeric constraints** (min/max) - Essential for validation
3. **String length constraints** - Common validation need
4. **TypeScript export** - Popular for web development
5. **Batch processing** - Performance for large datasets

### Medium Priority (Nice to Have)
1. **Go struct export** - Useful for Go developers
2. **Required threshold configuration** - More flexible field detection
3. **Custom format detectors** - Extensibility
4. **Statistics API** - Understanding data better
5. **Const detection** - Useful for literal values

### Low Priority (Specialized)
1. **Protobuf/Avro export** - Niche use cases
2. **Schema versioning** - Complex feature
3. **WASM bindings** - Cross-platform scenarios
4. **CLI tool** - Convenience feature

---

## 🤝 Contributing

If you'd like to implement any of these features, please:
1. Open an issue to discuss the approach
2. Reference this document in your PR
3. Update this file to mark features as implemented (⬜ → ✅)
4. Add tests and documentation

---

## 📝 Notes

- Features marked ⬜ are planned but not yet implemented
- Features marked ✅ are implemented and tested
- This list is not exhaustive and may evolve based on user feedback
- Some features may be split into separate packages/modules

---

**Last Updated:** 2025-10-08
