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

var providerFactories = map[string]func() (*schema.Provider, error){
	"jsonschema": func() (*schema.Provider, error) {
		return Provider(), nil
	},
}

func testAccPreCheck(t *testing.T) {}
