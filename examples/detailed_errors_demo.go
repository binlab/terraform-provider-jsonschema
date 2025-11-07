package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/iilei/terraform-provider-jsonschema/internal/provider"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func main() {
	fmt.Println("JSON Schema Validation with Detailed Error Options")
	fmt.Println("=================================================")

	// Test schema
	schemaJSON := `{
		"type": "object",
		"required": ["name", "age", "email"],
		"properties": {
			"name": {"type": "string", "minLength": 3},
			"age": {"type": "integer", "minimum": 0, "maximum": 120},
			"email": {"type": "string", "format": "email"}
		}
	}`

	// Invalid test document
	documentJSON := `{
		"name": "Jo",
		"age": "twenty"
	}`

	// Compile schema using v6 API
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	
	var schemaData interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &schemaData); err != nil {
		log.Fatalf("Schema parsing error: %v", err)
	}
	
	schemaURL := "test://schema.json"
	if err := compiler.AddResource(schemaURL, schemaData); err != nil {
		log.Fatalf("Schema resource error: %v", err)
	}
	
	schema, err := compiler.Compile(schemaURL)
	if err != nil {
		log.Fatalf("Schema compilation error: %v", err)
	}

	// Parse document
	var document interface{}
	if err := json.Unmarshal([]byte(documentJSON), &document); err != nil {
		log.Fatalf("Document parsing error: %v", err)
	}

	// Validate and get error
	validationErr := schema.Validate(document)
	if validationErr == nil {
		fmt.Println("Document is valid!")
		return
	}

	fmt.Println("\n1. Basic Error Format (detailed_errors = false):")
	fmt.Println("------------------------------------------------")
	basicErr := provider.FormatValidationError(
		validationErr,
		"test://schema.json",
		documentJSON,
		"Validation failed: {error}",
		false, // detailed_errors = false
	)
	fmt.Println(basicErr.Error())

	fmt.Println("\n2. Detailed Error Format (detailed_errors = true):")
	fmt.Println("--------------------------------------------------")
	detailedErr := provider.FormatValidationError(
		validationErr,
		"test://schema.json",
		documentJSON,
		"Validation failed: {error}",
		true, // detailed_errors = true
	)
	fmt.Println(detailedErr.Error())

	fmt.Println("\n3. Using Structured Output Templates:")
	fmt.Println("------------------------------------")
	
	// Basic output template
	basicOutputErr := provider.FormatValidationError(
		validationErr,
		"test://schema.json",
		documentJSON,
		"Validation failed with basic output:\n{basic_output}",
		true, // detailed_errors = true to enable structured output
	)
	fmt.Printf("Basic Output:\n%s\n", basicOutputErr.Error())

	// Detailed output template
	detailedOutputErr := provider.FormatValidationError(
		validationErr,
		"test://schema.json",
		documentJSON,
		"Validation failed with detailed output:\n{detailed_output}",
		true, // detailed_errors = true to enable structured output
	)
	fmt.Printf("Detailed Output:\n%s\n", detailedOutputErr.Error())

	fmt.Println("\n4. Available Template Variables:")
	fmt.Println("-------------------------------")
	allVarsErr := provider.FormatValidationError(
		validationErr,
		"test://schema.json",
		documentJSON,
		"Schema: {schema}\nPath: {path}\nError: {error}\nDocument: {document}",
		true,
	)
	fmt.Printf("%s\n", allVarsErr.Error())

	fmt.Println("\n5. Common Predefined Templates:")
	fmt.Println("------------------------------")
	
	templates := []string{"simple", "detailed", "compact", "json", "verbose", "structured_basic", "structured_full"}
	for _, templateName := range templates {
		if template, exists := provider.GetCommonTemplate(templateName); exists {
			err := provider.FormatValidationError(
				validationErr,
				"test://schema.json",
				documentJSON,
				template,
				true, // Enable detailed errors for structured templates
			)
			fmt.Printf("%s: %s\n", templateName, err.Error())
		}
	}
}