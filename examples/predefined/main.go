package main

import (
	"fmt"
	"log"

	"github.com/JLugagne/jsonschema-infer"
)

func main() {
	fmt.Println("=== Predefined Types Example ===")

	// Create generator with predefined types
	generator := jsonschema.New(
		jsonschema.WithPredefined("created_at", jsonschema.DateTime),
		jsonschema.WithPredefined("updated_at", jsonschema.DateTime),
		jsonschema.WithPredefined("user_id", jsonschema.Integer),
	)

	fmt.Println("Configured predefined types:")
	fmt.Println("  - created_at: DateTime")
	fmt.Println("  - updated_at: DateTime")
	fmt.Println("  - user_id: Integer")

	// Add samples
	samples := []string{
		`{"user_id": 123, "created_at": "2023-01-15T10:30:00Z", "updated_at": "2023-01-15T10:30:00Z"}`,
		`{"user_id": 456, "created_at": "2023-02-20T14:45:00Z", "updated_at": "2023-02-20T14:45:00Z"}`,
	}

	fmt.Println("\nAdding samples:")
	for i, sample := range samples {
		fmt.Printf("  Sample %d: %s\n", i+1, sample)
		err := generator.AddSample(sample)
		if err != nil {
			log.Fatal(err)
		}
	}

	schema, err := generator.Generate()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nGenerated Schema:")
	fmt.Println(schema)

	fmt.Println("\nNote: The predefined types are enforced, ensuring consistent schema")
	fmt.Println("      regardless of actual data variations.")
}
