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

### Documentation Requirements

**CRITICAL: ALWAYS UPDATE DOCUMENTATION WHEN CODE CHANGES**

When you add, modify, or remove features, you **MUST** update all relevant documentation files in the same commit/session:

#### Required Documentation Files
1. **`README.md`** - Update feature list, quick start examples if API changed
2. **`USAGE.md`** - Add usage examples for new features, update existing examples
3. **`ARCHITECTURE.md`** - Document internal design changes, new algorithms
4. **`FEATURES.md`** - Mark features as ✅ implemented or update planned features
5. **`CLAUDE.md`** - Update this file with new patterns, requirements, test counts
6. **`doc.go`** - Update package-level documentation and examples
7. **Test count** - Update test counts in all relevant files when tests are added

#### Documentation Checklist
When implementing a feature, complete ALL of these:
- [ ] Update `README.md` feature list
- [ ] Add examples to `USAGE.md`
- [ ] Document design in `ARCHITECTURE.md`
- [ ] Update `FEATURES.md` status
- [ ] Update `CLAUDE.md` feature list
- [ ] Add/update examples in `doc.go`
- [ ] Update test count everywhere (currently: 27 tests)

**DO NOT** consider a feature complete until all documentation is updated.

#### Documentation Files Present
- `doc.go` - Package-level API documentation with examples
- `USAGE.md` - Detailed usage guide with patterns and best practices
- `ARCHITECTURE.md` - Internal design and algorithms
- `FEATURES.md` - Feature tracking (implemented vs planned)
- `README.md` - Project overview and quick start
- `CLAUDE.md` - This file, project memory and requirements
- `examples/` - 7 runnable example programs with README

## Implementation Details

### File Structure
- `node.go` - Contains `SchemaNode` tree structure with recursive value observation
- `jsonschema.go` - Uses `encoding/json` Unmarshal for JSON parsing; includes `getBuiltInFormats()`
- `schema.go` - JSON Schema output structures
- `options.go` - Functional options pattern for configuration; includes `WithCustomFormat()`, `WithoutBuiltInFormats()`
- `jsonschema_test.go` - Comprehensive tests (27 tests, all passing)
- `doc.go` - Package documentation with examples
- `README.md` - Project overview and quick start
- `USAGE.md` - Detailed usage guide with patterns
- `ARCHITECTURE.md` - Internal design and algorithms
- `FEATURES.md` - Feature tracking (implemented vs planned)
- `CLAUDE.md` - This file, project memory and requirements
- `examples/` - 7 runnable example programs with README

### Key Design Patterns
1. **Tree-based observation**: Each node observes values and delegates to children
2. **No temporary generators**: Single tree structure that accumulates all observations
3. **Incremental updates**: Schema is always current after adding a sample
4. **Unified format detection**: All formats (built-in and custom) use the same `FormatDetector` mechanism
5. **Predefined types**: Support for field-specific type overrides
6. **Load/Resume**: Can load existing schemas and continue adding samples
7. **Flexible root types**: Supports objects, arrays, and primitives at root level

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
- ✅ Unified format detection mechanism (all formats use same pattern)
- ✅ Built-in formats: datetime, email, UUID, IPv4, IPv6, URL (HTTP/HTTPS/FTP/FTPS)
- ✅ Custom format detectors (user-defined pattern detection)
- ✅ Disable built-in formats option
- ✅ Predefined types configuration
- ✅ Flexible root types (object, array, primitives)
- ✅ Incremental schema updates
- ✅ Load/Resume functionality
- ✅ Deeply nested structures support
- ✅ Multiple type detection (union types)
- ✅ Max samples limit option
- ✅ Comprehensive test coverage (27 tests)

## Development Workflow

### When Adding/Modifying Features
1. **Implement the feature** in the appropriate `.go` file
2. **Write tests** in `jsonschema_test.go`
3. **Run tests** to ensure they pass: `go test -v`
4. **Update ALL documentation** (see Documentation Requirements above)
5. **Verify documentation** is consistent across all files
6. **Update test counts** in CLAUDE.md, FEATURES.md, README.md

### Common Mistakes to Avoid
- ❌ Implementing a feature without updating documentation
- ❌ Adding tests without updating test counts
- ❌ Updating only some documentation files
- ❌ Forgetting to update examples when API changes
- ❌ Not updating ARCHITECTURE.md when design changes

### Build Requirements
- Go 1.25 or higher recommended
- Standard build: `go build`
- Standard test: `go test -v`
- No special flags required

### Current Metrics
- **Test count**: 27 tests (update this when adding/removing tests)
- **Lines of code**: ~2034 lines across all .go files
- **Documentation files**: 7 files
