package main

import (
	"fmt"
	"log"

	"github.com/JLugagne/jsonschema-infer"
)

func main() {
	fmt.Println("=== Load and Resume Example ===")

	// Step 1: Generate initial schema
	fmt.Println("Step 1: Create initial schema")
	generator1 := jsonschema.New()

	err := generator1.AddSample(`{"id": 1, "name": "John", "age": 30}`)
	if err != nil {
		log.Fatal(err)
	}

	err = generator1.AddSample(`{"id": 2, "name": "Jane", "age": 25}`)
	if err != nil {
		log.Fatal(err)
	}

	initialSchema, err := generator1.Generate()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Initial schema:")
	fmt.Println(initialSchema)

	// Step 2: Load the schema and add new samples
	fmt.Println("\nStep 2: Load schema and add new sample with additional field")
	generator2 := jsonschema.New()

	err = generator2.Load(initialSchema)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Schema loaded successfully!")

	// Add new sample with additional "email" field
	err = generator2.AddSample(`{"id": 3, "name": "Bob", "age": 40, "email": "bob@example.com"}`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Added sample: {\"id\": 3, \"name\": \"Bob\", \"age\": 40, \"email\": \"bob@example.com\"}")

	// Generate updated schema
	updatedSchema, err := generator2.Generate()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nUpdated schema:")
	fmt.Println(updatedSchema)

	fmt.Println("\nNote: The updated schema now includes 'email' as an optional field")
	fmt.Println("      (appears in 1 of 2 samples after loading).")
}
