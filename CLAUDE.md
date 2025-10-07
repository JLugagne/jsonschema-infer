# Claude Memory - Project Preferences

## User Requirements

### JSON Parsing
- **Use `encoding/json` Unmarshal** for JSON parsing (simpler and sufficient for this use case)
- Previously considered `encoding/json/jsontext` but decided it's overkill for semantic processing
- Standard `encoding/json` package works fine for this use case

### Architecture Design
- **Use tree/recursive structure** for schema building
- Each `SchemaNode` handles **only simple primitives** and delegates to child nodes for complex types
- Arrays: single child node that merges all array items
- Objects: map of property names to their schema nodes
- Keep code maintainable by scoping complexity to primitives within arrays or simple objects

### Schema Generation
- **Update schema definition incrementally** after each sample is added
- Call `buildCurrentSchema()` after every `AddSample()` invocation
- Schema should evolve as more samples are observed

### Documentation
- **Comprehensive documentation created** including:
  - `doc.go` - Package-level API documentation with examples
  - `USAGE.md` - Detailed usage guide with patterns and best practices
  - `ARCHITECTURE.md` - Internal design and algorithms
  - `examples/` - 7 runnable example programs with README
  - Updated `README.md` with links to all documentation

## Implementation Details

### File Structure
- `node.go` - Contains `SchemaNode` tree structure with recursive value observation
- `jsonschema.go` - Uses `encoding/json` Unmarshal for JSON parsing
- `schema.go` - JSON Schema output structures
- `options.go` - Functional options pattern for configuration
- `jsonschema_test.go` - Comprehensive tests (14 tests, all passing)
- `doc.go` - Package documentation
- `USAGE.md` - Usage guide
- `ARCHITECTURE.md` - Architecture documentation
- `examples/` - Runnable examples directory

### Key Design Patterns
1. **Tree-based observation**: Each node observes values and delegates to children
2. **No temporary generators**: Single tree structure that accumulates all observations
3. **Incremental updates**: Schema is always current after adding a sample
4. **Predefined types**: Support for field-specific type overrides
5. **Load/Resume**: Can load existing schemas and continue adding samples

### Array Handling
- Do NOT create temporary generators for arrays
- Use recursive `SchemaNode` structure where `arrayItemNode` accumulates all array items
- All items from all arrays across all samples merge into single child node
- For arrays of objects, track which fields are required across all array items

### Features Implemented
- ✅ Basic type inference (string, boolean, number, integer, array, object, null)
- ✅ Optional vs required field detection based on sample frequency
- ✅ Array handling with item schema merging
- ✅ Arrays of objects with optional field detection
- ✅ DateTime detection (ISO 8601 pattern matching)
- ✅ Predefined types configuration
- ✅ Incremental schema updates
- ✅ Load/Resume functionality
- ✅ Deeply nested structures support
- ✅ Multiple type detection (union types)
- ✅ Comprehensive test coverage

## Build Requirements
- Go 1.25 or higher recommended
- Standard build: `go build`
- Standard test: `go test -v`
- No special flags required
