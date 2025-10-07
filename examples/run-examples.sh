#!/bin/bash
# Run all examples for jsonschema-infer

set -e

echo "========================================"
echo "Running jsonschema-infer Examples"
echo "========================================"
echo ""

EXAMPLES=(
    "basic"
    "arrays"
    "datetime"
    "predefined"
    "load_resume"
    "nested"
    "incremental"
)

for example in "${EXAMPLES[@]}"; do
    echo "----------------------------------------"
    echo "Running: $example"
    echo "----------------------------------------"
    (cd "$example" && go run main.go)
    echo ""
    echo ""
done

echo "========================================"
echo "All examples completed successfully!"
echo "========================================"
