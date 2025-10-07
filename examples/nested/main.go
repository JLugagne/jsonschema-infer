package main

import (
	"fmt"
	"log"

	"github.com/JLugagne/jsonschema-infer"
)

func main() {
	fmt.Println("=== Nested Objects and Arrays Example ===")

	generator := jsonschema.New()

	// Complex nested structure
	sample1 := `{
		"company": {
			"name": "Acme Corp",
			"founded": 1995,
			"departments": [
				{
					"name": "Engineering",
					"employees": [
						{"name": "Alice", "role": "Senior Engineer", "years": 5},
						{"name": "Bob", "role": "Engineer", "years": 2}
					]
				},
				{
					"name": "Design",
					"employees": [
						{"name": "Carol", "role": "Designer"}
					]
				}
			]
		}
	}`

	sample2 := `{
		"company": {
			"name": "TechCo",
			"founded": 2010,
			"departments": [
				{
					"name": "Sales",
					"employees": [
						{"name": "Dave", "role": "Sales Manager", "years": 3}
					]
				}
			]
		}
	}`

	fmt.Println("Sample 1 (Acme Corp):")
	fmt.Println(sample1)

	err := generator.AddSample(sample1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nSample 2 (TechCo):")
	fmt.Println(sample2)

	err = generator.AddSample(sample2)
	if err != nil {
		log.Fatal(err)
	}

	schema, err := generator.Generate()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nGenerated Schema:")
	fmt.Println(schema)

	fmt.Println("\nNote: The schema handles deep nesting:")
	fmt.Println("  - company.departments is an array of objects")
	fmt.Println("  - Each department has employees (array of objects)")
	fmt.Println("  - 'years' is optional in employee objects (not present for Carol)")
}
