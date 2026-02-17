package jsonschema

// Option is a functional option for configuring the Generator
type Option func(*Generator)

// SchemaVersion represents a JSON Schema draft version
type SchemaVersion string

const (
	// Draft06 represents JSON Schema Draft 06
	Draft06 SchemaVersion = "http://json-schema.org/draft-06/schema#"
	// Draft07 represents JSON Schema Draft 07 (default)
	Draft07 SchemaVersion = "http://json-schema.org/draft-07/schema#"
)

// FormatDetector is a function that checks if a string matches a custom format
type FormatDetector func(string) bool

// CustomFormat represents a user-defined format with its detector function
type CustomFormat struct {
	Name     string
	Detector FormatDetector
}

// PredefinedType represents a predefined type for a field
type PredefinedType string

const (
	DateTime PredefinedType = "datetime"
	String   PredefinedType = "string"
	Boolean  PredefinedType = "boolean"
	Number   PredefinedType = "number"
	Integer  PredefinedType = "integer"
	Array    PredefinedType = "array"
	Object   PredefinedType = "object"
)

// WithPredefined sets a predefined type for a field name
func WithPredefined(fieldName string, typeValue PredefinedType) Option {
	return func(g *Generator) {
		if g.predefined == nil {
			g.predefined = make(map[string]PredefinedType)
		}
		g.predefined[fieldName] = typeValue
	}
}

// WithMaxSamples sets the maximum number of samples to process
// Once this limit is reached, AddSample will return nil but do nothing
func WithMaxSamples(max int) Option {
	return func(g *Generator) {
		g.maxSamples = max
	}
}

// WithCustomFormat registers a custom format detector
// Custom formats are checked after built-in formats (date-time, email, uuid, ipv6, ipv4, uri)
// The formatName will be used as the value for the "format" field in the schema
func WithCustomFormat(formatName string, detector FormatDetector) Option {
	return func(g *Generator) {
		g.customFormats = append(g.customFormats, CustomFormat{
			Name:     formatName,
			Detector: detector,
		})
	}
}

// WithoutBuiltInFormats disables all built-in format detectors
// Use this if you want to provide your own complete set of format detectors
func WithoutBuiltInFormats() Option {
	return func(g *Generator) {
		g.customFormats = []CustomFormat{}
	}
}

// WithSchemaVersion sets the JSON Schema draft version
// Defaults to Draft07 if not specified
func WithSchemaVersion(version SchemaVersion) Option {
	return func(g *Generator) {
		g.schemaVersion = version
	}
}
