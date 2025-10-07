package jsonschema

// Option is a functional option for configuring the Generator
type Option func(*Generator)

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
