package main

import (
	"fmt"
	"log"

	"github.com/JLugagne/jsonschema-infer"
)

func main() {
	fmt.Println("=== Incremental Schema Updates Example ===")

	generator := jsonschema.New()

	samples := []string{
		`{"name": "John"}`,
		`{"name": "Jane", "age": 25}`,
		`{"name": "Bob", "age": 30, "email": "bob@example.com"}`,
	}

	for i, sample := range samples {
		fmt.Printf("After adding sample %d: %s\n", i+1, sample)

		err := generator.AddSample(sample)
		if err != nil {
			log.Fatal(err)
		}

		// Get current schema after each sample
		schema := generator.GetCurrentSchema()

		fmt.Printf("  Properties: %d\n", len(schema.Properties))
		fmt.Printf("  Required fields: %v\n", schema.Required)

		// Show property types
		fmt.Print("  Fields: ")
		for key := range schema.Properties {
			fmt.Printf("%s ", key)
		}
		fmt.Println()
		fmt.Println()
	}

	// Final schema
	finalSchema, err := generator.Generate()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Final Schema:")
	fmt.Println(finalSchema)

	fmt.Println("\nNote: The schema evolves incrementally:")
	fmt.Println("  - After sample 1: only 'name' (required)")
	fmt.Println("  - After sample 2: 'name' (required), 'age' (optional)")
	fmt.Println("  - After sample 3: 'name' (required), 'age' and 'email' (both optional)")
}
