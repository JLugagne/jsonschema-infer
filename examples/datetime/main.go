package main

import (
	"fmt"
	"log"

	"github.com/JLugagne/jsonschema-infer"
)

func main() {
	fmt.Println("=== DateTime Detection Example ===")

	generator := jsonschema.New()

	// Add samples with ISO 8601 datetime strings
	samples := []string{
		`{"event": "login", "created_at": "2023-01-15T10:30:00Z"}`,
		`{"event": "logout", "created_at": "2023-01-15T11:45:00Z"}`,
		`{"event": "update", "created_at": "2023-01-15T14:20:00Z"}`,
	}

	fmt.Println("Adding samples with ISO 8601 timestamps:")
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

	fmt.Println("\nNote: The 'created_at' field has type 'string' with format 'date-time'")
	fmt.Println("      because all values match ISO 8601 datetime pattern.")
}
