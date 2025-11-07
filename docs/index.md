# JSON Schema Provider

Terraform provider for validating JSON and JSON5 documents using [JSON Schema](https://json-schema.org/) specifications.

## Features

- **JSON5 Support**: Parse and validate both JSON and JSON5 format documents and schemas
- **Multiple Schema Versions**: Support for JSON Schema Draft 4, 6, 7, 2019-09, and 2020-12
- **Automatic Reference Resolution**: Resolves `$ref` URIs relative to schema file location
- **Custom Error Templates**: Customize validation error messages with templating support
- **Detailed Error Output**: Enhanced error reporting with structured JSON output for debugging
- **Flexible Error Control**: Configure error detail level at provider and resource level
- **Robust Validation**: Powered by `santhosh-tekuri/jsonschema/v5` for comprehensive validation

## Provider Configuration

```hcl-terraform
provider "jsonschema" {
  schema_version = "draft/2020-12"  # Optional: JSON Schema version
  detailed_errors = true            # Optional: Enhanced error output (default)
  error_message_template = "{error}"  # Optional: Custom error template
}
```

### Configuration Arguments

- `schema_version` (Optional) - JSON Schema draft version. Defaults to `"draft/2020-12"`.
- `detailed_errors` (Optional) - Enhanced error reporting with structured output. Defaults to `true`.
- `error_message_template` (Optional) - Error message template. Available variables: `{error}`, `{schema}`, `{path}`, `{document}`. When `detailed_errors=true`: `{details}`, `{basic_output}`, `{detailed_output}`.

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
  document               = file("config.json")
  schema                 = "config.schema.json"
  error_message_template = "Configuration error in {schema}: {error}"
}

# Enable detailed errors for specific validations
data "jsonschema_validator" "debug_validation" {
  document        = file("complex-config.json")
  schema          = "complex.schema.json"
  detailed_errors = true  # Override provider default for enhanced debugging
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

Customize validation error messages using templates:

```hcl-terraform
# Simple string replacement
data "jsonschema_validator" "config" {
  document = file("config.json")
  schema   = "config.schema.json"
  error_message_template = "Configuration validation failed: {error} (Schema: {schema})"
}

# Go template syntax with more control
data "jsonschema_validator" "api_config" {
  document = file("api-config.json")
  schema   = "api.schema.json"
  error_message_template = <<-EOT
    JSON Schema Validation Error:
    - Error: {{.Error}}
    - Schema: {{.Schema}}
    - JSON Path: {{.Path}}
    - Document: {{.Document}}
  EOT
}

# CI/CD friendly format
provider "jsonschema" {
  error_message_template = "::error file={schema},line=1::{error}"
}

# JSON structured output
data "jsonschema_validator" "structured_errors" {
  document = file("config.json")
  schema   = "config.schema.json"
  error_message_template = jsonencode({
    "error": "{error}",
    "schema": "{schema}", 
    "path": "{path}",
    "timestamp": "{{now}}"
  })
}
```
