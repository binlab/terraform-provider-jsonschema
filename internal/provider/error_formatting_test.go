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
	config, err := NewProviderConfig("", "")
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	expectedErrorTemplate := "JSON Schema validation failed: {error}"
	if config.DefaultErrorTemplate != expectedErrorTemplate {
		t.Errorf("Expected default error template: %s\nGot: %s", expectedErrorTemplate, config.DefaultErrorTemplate)
	}

	// Test that custom template is preserved
	customTemplate := "Custom error: {error}"
	config2, err := NewProviderConfig("", customTemplate)
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

// MockValidationError is a simple error for testing - not implementing ValidationError interface
// since the actual jsonschema.ValidationError is used in production
type MockValidationError struct {
	message string
	path    string
}

func (m *MockValidationError) Error() string {
	return m.message
}

func TestFormatValidationErrorComplexScenarios(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		schemaPath       string
		document         string
		template         string
		expectedContains []string
	}{
		{
			name:       "Go template with multiple variables",
			err:        &MockValidationError{message: "validation failed", path: "/name"},
			schemaPath: "/path/to/schema.json",
			document:   `{"name": 123}`,
			template:   "Error: {{.Error}} | Schema: {{.Schema}}",
			expectedContains: []string{"Error:", "validation failed", "Schema:", "/path/to/schema.json"},
		},
		{
			name:       "Simple template with all variables",
			err:        &MockValidationError{message: "type error", path: "/age"},
			schemaPath: "user.schema.json",
			document:   `{"age": "twenty"}`,
			template:   "Field failed: {error} in schema {schema} for document {document}",
			expectedContains: []string{"Field failed:", "type error", "schema user.schema.json", "document"},
		},
		{
			name:       "Template with missing variables",
			err:        &MockValidationError{message: "missing field"},
			schemaPath: "schema.json",
			document:   "{}",
			template:   "Error: {error} | Unknown: {unknown_var}",
			expectedContains: []string{"Error:", "missing field", "{unknown_var}"},
		},
		{
			name:       "Invalid Go template syntax",
			err:        &MockValidationError{message: "test error"},
			schemaPath: "test.json",
			document:   "{}",
			template:   "{{.Error", // Missing closing braces
			expectedContains: []string{"test error"}, // Should fall back to simple replacement
		},
		{
			name:       "Non-validation error",
			err:        fmt.Errorf("generic error"),
			schemaPath: "test.json",
			document:   "{}",
			template:   "Custom: {error}",
			expectedContains: []string{"Custom:", "generic error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatValidationError(tt.err, tt.schemaPath, tt.document, tt.template)

			if result == nil {
				t.Errorf("expected error but got nil")
				return
			}

			errorMessage := result.Error()
			for _, expected := range tt.expectedContains {
				if !strings.Contains(errorMessage, expected) {
					t.Errorf("expected error message to contain %q, got: %s", expected, errorMessage)
				}
			}
		})
	}
}

func TestNilErrorHandling(t *testing.T) {
	// Test that passing nil error doesn't crash - should return nil or handle gracefully
	result := FormatValidationError(nil, "test.json", "{}", "Error: {error}")
	// The function should handle nil gracefully, either by returning nil or a reasonable error
	if result == nil {
		// This is acceptable - function handles nil by returning nil
		return
	}
	// If not nil, it should be a reasonable error message
	if !strings.Contains(result.Error(), "error") {
		t.Errorf("expected error message to contain 'error', got: %v", result)
	}
}

func TestGetCommonTemplate(t *testing.T) {
	tests := []struct {
		name       string
		templateName string
		expectFound bool
		expectedTemplate string
	}{
		{
			name:         "simple template",
			templateName: "simple",
			expectFound:  true,
			expectedTemplate: "Validation failed: {error}",
		},
		{
			name:         "detailed template",
			templateName: "detailed",
			expectFound:  true,
			expectedTemplate: "JSON Schema validation failed:\n  Error: {error}\n  Schema: {schema}\n  Path: {path}",
		},
		{
			name:         "compact template",
			templateName: "compact",
			expectFound:  true,
			expectedTemplate: "[{schema}] {error} at {path}",
		},
		{
			name:         "ci template",
			templateName: "ci",
			expectFound:  true,
			expectedTemplate: "::error file={schema},line=1::{error}",
		},
		{
			name:         "json template",
			templateName: "json",
			expectFound:  true,
			expectedTemplate: `{"error": "{error}", "schema": "{schema}", "path": "{path}"}`,
		},
		{
			name:         "non-existent template",
			templateName: "nonexistent",
			expectFound:  false,
			expectedTemplate: "",
		},
		{
			name:         "empty template name",
			templateName: "",
			expectFound:  false,
			expectedTemplate: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, found := GetCommonTemplate(tt.templateName)
			
			if found != tt.expectFound {
				t.Errorf("GetCommonTemplate() found = %v, expectFound = %v", found, tt.expectFound)
			}
			
			if found && template != tt.expectedTemplate {
				t.Errorf("GetCommonTemplate() template = %v, expectedTemplate = %v", template, tt.expectedTemplate)
			}
		})
	}
}

func TestCommonErrorTemplatesExist(t *testing.T) {
	// Test that all expected common templates exist
	expectedTemplates := []string{"simple", "detailed", "compact", "ci", "json"}
	
	for _, templateName := range expectedTemplates {
		t.Run("template_"+templateName, func(t *testing.T) {
			template, found := GetCommonTemplate(templateName)
			if !found {
				t.Errorf("Expected common template '%s' not found", templateName)
			}
			if template == "" {
				t.Errorf("Common template '%s' is empty", templateName)
			}
		})
	}
}