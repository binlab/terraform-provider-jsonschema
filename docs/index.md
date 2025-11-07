# JSON Schema Provider

Terraform provider for validating JSON and JSON5 documents using [JSON Schema](https://json-schema.org/) specifications.

## Features

- **JSON5 Support**: Parse and validate both JSON and JSON5 format documents and schemas
- **Multiple Schema Versions**: Support for JSON Schema Draft 4, 6, 7, 2019-09, and 2020-12
- **External Reference Resolution**: Resolves `$ref` URIs including JSON5 files relative to schema location
- **Enhanced Error Templating**: Go template system with individual error iteration capabilities
- **Consistent Error Ordering**: Deterministic error ordering for reliable testing and CI/CD
- **JSON5 External References**: Support for JSON5 files in `$ref` schema references
- **Robust Validation**: Powered by `santhosh-tekuri/jsonschema/v6` for comprehensive validation

## Provider Configuration

```hcl-terraform
provider "jsonschema" {
  schema_version = "draft/2020-12"  # Optional: JSON Schema version
  error_message_template = "{{.FullMessage}}"  # Optional: Go template for errors
}
```

### Configuration Arguments

- `schema_version` (Optional) - JSON Schema draft version. Defaults to `"draft/2020-12"`.
- `error_message_template` (Optional) - Go template for error messages. Available variables: `{{.Schema}}`, `{{.Document}}`, `{{.FullMessage}}`, `{{.Errors}}`, `{{.ErrorCount}}`. Use `{{range .Errors}}` to iterate over individual errors.

## Basic Example

```hcl-terraform
provider "jsonschema" {
  schema_version = "draft/2020-12"
}

# Validate a JSON document
data "jsonschema_validator" "config" {
  document = file("${path.module}/config.json")
  schema   = "${path.module}/config.schema.json"
}

# Use the validated document
resource "helm_release" "app" {
  name   = "my-app"
  values = [data.jsonschema_validator.config.validated]
}
```

## JSON5 Support Example

```hcl-terraform
# Validate a JSON5 document with JSON5 schema
data "jsonschema_validator" "json5_config" {
  document = <<-EOT
    {
      // JSON5 comments supported
      "name": "my-service",
      "ports": [8080, 8081,], // Trailing commas allowed
      "features": {
        enabled: true,  // Unquoted keys supported
      }
    }
  EOT
  schema = "${path.module}/service.schema.json5"
}
```

## Advanced Configuration

```hcl-terraform
# Override schema version per validation
data "jsonschema_validator" "legacy_config" {
  document       = file("legacy-config.json")
  schema         = "legacy.schema.json"
  schema_version = "draft-04"  # Override provider default
}

# Custom error message template per validation
data "jsonschema_validator" "detailed_validation" {
  document = file("config.json")
  schema   = "config.schema.json"
  error_message_template = "Schema {{.Schema}} failed: {{.FullMessage}}"
}

# Individual error iteration for multiple validation errors
data "jsonschema_validator" "individual_errors" {
  document = file("complex-config.json")
  schema   = "complex.schema.json"
  error_message_template = "{{range .Errors}}{{.Path}}: {{.Message}}\n{{end}}"
}

# Schema references are resolved relative to schema file location
# For example, if schema is at "/path/to/schemas/main.schema.json"
# then "$ref": "./types.json" resolves to "/path/to/schemas/types.json"
data "jsonschema_validator" "with_refs" {
  document = file("document.json")
  schema   = "${path.module}/schemas/main.schema.json"  # Contains $ref references
}

# Multiple schema validations with different versions
data "jsonschema_validator" "api_validation" {
  document       = file("api-config.json")
  schema         = "${path.module}/schemas/openapi/config.schema.json"
  schema_version = "draft/2019-09"
}

data "jsonschema_validator" "legacy_validation" {
  document       = file("legacy-config.json") 
  schema         = "${path.module}/schemas/legacy/service.schema.json"
  schema_version = "draft-04"
}
```

## Error Message Templating

Customize validation error messages using Go templates:

```hcl-terraform
# Default full message
data "jsonschema_validator" "config" {
  document = file("config.json")
  schema   = "config.schema.json"
  error_message_template = "{{.FullMessage}}"
}

# Individual error iteration
data "jsonschema_validator" "individual_errors" {
  document = file("api-config.json")
  schema   = "api.schema.json"
  error_message_template = "{{range .Errors}}{{.Path}}: {{.Message}}\n{{end}}"
}

# Detailed format with error count
data "jsonschema_validator" "detailed_format" {
  document = file("config.json")
  schema   = "config.schema.json"
  error_message_template = <<-EOT
    Found {{.ErrorCount}} validation errors in {{.Schema}}:
    {{range $i, $e := .Errors}}{{add $i 1}}. {{.Path}}: {{.Message}}
    {{end}}
  EOT
}

# CI/CD friendly format
provider "jsonschema" {
  error_message_template = "{{range .Errors}}::error file={{$.Schema}}::{{.Message}}{{if .Path}} at {{.Path}}{{end}}\n{{end}}"
}

# JSON structured output
data "jsonschema_validator" "structured_errors" {
  document = file("config.json")
  schema   = "config.schema.json"
  error_message_template = <<-EOT
    {"validation_failed":true,"schema":"{{.Schema}}","error_count":{{.ErrorCount}},"errors":[{{range $i,$e := .Errors}}{{if $i}},{{end}}{"path":"{{.Path}}","message":"{{.Message}}"}{{end}}]}
  EOT
}
```
