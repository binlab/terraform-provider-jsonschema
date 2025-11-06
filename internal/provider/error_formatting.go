package provider

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// ErrorContext holds data available for error message templating
type ErrorContext struct {
	Error    string // The original validation error message
	Schema   string // Path to the schema file
	Document string // The document content (truncated if too long)
	Path     string // JSON path where the error occurred
	Details  string // Detailed error information
}

// FormatValidationError formats a validation error using the provided template
func FormatValidationError(err error, schemaPath, document, errorTemplate string) error {
	if errorTemplate == "" {
		// No template provided, use sensible default
		errorTemplate = "JSON Schema validation failed: {error}"
	}

	// Extract detailed error information
	var errorMsg, jsonPath, details string
	if validationErr, ok := err.(*jsonschema.ValidationError); ok {
		errorMsg = validationErr.Message
		jsonPath = validationErr.InstanceLocation
		// Try to get detailed output if available
		detailedOutput := validationErr.DetailedOutput()
		details = fmt.Sprintf("%v", detailedOutput)
	} else {
		errorMsg = err.Error()
	}

	// Truncate document if it's too long for template context
	truncatedDoc := document
	if len(document) > 200 {
		truncatedDoc = document[:200] + "..."
	}

	// Create template context
	ctx := ErrorContext{
		Error:    errorMsg,
		Schema:   schemaPath,
		Document: truncatedDoc,
		Path:     jsonPath,
		Details:  details,
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
	"simple":   "Validation failed: {error}",
	"detailed": "JSON Schema validation failed:\n  Error: {error}\n  Schema: {schema}\n  Path: {path}",
	"compact":  "[{schema}] {error} at {path}",
	"ci":       "::error file={schema},line=1::{error}",
	"json":     `{"error": "{error}", "schema": "{schema}", "path": "{path}"}`,
}

// GetCommonTemplate returns a predefined error template by name
func GetCommonTemplate(name string) (string, bool) {
	template, exists := CommonErrorTemplates[name]
	return template, exists
}