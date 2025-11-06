package provider

import (
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// ProviderConfig holds the provider-level configuration
type ProviderConfig struct {
	// DefaultSchemaVersion specifies the JSON Schema version to use when not specified in the schema
	DefaultSchemaVersion string
	
	// DefaultBaseURL is used as the default base URL for resolving $ref URIs
	DefaultBaseURL string
	
	// DefaultErrorTemplate is the default error message template
	DefaultErrorTemplate string
	
	// DefaultDraft is the default draft to use
	DefaultDraft *jsonschema.Draft
}

// NewProviderConfig creates a new provider configuration with defaults
func NewProviderConfig(schemaVersion, baseURL, errorTemplate string) (*ProviderConfig, error) {
	// Set sensible default for error template if empty
	if errorTemplate == "" {
		errorTemplate = "JSON Schema validation failed: {error}"
	}
	
	config := &ProviderConfig{
		DefaultSchemaVersion: schemaVersion,
		DefaultBaseURL:       baseURL,
		DefaultErrorTemplate: errorTemplate,
		DefaultDraft:         jsonschema.Draft2020, // Default to latest draft
	}
	
	// Set draft based on schema version if provided
	if schemaVersion != "" {
		draft, err := GetDraftForVersion(schemaVersion)
		if err != nil {
			return nil, err
		}
		config.DefaultDraft = draft
	}
	
	// Note: base_url validation is intentionally omitted here
	// The jsonschema library will handle URL validation and path resolution
	// This allows maximum flexibility for users (HTTP URLs, file:// URLs, relative paths, etc.)
	
	return config, nil
}

// GetDraftForVersion returns the appropriate draft for a given schema version string
func GetDraftForVersion(version string) (*jsonschema.Draft, error) {
	switch version {
	case "draft-04", "http://json-schema.org/draft-04/schema#":
		return jsonschema.Draft4, nil
	case "draft-06", "http://json-schema.org/draft-06/schema#":
		return jsonschema.Draft6, nil
	case "draft-07", "http://json-schema.org/draft-07/schema#":
		return jsonschema.Draft7, nil
	case "draft/2019-09", "https://json-schema.org/draft/2019-09/schema":
		return jsonschema.Draft2019, nil
	case "draft/2020-12", "https://json-schema.org/draft/2020-12/schema":
		return jsonschema.Draft2020, nil
	default:
		return jsonschema.Draft2020, fmt.Errorf("unsupported JSON Schema version: %s", version)
	}
}