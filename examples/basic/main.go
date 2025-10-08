package main

import (
	"fmt"
	"log"

	"github.com/JLugagne/jsonschema-infer"
)

func main() {
	fmt.Println("=== Basic Type Inference Example ===")

	// Create a new generator
	generator := jsonschema.New()

	// Add JSON samples
	samples := []string{
		`{"name": "John", "age": 30, "active": true}`,
		`{"name": "Jane", "age": 25, "active": false}`,
		`{"name": "Bob", "age": 35}`,
	}

	fmt.Println("Adding samples:")
	for i, sample := range samples {
		fmt.Printf("  Sample %d: %s\n", i+1, sample)
		err := generator.AddSample(sample)
		if err != nil {
			log.Fatalf("Failed to add sample: %v", err)
		}
	}

	// Generate the schema
	schema, err := generator.Generate()
	if err != nil {
		log.Fatalf("Failed to generate schema: %v", err)
	}

	fmt.Println("\nGenerated Schema:")
	fmt.Println(schema)

	fmt.Println("\nNote: 'active' is not in 'required' because it doesn't appear in all samples.")
}
