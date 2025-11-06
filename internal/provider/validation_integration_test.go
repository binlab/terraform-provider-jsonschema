package provider

import (
	"testing"
	
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// Test the actual validation schema and compiler behavior
func TestSchemaCompilationAndValidation(t *testing.T) {
	tests := []struct {
		name       string
		schema     string
		document   string
		version    string
		expectErr  bool
	}{
		{
			name: "draft-07 validation success",
			schema: `{
				"$schema": "http://json-schema.org/draft-07/schema#",
				"type": "object",
				"required": ["name"],
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "integer", "minimum": 0}
				}
			}`,
			document:  `{"name": "John", "age": 25}`,
			version:   "draft-07",
			expectErr: false,
		},
		{
			name: "draft-04 validation success",
			schema: `{
				"$schema": "http://json-schema.org/draft-04/schema#",
				"type": "object",
				"required": ["test"],
				"properties": {
					"test": {"type": "string"}
				}
			}`,
			document:  `{"test": "value"}`,
			version:   "draft-04",
			expectErr: false,
		},
		{
			name: "validation failure - missing required field",
			schema: `{
				"type": "object",
				"required": ["name"],
				"properties": {
					"name": {"type": "string"}
				}
			}`,
			document:  `{"age": 25}`,
			version:   "draft-07",
			expectErr: true,
		},
		{
			name: "validation failure - type mismatch",
			schema: `{
				"type": "object",
				"properties": {
					"age": {"type": "integer"}
				}
			}`,
			document:  `{"age": "twenty-five"}`,
			version:   "draft-07",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test schema compilation using the correct API
			schema, err := jsonschema.CompileString("test-schema", tt.schema)
			if err != nil {
				if !tt.expectErr {
					t.Fatalf("failed to compile schema: %v", err)
				}
				return // Expected compilation error
			}

			// Parse the document
			document, err := ParseJSON5String(tt.document)
			if err != nil {
				if !tt.expectErr {
					t.Fatalf("failed to parse document: %v", err)
				}
				return // Expected parsing error
			}

			// Validate the document
			err = schema.Validate(document)
			if tt.expectErr && err == nil {
				t.Errorf("expected validation error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

// Test configuration resolution between provider and resource levels
func TestConfigurationResolution(t *testing.T) {
	tests := []struct {
		name                      string
		providerSchemaVersion     string
		providerBaseURL           string
		providerErrorTemplate     string
		resourceSchemaVersion     string
		resourceBaseURL           string
		resourceErrorTemplate     string
		expectedFinalVersion      string
		expectedFinalBaseURL      string
		expectedFinalErrorTemplate string
	}{
		{
			name:                      "all provider defaults",
			providerSchemaVersion:     "draft-07",
			providerBaseURL:           "https://example.com/",
			providerErrorTemplate:     "Provider: {error}",
			expectedFinalVersion:      "draft-07",
			expectedFinalBaseURL:      "https://example.com/",
			expectedFinalErrorTemplate: "Provider: {error}",
		},
		{
			name:                      "resource overrides all",
			providerSchemaVersion:     "draft-07",
			providerBaseURL:           "https://example.com/",
			providerErrorTemplate:     "Provider: {error}",
			resourceSchemaVersion:     "draft-04",
			resourceBaseURL:           "https://resource.com/",
			resourceErrorTemplate:     "Resource: {error}",
			expectedFinalVersion:      "draft-04",
			expectedFinalBaseURL:      "https://resource.com/",
			expectedFinalErrorTemplate: "Resource: {error}",
		},
		{
			name:                      "partial resource override",
			providerSchemaVersion:     "draft-07",
			providerBaseURL:           "https://example.com/",
			providerErrorTemplate:     "Provider: {error}",
			resourceSchemaVersion:     "draft-04",
			// resource doesn't specify base URL or error template
			expectedFinalVersion:      "draft-04",
			expectedFinalBaseURL:      "https://example.com/",
			expectedFinalErrorTemplate: "Provider: {error}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create provider config
			providerConfig, err := NewProviderConfig(tt.providerSchemaVersion, tt.providerBaseURL, tt.providerErrorTemplate)
			if err != nil {
				t.Fatalf("failed to create provider config: %v", err)
			}

			// Simulate configuration resolution logic (like what happens in the data source)
			finalVersion := tt.resourceSchemaVersion
			if finalVersion == "" {
				finalVersion = providerConfig.DefaultSchemaVersion
			}

			finalBaseURL := tt.resourceBaseURL
			if finalBaseURL == "" {
				finalBaseURL = providerConfig.DefaultBaseURL
			}

			finalErrorTemplate := tt.resourceErrorTemplate
			if finalErrorTemplate == "" {
				finalErrorTemplate = providerConfig.DefaultErrorTemplate
			}

			// Verify the resolution matches expectations
			if finalVersion != tt.expectedFinalVersion {
				t.Errorf("expected final version %q, got %q", tt.expectedFinalVersion, finalVersion)
			}
			if finalBaseURL != tt.expectedFinalBaseURL {
				t.Errorf("expected final base URL %q, got %q", tt.expectedFinalBaseURL, finalBaseURL)
			}
			if finalErrorTemplate != tt.expectedFinalErrorTemplate {
				t.Errorf("expected final error template %q, got %q", tt.expectedFinalErrorTemplate, finalErrorTemplate)
			}
		})
	}
}

// Test edge cases for validation
func TestValidationEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		document string
		expectErr bool
	}{
		{
			name: "empty schema",
			schema: `{}`,
			document: `{"anything": "goes"}`,
			expectErr: false, // Empty schema should allow anything
		},
		{
			name: "schema with $ref (should work with proper base URL)",
			schema: `{
				"type": "object",
				"properties": {
					"test": {"$ref": "#/definitions/testDef"}
				},
				"definitions": {
					"testDef": {"type": "string"}
				}
			}`,
			document: `{"test": "value"}`,
			expectErr: false,
		},
		{
			name: "complex nested validation",
			schema: `{
				"type": "object",
				"required": ["users"],
				"properties": {
					"users": {
						"type": "array",
						"items": {
							"type": "object",
							"required": ["name", "email"],
							"properties": {
								"name": {"type": "string"},
								"email": {"type": "string", "format": "email"}
							}
						}
					}
				}
			}`,
			document: `{
				"users": [
					{"name": "John", "email": "john@example.com"},
					{"name": "Jane", "email": "jane@example.com"}
				]
			}`,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := jsonschema.CompileString("test-schema", tt.schema)
			if err != nil {
				if !tt.expectErr {
					t.Fatalf("failed to compile schema: %v", err)
				}
				return
			}

			document, err := ParseJSON5String(tt.document)
			if err != nil {
				if !tt.expectErr {
					t.Fatalf("failed to parse document: %v", err)
				}
				return
			}

			err = schema.Validate(document)
			if tt.expectErr && err == nil {
				t.Errorf("expected validation error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}