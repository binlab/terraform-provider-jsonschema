package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderNew(t *testing.T) {
	if New("dev") == nil {
		t.Error("Provider should not be nil")
	}
}

func TestProviderConfigure(t *testing.T) {
	tests := []struct {
		name          string
		schemaVersion string
		baseURL       string
		errorTemplate string
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid configuration",
			schemaVersion: "draft-07",
			baseURL:       "https://example.com/",
			errorTemplate: "Error: {error}",
			expectError:   false,
		},
		{
			name:          "invalid schema version",
			schemaVersion: "invalid-version",
			baseURL:       "",
			errorTemplate: "",
			expectError:   true,
			errorContains: "unsupported JSON Schema version",
		},
		{
			name:          "valid with base URL",
			schemaVersion: "draft-07",
			baseURL:       "./schemas/",
			errorTemplate: "",
			expectError:   false,
		},
		{
			name:          "defaults",
			schemaVersion: "",
			baseURL:       "",
			errorTemplate: "",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the config creation logic directly (since testing the actual
			// providerConfigure function would require complex mocking)
			config, err := NewProviderConfig(tt.schemaVersion, tt.baseURL, tt.errorTemplate)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if config == nil {
				t.Errorf("expected config to be non-nil")
				return
			}

			// Test passed - config creation works correctly
		})
	}
}

func TestProviderConfigureFunction(t *testing.T) {
	tests := []struct {
		name          string
		configData    map[string]interface{}
		expectError   bool
		errorContains string
	}{
		{
			name: "valid configuration with all fields",
			configData: map[string]interface{}{
				"schema_version":        "draft-07",
				"base_url":             "https://example.com/schemas/",
				"error_message_template": "Error in {schema}: {error}",
			},
			expectError: false,
		},
		{
			name: "valid configuration with defaults",
			configData: map[string]interface{}{
				"schema_version":        "draft/2020-12",
				"base_url":             "",
				"error_message_template": "JSON Schema validation failed: {error}",
			},
			expectError: false,
		},
		{
			name: "invalid schema version",
			configData: map[string]interface{}{
				"schema_version":        "invalid-draft",
				"base_url":             "",
				"error_message_template": "",
			},
			expectError:   true,
			errorContains: "unsupported JSON Schema version",
		},
		{
			name: "empty configuration (should use defaults)",
			configData: map[string]interface{}{
				"schema_version":        "",
				"base_url":             "",
				"error_message_template": "",
			},
			expectError: false,
		},
		{
			name: "file path as base URL",
			configData: map[string]interface{}{
				"schema_version":        "draft-06",
				"base_url":             "./local/schemas/",
				"error_message_template": "Validation error: {error}",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock resource data
			provider := Provider()
			resourceData := schema.TestResourceDataRaw(t, provider.Schema, tt.configData)

			// Call providerConfigure
			result, diags := providerConfigure(nil, resourceData)

			if tt.expectError {
				if !diags.HasError() {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" {
					found := false
					for _, diag := range diags {
						if strings.Contains(diag.Summary, tt.errorContains) || strings.Contains(diag.Detail, tt.errorContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error to contain %q, got diagnostics: %v", tt.errorContains, diags)
					}
				}
				return
			}

			if diags.HasError() {
				t.Errorf("unexpected error: %v", diags)
				return
			}

			if result == nil {
				t.Errorf("expected result to be non-nil")
				return
			}

			// Verify result is a ProviderConfig
			config, ok := result.(*ProviderConfig)
			if !ok {
				t.Errorf("expected result to be *ProviderConfig, got %T", result)
				return
			}

			// Verify configuration values are correctly set
			expectedSchemaVersion := tt.configData["schema_version"].(string)
			// When schema_version is empty, it's stored as-is (empty string) in DefaultSchemaVersion
			// The default draft is set to Draft2020 in the DefaultDraft field instead
			
			if config.DefaultSchemaVersion != expectedSchemaVersion {
				t.Errorf("expected schema version %q, got %q", expectedSchemaVersion, config.DefaultSchemaVersion)
			}

			expectedBaseURL := tt.configData["base_url"].(string)
			if config.DefaultBaseURL != expectedBaseURL {
				t.Errorf("expected base URL %q, got %q", expectedBaseURL, config.DefaultBaseURL)
			}

			expectedErrorTemplate := tt.configData["error_message_template"].(string)
			if expectedErrorTemplate == "" {
				expectedErrorTemplate = "JSON Schema validation failed: {error}" // Default
			}
			if config.DefaultErrorTemplate != expectedErrorTemplate {
				t.Errorf("expected error template %q, got %q", expectedErrorTemplate, config.DefaultErrorTemplate)
			}
		})
	}
}

func TestProviderSchemaDefinition(t *testing.T) {
	provider := Provider()

	// Test that all expected schema fields exist
	expectedFields := []string{"schema_version", "base_url", "error_message_template"}
	
	for _, field := range expectedFields {
		if _, exists := provider.Schema[field]; !exists {
			t.Errorf("expected schema field %q not found", field)
		}
	}

	// Test schema_version field properties
	schemaVersionField := provider.Schema["schema_version"]
	if schemaVersionField.Type != schema.TypeString {
		t.Errorf("expected schema_version to be TypeString, got %v", schemaVersionField.Type)
	}
	if schemaVersionField.Default != "draft/2020-12" {
		t.Errorf("expected schema_version default to be 'draft/2020-12', got %v", schemaVersionField.Default)
	}
	if !schemaVersionField.Optional {
		t.Errorf("expected schema_version to be optional")
	}

	// Test base_url field properties
	baseURLField := provider.Schema["base_url"]
	if baseURLField.Type != schema.TypeString {
		t.Errorf("expected base_url to be TypeString, got %v", baseURLField.Type)
	}
	if baseURLField.Default != "" {
		t.Errorf("expected base_url default to be empty string, got %v", baseURLField.Default)
	}

	// Test error_message_template field properties
	errorTemplateField := provider.Schema["error_message_template"]
	if errorTemplateField.Type != schema.TypeString {
		t.Errorf("expected error_message_template to be TypeString, got %v", errorTemplateField.Type)
	}
	if errorTemplateField.Default != "JSON Schema validation failed: {error}" {
		t.Errorf("expected error_message_template default to be 'JSON Schema validation failed: {error}', got %v", errorTemplateField.Default)
	}
}

var providerFactories = map[string]func() (*schema.Provider, error){
	"jsonschema": func() (*schema.Provider, error) {
		return Provider(), nil
	},
}

func testAccPreCheck(t *testing.T) {}
