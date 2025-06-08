#!/bin/bash

echo "Generating Swagger documentation..."

if ! command -v swag &> /dev/null; then
    echo "swag CLI not found. Installing..."
    go install github.com/swaggo/swag/cmd/swag@latest
    echo "swag CLI installed!"
fi

echo "Generating documentation..."
swag init -g cmd/server/main.go -o ./docs

if [ -f "./docs/swagger.json" ]; then
    echo "Documentation generated with success!"
    echo ""
    echo "Documentation available in:"
    echo "      * Swagger UI: http://localhost:8080/swagger/index.html"
    echo "      * JSON: http://localhost:8080/swagger/doc.json"
    echo "      * YAML: http://localhost:8080/swagger/swagger.yaml"
    echo ""
else 
    echo "Failed to generate documentation"
    exit 1
fi