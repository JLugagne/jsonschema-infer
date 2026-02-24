package jsonschema

import (
	"net"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	// Email pattern (RFC 5322 simplified)
	emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// UUID pattern (supports v1-v5)
	uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
)

// SchemaNode represents a node in the schema tree
// Each node handles only primitives and delegates to children for complex types
type SchemaNode struct {
	// Type tracking for primitives
	observedTypes map[string]int // type name -> count
	sampleCount   int            // number of times this node was observed

	// For primitive string values - format detection
	// Candidates are eliminated incrementally in ObserveValue as each string arrives,
	// so no buffering of string values is required.  Memory cost is O(1) per field.
	stringCount        int                  // total number of string values ever observed
	candidateFormats   []string             // format names not yet eliminated; nil = not yet initialised
	candidateDetectors []func(string) bool  // detectors parallel to candidateFormats

	// Const tracking for primitive values (string, integer, number, boolean).
	// If all observed values are identical, constValue holds that value and
	// constDiffer is false, allowing "const" to be emitted in the schema.
	constValue  interface{}
	constSet    bool
	constDiffer bool

	// First value seen (used as example in schema)
	firstValue interface{}

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

// ObserveValue updates this node with a new observed value.
// formats is the list of format detectors to evaluate against string values;
// passing the same slice on every call is fine — it is read-only here.
func (n *SchemaNode) ObserveValue(value interface{}, examplesEnabled bool, formats []CustomFormat) {
	// Capture first value as example
	if examplesEnabled && n.sampleCount == 0 {
		n.firstValue = value
	}

	n.sampleCount++

	// Determine the primitive type
	typeName := getPrimitiveType(value)
	n.observedTypes[typeName]++

	// Track const candidate for primitive types. Null values and complex types
	// (object, array) are excluded — they cannot produce a useful const.
	switch typeName {
	case "string", "integer", "number", "boolean":
		if !n.constDiffer {
			if !n.constSet {
				n.constValue = value
				n.constSet = true
			} else if n.constValue != value {
				n.constDiffer = true
				n.constValue = nil
			}
		}
	}

	// Handle each type specifically
	switch typeName {
	case "string":
		if str, ok := value.(string); ok {
			n.stringCount++

			// Initialise candidate list on the very first string value.
			if n.candidateFormats == nil {
				n.candidateFormats = make([]string, 0, len(formats))
				n.candidateDetectors = make([]func(string) bool, 0, len(formats))
				for _, f := range formats {
					n.candidateFormats = append(n.candidateFormats, f.Name)
					n.candidateDetectors = append(n.candidateDetectors, f.Detector)
				}
			}

			// Eliminate candidates that don't match this string.
			// Compact in-place so we allocate nothing.
			if len(n.candidateFormats) > 0 {
				j := 0
				for i, detect := range n.candidateDetectors {
					if detect(str) {
						n.candidateFormats[j] = n.candidateFormats[i]
						n.candidateDetectors[j] = n.candidateDetectors[i]
						j++
					}
				}
				n.candidateFormats = n.candidateFormats[:j]
				n.candidateDetectors = n.candidateDetectors[:j]
			}
		}

	case "array":
		if arr, ok := value.([]interface{}); ok {
			// Ensure we have a node for array items
			if n.arrayItemNode == nil {
				n.arrayItemNode = NewSchemaNode()
			}
			// Observe each item in the array
			for _, item := range arr {
				n.arrayItemNode.ObserveValue(item, examplesEnabled, formats)
			}
		}

	case "object":
		if obj, ok := value.(map[string]interface{}); ok {
			// Observe each property. Null values are skipped: the node is still
			// created so the field appears in Properties, but its sampleCount is
			// not incremented, which makes the field optional (sampleCount < parent).
			for key, val := range obj {
				if n.objectProperties[key] == nil {
					n.objectProperties[key] = NewSchemaNode()
				}
				if val != nil {
					n.objectProperties[key].ObserveValue(val, examplesEnabled, formats)
				}
			}
		}
	}
}

// ToSchema converts this node to a JSON Schema.
// Format detection state is already fully up-to-date in candidateFormats — no
// formats argument is needed here.
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

	// Emit const when all observed primitive values were identical
	if n.constSet && !n.constDiffer {
		schema.Const = n.constValue
	}

	// Add example (first value observed)
	if n.firstValue != nil {
		schema.Example = n.firstValue
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

// applyStringPatterns sets the format on the schema based on the candidates that
// survived incremental elimination during ObserveValue calls.
// No processing happens here — all elimination is done eagerly as strings arrive.
func (n *SchemaNode) applyStringPatterns(schema *Schema) {
	if n.stringCount == 0 {
		return
	}
	if len(n.candidateFormats) > 0 {
		schema.Format = n.candidateFormats[0]
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

// isDateTime checks if a string value matches RFC 3339 (ISO 8601) datetime format.
// time.Parse is used directly – no regex pre-check needed.
func isDateTime(value string) bool {
	// Shortest valid RFC3339 value is "2006-01-02T15:04:05Z" (20 chars).
	if len(value) < 20 {
		return false
	}
	_, err := time.Parse(time.RFC3339, value)
	if err != nil {
		_, err = time.Parse(time.RFC3339Nano, value)
	}
	return err == nil
}

// isEmail checks if a string value matches email format.
// A cheap '@' presence check is done before the regex.
func isEmail(value string) bool {
	if !strings.ContainsRune(value, '@') {
		return false
	}
	return emailPattern.MatchString(value)
}

// isUUID checks if a string value matches UUID format.
// UUIDs are always exactly 36 characters; skip the regex for wrong lengths.
func isUUID(value string) bool {
	return len(value) == 36 && uuidPattern.MatchString(value)
}

// isIPv4 checks if a string value is a valid IPv4 address using net.ParseIP.
// This avoids the backtracking-prone regex entirely.
func isIPv4(value string) bool {
	ip := net.ParseIP(value)
	return ip != nil && ip.To4() != nil && strings.ContainsRune(value, '.')
}

// isIPv6 checks if a string value is a valid IPv6 address using net.ParseIP.
// This replaces the large alternation-heavy regex.
func isIPv6(value string) bool {
	ip := net.ParseIP(value)
	return ip != nil && strings.ContainsRune(value, ':')
}

// isURL checks if a string value is a valid HTTP/HTTPS/FTP/FTPS URL using net/url.
// net/url.Parse replaces the backtracking regex; a scheme allow-list is enforced.
func isURL(value string) bool {
	// Quick scheme pre-check to avoid parsing obviously non-URL strings.
	if !strings.HasPrefix(value, "http") && !strings.HasPrefix(value, "ftp") {
		return false
	}
	u, err := url.Parse(value)
	if err != nil || u.Host == "" {
		return false
	}
	switch u.Scheme {
	case "http", "https", "ftp", "ftps":
		return true
	}
	return false
}
