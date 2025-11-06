package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderNew(t *testing.T) {
	provider := New("test")()
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test that schema_version has correct default
	schemaVersionSchema := provider.Schema["schema_version"]
	if schemaVersionSchema.Default != "draft/2020-12" {
		t.Errorf("Expected default schema_version to be 'draft/2020-12', got %v", schemaVersionSchema.Default)
	}

	// Test that base_url exists
	if _, exists := provider.Schema["base_url"]; !exists {
		t.Error("Expected base_url to be in provider schema")
	}
}

var providerFactories = map[string]func() (*schema.Provider, error){
	"jsonschema": func() (*schema.Provider, error) {
		return Provider(), nil
	},
}

func testAccPreCheck(t *testing.T) {}
