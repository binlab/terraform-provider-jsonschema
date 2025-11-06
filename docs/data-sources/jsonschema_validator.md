# jsonschema_validator Data Source

The `jsonschema_validator` data source validates JSON or JSON5 documents using [JSON Schema](https://json-schema.org/) specifications.

## Example Usage

### Basic Validation

```hcl-terraform
data "jsonschema_validator" "config" {
  document = file("${path.module}/config.json")
  schema   = "${path.module}/config.schema.json"
}
```

### JSON5 Document and Schema

```hcl-terraform
data "jsonschema_validator" "json5_example" {
  document = <<-EOT
    {
      // JSON5 comments are supported
      "name": "example-service",
      "ports": [8080, 9090,], // Trailing commas allowed
      "config": {
        enabled: true,  // Unquoted keys supported
        timeout: 30_000 // Numeric separators supported
      }
    }
  EOT
  schema = "${path.module}/service.schema.json5"
}
```

### Schema Version Override

```hcl-terraform
data "jsonschema_validator" "legacy_validation" {
  document       = file("legacy-data.json")
  schema         = "${path.module}/legacy.schema.json"
  schema_version = "draft-04"  # Override provider default
}
```

### Remote Schema with Base URL

```hcl-terraform
data "jsonschema_validator" "remote_validation" {
  document = file("config.json")
  schema   = "api/v1/config.schema.json"
  base_url = "https://schemas.example.com/"  # Resolves to https://schemas.example.com/api/v1/config.schema.json
}
```

### Custom Error Message Templates

```hcl-terraform
# Simple error template
data "jsonschema_validator" "simple_errors" {
  document = file("config.json")
  schema   = "config.schema.json"
  error_message_template = "Config validation failed: {error}"
}

# Detailed error information
data "jsonschema_validator" "detailed_errors" {
  document = file("config.json")
  schema   = "config.schema.json"
  error_message_template = <<-EOT
    Validation Error:
    - Message: {{.Error}}
    - Schema File: {{.Schema}}
    - JSON Path: {{.Path}}
  EOT
}

# CI/CD integration format
data "jsonschema_validator" "ci_errors" {
  document = file("deployment.json")
  schema   = "deployment.schema.json"
  error_message_template = "::error file={{.Schema}},line=1::{{.Error}}"
}
```

## Argument Reference

* `document` (Required) - Content of a JSON or JSON5 document to validate. Supports both inline content and `file()` function.
* `schema` (Required) - Path to a JSON or JSON5 schema file. Can be a local file path or a URL (when `base_url` is configured).
* `schema_version` (Optional) - JSON Schema version override for this specific validation. Overrides the provider's default `schema_version`. Supported values: `"draft-04"`, `"draft-06"`, `"draft-07"`, `"draft/2019-09"`, `"draft/2020-12"`.
* `base_url` (Optional) - Base URL for resolving relative `$ref` URIs in schemas for this specific validation. Overrides the provider's default `base_url`.
* `error_message_template` (Optional) - Template for formatting validation error messages. Overrides the provider's default template. Available variables: `{{.Error}}`, `{{.Schema}}`, `{{.Document}}`, `{{.Path}}` (Go template syntax) or `{error}`, `{schema}`, `{document}`, `{path}` (simple syntax).

## Attributes Reference

* `validated` - The validated document in canonical JSON format. This is the `document` content parsed, validated, and re-serialized as standard JSON (even if the input was JSON5).

## Schema File Resolution

- **Local files**: Schema paths are resolved relative to the Terraform configuration directory
- **Remote schemas**: When `base_url` is configured (either at data source or provider level), schema paths are resolved against the base URL
- **Absolute URLs**: Full URLs in the `schema` argument are used as-is
- **Resolution priority**: Data source `base_url` takes precedence over provider `base_url`

## JSON5 Features Supported

Both `document` content and schema files support JSON5 syntax:

- **Comments**: `// single-line` and `/* multi-line */` comments
- **Trailing commas**: Arrays and objects can have trailing commas
- **Unquoted keys**: Object keys don't require quotes (when they're valid identifiers)
- **Single quotes**: Strings can use single quotes
- **Multi-line strings**: Strings can span multiple lines
- **Numeric literals**: Hexadecimal numbers, leading/trailing decimal points, numeric separators

## Error Handling

Validation failures provide detailed error messages including:
- The specific validation rule that failed
- The location in the document where validation failed
- Expected vs. actual values
- JSON Schema path references
