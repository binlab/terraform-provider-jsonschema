package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// ErrorContext holds data available for error message templating
type ErrorContext struct {
	Error          string // The original validation error message
	Schema         string // Path to the schema file
	Document       string // The document content (truncated if too long)
	Path           string // JSON path where the error occurred
	Details        string // Detailed error information (verbose format)
	BasicOutput    string // Basic output format (flat list of errors)
	DetailedOutput string // Detailed output format (hierarchical structure)
}

// FormatValidationError formats a validation error using the provided template
func FormatValidationError(err error, schemaPath, document, errorTemplate string, detailedErrors ...bool) error {
	if err == nil {
		return nil // Handle nil error gracefully
	}

	if errorTemplate == "" {
		// No template provided, use sensible default
		errorTemplate = "JSON Schema validation failed: {error}"
	}

	// Check if detailed errors are enabled (default to true for better user experience)
	enableDetailedErrors := true
	if len(detailedErrors) > 0 {
		enableDetailedErrors = detailedErrors[0]
	}

	// Extract detailed error information
	var errorMsg, jsonPath, details, basicOutput, detailedOutput string
	if validationErr, ok := err.(*jsonschema.ValidationError); ok {
		// Always use Error() method for error message (v6 doesn't have Message field)
		errorMsg = validationErr.Error()
		
		// v6 has InstanceLocation as []string, join with "/"
		if len(validationErr.InstanceLocation) > 0 {
			jsonPath = "/" + strings.Join(validationErr.InstanceLocation, "/")
		}
		
		// Only generate structured output if detailed errors are enabled
		if enableDetailedErrors {
			// Get structured detailed output
			detailedOut := validationErr.DetailedOutput()
			if detailedJSON, err := json.Marshal(detailedOut); err == nil {
				detailedOutput = string(detailedJSON)
			}

			// Get basic output (flat list of errors)
			basicOut := validationErr.BasicOutput()
			if basicJSON, err := json.Marshal(basicOut); err == nil {
				basicOutput = string(basicJSON)
			}
		}

		// Use the verbose format (similar to GoString) for details
		details = validationErr.GoString()
	} else {
		// For non-validation errors, use the basic error message
		errorMsg = err.Error()
	}

	// Truncate document if it's too long for template context
	truncatedDoc := document
	if len(document) > 200 {
		truncatedDoc = document[:200] + "..."
	}

	// Create template context
	ctx := ErrorContext{
		Error:          errorMsg,
		Schema:         schemaPath,
		Document:       truncatedDoc,
		Path:           jsonPath,
		Details:        details,
		BasicOutput:    basicOutput,
		DetailedOutput: detailedOutput,
	}

	// First try simple string replacement (faster for simple cases)
	if !strings.Contains(errorTemplate, "{{") {
		// Simple string templates with placeholders like {error}, {schema}, etc.
		result := errorTemplate
		result = strings.ReplaceAll(result, "{error}", ctx.Error)
		result = strings.ReplaceAll(result, "{schema}", ctx.Schema)
		result = strings.ReplaceAll(result, "{document}", ctx.Document)
		result = strings.ReplaceAll(result, "{path}", ctx.Path)
		result = strings.ReplaceAll(result, "{details}", ctx.Details)
		result = strings.ReplaceAll(result, "{basic_output}", ctx.BasicOutput)
		result = strings.ReplaceAll(result, "{detailed_output}", ctx.DetailedOutput)
		return fmt.Errorf("%s", result)
	}

	// Use Go text/template for more advanced templating
	tmpl, err := template.New("error").Parse(errorTemplate)
	if err != nil {
		// Template parsing failed, fall back to simple format
		return fmt.Errorf("validation failed (template error: %v): %s", err, errorMsg)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		// Template execution failed, fall back to simple format
		return fmt.Errorf("validation failed (template execution error: %v): %s", err, errorMsg)
	}

	return fmt.Errorf("%s", buf.String())
}

// Common error message templates that users can reference
var CommonErrorTemplates = map[string]string{
	"simple":            "Validation failed: {error}",
	"detailed":          "JSON Schema validation failed:\n  Error: {error}\n  Schema: {schema}\n  Path: {path}",
	"compact":           "[{schema}] {error} at {path}",
	"ci":                "::error file={schema},line=1::{error}",
	"json":              `{"error": "{error}", "schema": "{schema}", "path": "{path}"}`,
	"verbose":           "Validation failed: {error}\n\nDetails:\n{details}",
	"structured_basic":  "Validation failed: {error}\n\nBasic Output:\n{basic_output}",
	"structured_full":   "Validation failed: {error}\n\nDetailed Output:\n{detailed_output}",
	"debug":             "Schema: {schema}\nDocument: {document}\nPath: {path}\nError: {error}\n\nVerbose Details:\n{details}",
}

// GetCommonTemplate returns a predefined error template by name
func GetCommonTemplate(name string) (string, bool) {
	template, exists := CommonErrorTemplates[name]
	return template, exists
}