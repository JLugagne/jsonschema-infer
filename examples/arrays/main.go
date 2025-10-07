package main

import (
	"fmt"
	"log"

	"github.com/JLugagne/jsonschema-infer"
)

func main() {
	fmt.Println("=== Arrays of Objects Example ===")

	generator := jsonschema.New()

	// Sample 1: Two products, both with price
	sample1 := `{
		"products": [
			{"id": 1, "name": "Product A", "price": 10.50},
			{"id": 2, "name": "Product B", "price": 15.99}
		]
	}`

	// Sample 2: One product without price
	sample2 := `{
		"products": [
			{"id": 3, "name": "Product C"}
		]
	}`

	fmt.Println("Sample 1 (all products have price):")
	fmt.Println(sample1)

	err := generator.AddSample(sample1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nSample 2 (product without price):")
	fmt.Println(sample2)

	err = generator.AddSample(sample2)
	if err != nil {
		log.Fatal(err)
	}

	// Generate schema
	schema, err := generator.Generate()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nGenerated Schema:")
	fmt.Println(schema)

	fmt.Println("\nNote: In array items, 'id' and 'name' are required (appear in all items),")
	fmt.Println("      but 'price' is optional (missing from Product C).")
}
