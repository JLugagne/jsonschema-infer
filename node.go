package jsonschema

import (
	"regexp"
	"sort"
	"time"
)

var (
	// ISO 8601 datetime pattern
	iso8601Pattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`)
)

// SchemaNode represents a node in the schema tree
// Each node handles only primitives and delegates to children for complex types
type SchemaNode struct {
	// Type tracking for primitives
	observedTypes map[string]int // type name -> count
	sampleCount   int            // number of times this node was observed

	// For primitive string values - pattern detection
	stringValues []string

	// For arrays - single child node that merges all array items
	arrayItemNode *SchemaNode

	// For objects - map of property names to their schema nodes
	objectProperties map[string]*SchemaNode

	// Predefined type override
	predefinedType *PredefinedType
}

// NewSchemaNode creates a new schema node
func NewSchemaNode() *SchemaNode {
	return &SchemaNode{
		observedTypes:    make(map[string]int),
		objectProperties: make(map[string]*SchemaNode),
	}
}

// ObserveValue updates this node with a new observed value
func (n *SchemaNode) ObserveValue(value interface{}) {
	n.sampleCount++

	// Determine the primitive type
	typeName := getPrimitiveType(value)
	n.observedTypes[typeName]++

	// Handle each type specifically
	switch typeName {
	case "string":
		if str, ok := value.(string); ok {
			n.stringValues = append(n.stringValues, str)
		}

	case "array":
		if arr, ok := value.([]interface{}); ok {
			// Ensure we have a node for array items
			if n.arrayItemNode == nil {
				n.arrayItemNode = NewSchemaNode()
			}
			// Observe each item in the array
			for _, item := range arr {
				n.arrayItemNode.ObserveValue(item)
			}
		}

	case "object":
		if obj, ok := value.(map[string]interface{}); ok {
			// Observe each property
			for key, val := range obj {
				if n.objectProperties[key] == nil {
					n.objectProperties[key] = NewSchemaNode()
				}
				n.objectProperties[key].ObserveValue(val)
			}
		}
	}
}

// ToSchema converts this node to a JSON Schema
func (n *SchemaNode) ToSchema() *Schema {
	schema := &Schema{}

	// Handle predefined types first
	if n.predefinedType != nil {
		return n.applyPredefinedType()
	}

	// Determine the primary type
	primaryType := n.getPrimaryType()

	// Handle multiple types
	if len(n.observedTypes) > 1 {
		types := make([]string, 0, len(n.observedTypes))
		for typ := range n.observedTypes {
			if typ != "null" {
				types = append(types, typ)
			}
		}
		sort.Strings(types) // Ensure consistent output
		if len(types) == 1 {
			schema.Type = types[0]
		} else if len(types) > 1 {
			schema.Type = types
		}
	} else {
		schema.Type = primaryType
	}

	// Apply type-specific logic
	switch primaryType {
	case "string":
		n.applyStringPatterns(schema)

	case "array":
		schema.Type = "array"
		if n.arrayItemNode != nil {
			schema.Items = n.arrayItemNode.ToSchema()
		}

	case "object":
		schema.Type = "object"
		if len(n.objectProperties) > 0 {
			schema.Properties = make(map[string]*Schema)
			required := []string{}

			for key, childNode := range n.objectProperties {
				schema.Properties[key] = childNode.ToSchema()
				// A property is required if it appeared in every observation of this object
				if childNode.sampleCount == n.sampleCount {
					required = append(required, key)
				}
			}

			if len(required) > 0 {
				sort.Strings(required) // Ensure consistent output
				schema.Required = required
			}
		}
	}

	return schema
}

// getPrimaryType returns the most commonly observed type
func (n *SchemaNode) getPrimaryType() string {
	var primaryType string
	maxCount := 0

	for typ, count := range n.observedTypes {
		if count > maxCount {
			maxCount = count
			primaryType = typ
		}
	}

	return primaryType
}

// applyStringPatterns detects and applies patterns for string types
func (n *SchemaNode) applyStringPatterns(schema *Schema) {
	if len(n.stringValues) == 0 {
		return
	}

	// Check if all strings match ISO 8601 datetime format
	allDateTime := true
	for _, str := range n.stringValues {
		if !isDateTime(str) {
			allDateTime = false
			break
		}
	}

	if allDateTime {
		schema.Format = "date-time"
	}
}

// applyPredefinedType applies a predefined type configuration
func (n *SchemaNode) applyPredefinedType() *Schema {
	schema := &Schema{}

	switch *n.predefinedType {
	case DateTime:
		schema.Type = "string"
		schema.Format = "date-time"
	case String:
		schema.Type = "string"
	case Boolean:
		schema.Type = "boolean"
	case Number:
		schema.Type = "number"
	case Integer:
		schema.Type = "integer"
	case Array:
		schema.Type = "array"
		if n.arrayItemNode != nil {
			schema.Items = n.arrayItemNode.ToSchema()
		}
	case Object:
		schema.Type = "object"
		if len(n.objectProperties) > 0 {
			schema.Properties = make(map[string]*Schema)
			for key, childNode := range n.objectProperties {
				schema.Properties[key] = childNode.ToSchema()
			}
		}
	}

	return schema
}

// getPrimitiveType determines the primitive type of a value
func getPrimitiveType(value interface{}) string {
	switch v := value.(type) {
	case bool:
		return "boolean"
	case float64:
		// Check if it's an integer
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

// isDateTime checks if a string value matches ISO 8601 datetime format
func isDateTime(value string) bool {
	if iso8601Pattern.MatchString(value) {
		// Additional validation: try to parse it
		_, err := time.Parse(time.RFC3339, value)
		return err == nil
	}
	return false
}
