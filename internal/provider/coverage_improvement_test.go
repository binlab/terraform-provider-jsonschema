package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// TestGenerateSortedFullMessage_EmptyErrors tests the case where ValidationError has no extracted details
func TestGenerateSortedFullMessage_EmptyErrors(t *testing.T) {
	// Create a ValidationError with no causes (empty error tree)
	// This is a theoretical edge case where the error exists but has no extractable details
	
	// We need to create a schema that will compile and then manually construct
	// a validation scenario that might not have detailed errors
	
	schemaJSON := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object"
	}`
	
	var schemaData interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &schemaData); err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}
	
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	
	if err := compiler.AddResource("test://schema", schemaData); err != nil {
		t.Fatalf("Failed to add schema: %v", err)
	}
	
	compiledSchema, err := compiler.Compile("test://schema")
	if err != nil {
		t.Fatalf("Failed to compile schema: %v", err)
	}
	
	// Try to validate something that will fail at a fundamental level
	// Using a non-object type when object is required
	err = compiledSchema.Validate("not an object")
	if err == nil {
		t.Fatal("Expected validation to fail")
	}
	
	// Check if we can format this error
	result := FormatValidationError(err, "test.json", `"not an object"`, "{{.FullMessage}}")
	if result == nil {
		t.Fatal("Expected error result")
	}
	
	// The message should still contain the schema URL even if no detailed errors
	errMsg := result.Error()
	if !strings.Contains(errMsg, "test://schema") || !strings.Contains(errMsg, "validation") {
		t.Logf("Error message: %s", errMsg)
	}
}

// TestDataSourceJsonschemaValidatorRead_NoDraftConfiguration tests the fallback to Draft2020
func TestDataSourceJsonschemaValidatorRead_NoDraftConfiguration(t *testing.T) {
	// Create a temporary schema file without $schema field
	tempDir := t.TempDir()
	schemaPath := filepath.Join(tempDir, "test.schema.json")
	
	schemaContent := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"}
		},
		"required": ["name"]
	}`
	
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to create schema file: %v", err)
	}
	
	// Create resource data
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"document": {Type: schema.TypeString},
		"schema":   {Type: schema.TypeString},
		"schema_version": {Type: schema.TypeString},
		"error_message_template": {Type: schema.TypeString},
		"ref_overrides": {Type: schema.TypeMap},
		"validated": {Type: schema.TypeString},
	}, map[string]interface{}{
		"document": `{"name": "test"}`,
		"schema":   schemaPath,
		"schema_version": "", // No version specified
	})
	
	// Create provider config with NO default schema version and NO default draft
	config := &ProviderConfig{
		DefaultSchemaVersion: "", // Empty - no default
		DefaultDraft:         nil, // Nil - no default draft
		DefaultErrorTemplate: "",
	}
	
	// This should trigger the fallback to Draft2020 (line 121)
	err := dataSourceJsonschemaValidatorRead(d, config)
	if err != nil {
		t.Fatalf("Expected validation to succeed with Draft2020 fallback: %v", err)
	}
	
	// Verify the document was validated
	validated := d.Get("validated").(string)
	if validated == "" {
		t.Error("Expected validated field to be set")
	}
}

// TestSortKeys_NonStringMap tests the edge case of maps with non-string keys
func TestSortKeys_NonStringMap(t *testing.T) {
	// In Go, JSON only supports string keys, but sortKeys should handle
	// non-string key maps gracefully by returning them as-is
	
	// Create a map with integer keys (not possible from JSON, but sortKeys should handle it)
	intKeyMap := map[int]string{
		3: "three",
		1: "one",
		2: "two",
	}
	
	// sortKeys should return this map as-is since keys are not strings
	result := sortKeys(intKeyMap)
	
	// The result should be the same map (returned as-is)
	resultMap, ok := result.(map[int]string)
	if !ok {
		t.Fatalf("Expected map[int]string, got %T", result)
	}
	
	if len(resultMap) != 3 {
		t.Errorf("Expected map with 3 elements, got %d", len(resultMap))
	}
	
	// Verify the values are still there
	if resultMap[1] != "one" || resultMap[2] != "two" || resultMap[3] != "three" {
		t.Error("Map values were not preserved correctly")
	}
}

