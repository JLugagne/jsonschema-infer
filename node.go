package jsonschema

import (
	"regexp"
	"sort"
	"time"
)

var (
	// ISO 8601 datetime pattern
	iso8601Pattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`)

	// Email pattern (RFC 5322 simplified)
	emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// UUID pattern (supports v1-v5)
	uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

	// IPv4 pattern
	ipv4Pattern = regexp.MustCompile(`^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}$`)

	// IPv6 pattern (simplified - handles most common cases)
	ipv6Pattern = regexp.MustCompile(`^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`)

	// URL pattern (HTTP/HTTPS/FTP/FTPS)
	urlPattern = regexp.MustCompile(`^(https?|ftps?)://[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*(/.*)?$`)
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
func (n *SchemaNode) ToSchema(customFormats ...[]CustomFormat) *Schema {
	schema := &Schema{}

	// Extract custom formats from variadic parameter
	var formats []CustomFormat
	if len(customFormats) > 0 {
		formats = customFormats[0]
	}

	// Handle predefined types first
	if n.predefinedType != nil {
		return n.applyPredefinedType(formats)
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
		n.applyStringPatterns(schema, formats)

	case "array":
		schema.Type = "array"
		if n.arrayItemNode != nil {
			schema.Items = n.arrayItemNode.ToSchema(formats)
		}

	case "object":
		schema.Type = "object"
		if len(n.objectProperties) > 0 {
			schema.Properties = make(map[string]*Schema)
			required := []string{}

			for key, childNode := range n.objectProperties {
				schema.Properties[key] = childNode.ToSchema(formats)
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
// Checks all formats (built-in and custom) in order
func (n *SchemaNode) applyStringPatterns(schema *Schema, formats []CustomFormat) {
	if len(n.stringValues) == 0 {
		return
	}

	// Check all formats in order (built-in formats come first, then user custom formats)
	for _, format := range formats {
		if allMatch(n.stringValues, format.Detector) {
			schema.Format = format.Name
			return
		}
	}
}

// applyPredefinedType applies a predefined type configuration
func (n *SchemaNode) applyPredefinedType(formats []CustomFormat) *Schema {
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
			schema.Items = n.arrayItemNode.ToSchema(formats)
		}
	case Object:
		schema.Type = "object"
		if len(n.objectProperties) > 0 {
			schema.Properties = make(map[string]*Schema)
			for key, childNode := range n.objectProperties {
				schema.Properties[key] = childNode.ToSchema(formats)
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

// allMatch checks if all strings match a given pattern function
func allMatch(values []string, matchFunc func(string) bool) bool {
	for _, str := range values {
		if !matchFunc(str) {
			return false
		}
	}
	return true
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

// isEmail checks if a string value matches email format
func isEmail(value string) bool {
	return emailPattern.MatchString(value)
}

// isUUID checks if a string value matches UUID format
func isUUID(value string) bool {
	return uuidPattern.MatchString(value)
}

// isIPv4 checks if a string value matches IPv4 format
func isIPv4(value string) bool {
	return ipv4Pattern.MatchString(value)
}

// isIPv6 checks if a string value matches IPv6 format
func isIPv6(value string) bool {
	return ipv6Pattern.MatchString(value)
}

// isURL checks if a string value matches URL format
func isURL(value string) bool {
	return urlPattern.MatchString(value)
}
