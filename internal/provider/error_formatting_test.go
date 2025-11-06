package provider

import (
	"fmt"
	"strings"
	"testing"
)

func TestErrorMessageTemplating(t *testing.T) {
	mockError := fmt.Errorf("required property 'name' missing")
	
	testCases := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "default template (empty string)",
			template: "",
			expected: "JSON Schema validation failed: required property 'name' missing",
		},
		{
			name:     "simple custom template",
			template: "Error: {error} in {schema}",
			expected: "Error: required property 'name' missing in test.schema.json",
		},
		{
			name:     "go template syntax",
			template: "Validation failed: {{.Error}}",
			expected: "Validation failed: required property 'name' missing",
		},
		{
			name:     "detailed template with path",
			template: "Schema '{schema}' validation failed at '{path}': {error}",
			expected: "Schema 'test.schema.json' validation failed at '': required property 'name' missing",
		},
		{
			name:     "ci format template",
			template: "::error file={schema},line=1::{error}",
			expected: "::error file=test.schema.json,line=1::required property 'name' missing",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatValidationError(mockError, "test.schema.json", "{}", tc.template)
			if result.Error() != tc.expected {
				t.Errorf("Expected: %s\nGot: %s", tc.expected, result.Error())
			}
		})
	}
}

func TestProviderConfigDefaults(t *testing.T) {
	// Test that NewProviderConfig sets sensible defaults
	config, err := NewProviderConfig("", "", "")
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	expectedErrorTemplate := "JSON Schema validation failed: {error}"
	if config.DefaultErrorTemplate != expectedErrorTemplate {
		t.Errorf("Expected default error template: %s\nGot: %s", expectedErrorTemplate, config.DefaultErrorTemplate)
	}

	// Test that custom template is preserved
	customTemplate := "Custom error: {error}"
	config2, err := NewProviderConfig("", "", customTemplate)
	if err != nil {
		t.Fatalf("Failed to create provider config with custom template: %v", err)
	}

	if config2.DefaultErrorTemplate != customTemplate {
		t.Errorf("Expected custom error template: %s\nGot: %s", customTemplate, config2.DefaultErrorTemplate)
	}
}

func TestErrorFormattingEdgeCases(t *testing.T) {
	mockError := fmt.Errorf("validation error")

	// Test template with invalid Go template syntax
	result := FormatValidationError(mockError, "schema.json", "{}", "{{.InvalidField")
	if !strings.Contains(result.Error(), "template error") {
		t.Error("Expected template error fallback for invalid Go template")
	}

	// Test document truncation
	longDoc := strings.Repeat("x", 250)
	result2 := FormatValidationError(mockError, "schema.json", longDoc, "Doc: {document}")
	if !strings.Contains(result2.Error(), "...") {
		t.Error("Expected document to be truncated")
	}

	// Test all available variables
	result3 := FormatValidationError(mockError, "test.schema.json", `{"test": true}`, 
		"Error: {error}, Schema: {schema}, Path: {path}, Doc: {document}")
	expected := `Error: validation error, Schema: test.schema.json, Path: , Doc: {"test": true}`
	if result3.Error() != expected {
		t.Errorf("Expected: %s\nGot: %s", expected, result3.Error())
	}
}