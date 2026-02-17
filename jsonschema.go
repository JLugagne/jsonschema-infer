package jsonschema

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Generator generates JSON schemas from JSON samples
type Generator struct {
	mu            sync.RWMutex
	rootNode      *SchemaNode
	predefined    map[string]PredefinedType
	customFormats []CustomFormat
	sampleCount   int
	maxSamples    int
	currentSchema *Schema
	schemaVersion SchemaVersion
}

// New creates a new Generator with optional configuration
func New(opts ...Option) *Generator {
	g := &Generator{
		rootNode:      NewSchemaNode(),
		predefined:    make(map[string]PredefinedType),
		customFormats: getBuiltInFormats(),
		schemaVersion: Draft07, // Default to Draft 07
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

// getBuiltInFormats returns the default built-in format detectors
func getBuiltInFormats() []CustomFormat {
	return []CustomFormat{
		{Name: "date-time", Detector: isDateTime},
		{Name: "email", Detector: isEmail},
		{Name: "uuid", Detector: isUUID},
		{Name: "ipv6", Detector: isIPv6},
		{Name: "ipv4", Detector: isIPv4},
		{Name: "uri", Detector: isURL},
	}
}

// AddSample adds a JSON sample to the generator and updates the schema
// Thread-safe: can be called concurrently from multiple goroutines
func (g *Generator) AddSample(jsonData string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// If maxSamples is set and we've reached the limit, do nothing
	if g.maxSamples > 0 && g.sampleCount >= g.maxSamples {
		return nil
	}

	// Parse JSON using json/v2 Unmarshal
	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	g.sampleCount++

	// Observe the data with the root node
	g.rootNode.ObserveValue(data)

	// Apply predefined types to the tree
	g.applyPredefinedTypes()

	// Regenerate schema after each sample
	g.currentSchema = g.buildCurrentSchema()

	return nil
}

// applyPredefinedTypes applies predefined type configurations to nodes in the tree
func (g *Generator) applyPredefinedTypes() {
	for fieldName, predefinedType := range g.predefined {
		if node, exists := g.rootNode.objectProperties[fieldName]; exists {
			pt := predefinedType // Create a copy
			node.predefinedType = &pt
		}
	}
}

// buildCurrentSchema builds the current schema from the root node
func (g *Generator) buildCurrentSchema() *Schema {
	// Use the root node's ToSchema method which handles all types
	schema := g.rootNode.ToSchema(g.customFormats)

	// Add the $schema field
	if schema.Schema == "" {
		schema.Schema = string(g.schemaVersion)
	}

	return schema
}

// Generate generates a JSON schema from the accumulated samples
// Thread-safe: can be called concurrently from multiple goroutines
func (g *Generator) Generate() (string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.sampleCount == 0 {
		return "", fmt.Errorf("no samples added")
	}

	// Use the current schema (already built incrementally)
	schema := g.currentSchema
	if schema == nil {
		schema = g.buildCurrentSchema()
	}

	result, err := json.Marshal(schema)
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}

	return string(result), nil
}

// GetCurrentSchema returns the current schema as a Schema object
// This can be called after each AddSample to see the evolving schema
// Thread-safe: can be called concurrently from multiple goroutines
func (g *Generator) GetCurrentSchema() *Schema {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.currentSchema == nil {
		return g.buildCurrentSchema()
	}
	return g.currentSchema
}

// Load loads a previously generated JSON schema and initializes the generator
// This allows continuing to add samples to an existing schema
// Thread-safe: can be called concurrently from multiple goroutines
func (g *Generator) Load(schemaJSON string) error {
	var schema Schema
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	// Validate that it's an object schema
	if schema.Type != "object" {
		return fmt.Errorf("only object schemas can be loaded, got: %v", schema.Type)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Reset the generator
	g.rootNode = NewSchemaNode()
	g.currentSchema = nil

	// Reconstruct the tree structure from the schema
	// We set sampleCount to 1 to represent that this schema came from at least 1 sample
	if err := g.loadSchemaIntoNode(g.rootNode, &schema, 1); err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	// Set the generator's sample count based on the loaded schema
	// We use 1 as a baseline since we don't know the original count
	g.sampleCount = 1

	g.currentSchema = &schema

	return nil
}

// loadSchemaIntoNode recursively loads a schema into a node
func (g *Generator) loadSchemaIntoNode(node *SchemaNode, schema *Schema, parentSampleCount int) error {
	// Determine the type
	var typeStr string
	switch t := schema.Type.(type) {
	case string:
		typeStr = t
	case []interface{}:
		// Handle multiple types - use the first non-null type
		for _, typ := range t {
			if s, ok := typ.(string); ok && s != "null" {
				typeStr = s
				break
			}
		}
	default:
		return fmt.Errorf("unsupported type format: %T", t)
	}

	// Initialize the node based on type
	if node.observedTypes == nil {
		node.observedTypes = make(map[string]int)
	}
	node.observedTypes[typeStr] = parentSampleCount
	node.sampleCount = parentSampleCount

	// Handle arrays
	if typeStr == "array" && schema.Items != nil {
		node.arrayItemNode = NewSchemaNode()
		// Array items inherit the parent's sample count
		if err := g.loadSchemaIntoNode(node.arrayItemNode, schema.Items, parentSampleCount); err != nil {
			return err
		}
	}

	// Handle objects
	if typeStr == "object" && schema.Properties != nil {
		if node.objectProperties == nil {
			node.objectProperties = make(map[string]*SchemaNode)
		}
		for key, propSchema := range schema.Properties {
			childNode := NewSchemaNode()
			// Check if this property is required
			childSampleCount := parentSampleCount
			isRequired := false
			for _, req := range schema.Required {
				if req == key {
					isRequired = true
					break
				}
			}
			// If not required, it appeared in fewer samples
			if !isRequired {
				childSampleCount = parentSampleCount - 1
				if childSampleCount < 1 {
					childSampleCount = 1
				}
			}
			if err := g.loadSchemaIntoNode(childNode, propSchema, childSampleCount); err != nil {
				return err
			}
			node.objectProperties[key] = childNode
		}
	}

	// Handle string patterns
	if typeStr == "string" && schema.Format == "date-time" {
		node.stringValues = []string{"2023-01-01T00:00:00Z"} // Placeholder for datetime detection
	}

	return nil
}