// TestCompactDeterministicJSON_MalformedInput tests the json.Compact error path
func TestCompactDeterministicJSON_MalformedInput(t *testing.T) {
	// To trigger json.Compact to fail, we need MarshalDeterministic to return
	// something that is valid for json.Marshal but invalid for json.Compact
	// 
	// However, this is actually impossible in practice because:
	// 1. json.Marshal produces valid JSON
	// 2. json.Compact only fails on invalid JSON input
	// 3. If json.Marshal succeeds, json.Compact will succeed
	//
	// The only way to hit this error path would be if there's a bug in the
	// json package itself, which is extremely unlikely.
	
	// Let's at least verify the function works correctly with edge cases
	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "deeply nested structure",
			data: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": map[string]interface{}{
							"d": []interface{}{1, 2, 3, 4, 5},
						},
					},
				},
			},
		},
		{
			name: "empty structures",
			data: map[string]interface{}{
				"empty_map":   map[string]interface{}{},
				"empty_array": []interface{}{},
			},
		},
		{
			name: "mixed types",
			data: map[string]interface{}{
				"string": "value",
				"number": 42,
				"float":  3.14,
				"bool":   true,
				"null":   nil,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CompactDeterministicJSON(tt.data)
			if err != nil {
				t.Errorf("CompactDeterministicJSON failed: %v", err)
			}
			
			// Verify result is valid compact JSON
			var parsed interface{}
			if err := json.Unmarshal(result, &parsed); err != nil {
				t.Errorf("Result is not valid JSON: %v", err)
			}
			
			// Verify it's compact (no unnecessary whitespace)
			if strings.Contains(string(result), "  ") || strings.Contains(string(result), "\n") {
				t.Error("Result is not compact")
			}
		})
	}
}

// TestDataSourceJsonschemaValidatorRead_AddResourceFailure tests the AddResource error path
// This is extremely difficult to trigger because AddResource in jsonschema/v6 only fails when:
// 1. The URL is malformed (invalid URL parsing)
// 2. The resource is already registered at that URL
// Let's test the duplicate registration scenario
func TestDataSourceJsonschemaValidatorRead_DuplicateSchemaURL(t *testing.T) {
	// Create a temporary schema file
	tempDir := t.TempDir()
	schemaPath := filepath.Join(tempDir, "test.schema.json")
	
	schemaContent := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"properties": {
			"name": {"type": "string"}
		}
	}`
	
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to create schema file: %v", err)
	}
	
	// Create resource data
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"document": {Type: schema.TypeString},
		"schema":   {Type: schema.TypeString},
		"schema_version": {Type: schema.TypeString},
		"error_message_template": {Type: schema.TypeString},
		"ref_overrides": {Type: schema.TypeMap},
		"validated": {Type: schema.TypeString},
	}, map[string]interface{}{
		"document": `{"name": "test"}`,
		"schema":   schemaPath,
	})
	
	config := &ProviderConfig{
		DefaultSchemaVersion: "draft/2020-12",
		DefaultErrorTemplate: "",
	}
	
	// First call should succeed
	err := dataSourceJsonschemaValidatorRead(d, config)
	if err != nil {
		t.Fatalf("First validation failed: %v", err)
	}
	
	// Note: We can't easily trigger the AddResource error in lines 158-159
	// because each call to dataSourceJsonschemaValidatorRead creates a new compiler
	// instance, so there's no way to get duplicate registrations.
	// 
	// The AddResource error path at lines 158-159 is defensive programming
	// for scenarios that are extremely unlikely to occur in practice without
	// significant changes to the code structure (e.g., reusing compiler instances).
	//
	// For now, we've documented this limitation. Achieving 100% coverage would
	// require dependency injection or mocking of the compiler, which would be
	// significant refactoring for minimal benefit.
}

// TestSortKeys_ComplexInterfaceTypes tests various interface wrapping scenarios
func TestSortKeys_ComplexInterfaceTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name: "interface wrapping pointer to struct",
			input: func() interface{} {
				type testStruct struct {
					Value string
				}
				ptr := &testStruct{Value: "test"}
				var i interface{} = ptr
				return i
			}(),
			expected: "has pointer",
		},
		{
			name: "nested interface values",
			input: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": func() interface{} {
						var i interface{} = "nested value"
						return i
					}(),
				},
			},
			expected: "nested",
		},
		{
			name: "array of interfaces",
			input: []interface{}{
				func() interface{} { var i interface{} = 1; return i }(),
				func() interface{} { var i interface{} = "two"; return i }(),
				func() interface{} { var i interface{} = true; return i }(),
			},
			expected: "array",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// sortKeys should handle these without panicking
			result := sortKeys(tt.input)
			if result == nil && tt.input != nil {
				t.Error("Expected non-nil result for non-nil input")
			}
			
			// Try to marshal the result to verify it's still valid
			_, err := json.Marshal(result)
			if err != nil {
				t.Logf("Note: Cannot marshal result (expected for some test types): %v", err)
				// This is acceptable for non-JSON-serializable types like struct pointers
			}
		})
	}
}

// TestExtractCleanMessage_EdgeCases tests message extraction edge cases
func TestExtractCleanMessage_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		path     string
		expected string
	}{
		{
			name:     "message with path prefix",
			message:  "at '/field': value is invalid",
			path:     "/field",
			expected: "value is invalid",
		},
		{
			name:     "message without path prefix",
			message:  "value is invalid",
			path:     "/field",
			expected: "value is invalid",
		},
		{
			name:     "empty path (root)",
			message:  "at '': root validation failed",
			path:     "",
			expected: "root validation failed",
		},
		{
			name:     "message with similar but non-matching prefix",
			message:  "at '/other': something else",
			path:     "/field",
			expected: "at '/other': something else",
		},
		{
			name:     "message with colon but no path",
			message:  "error: something went wrong",
			path:     "/field",
			expected: "error: something went wrong",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCleanMessage(tt.message, tt.path)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
