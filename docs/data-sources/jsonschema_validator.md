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

### Schema with References

```hcl-terraform
# Schema file with $ref references resolved relative to schema location
data "jsonschema_validator" "with_refs" {
  document = file("config.json")
  schema   = "${path.module}/schemas/main.schema.json"  # Contains $ref: "./types.json"
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

### Enhanced Error Output

```hcl-terraform
# Detailed errors (enabled by default)
data "jsonschema_validator" "api_config" {
  document = file("config.json")
  schema   = "config.schema.json"
  # detailed_errors = true  # Default behavior
}

# Disable detailed errors for simple output
data "jsonschema_validator" "simple_validation" {
  document        = file("config.json")
  schema          = "config.schema.json"
  detailed_errors = false  # Simple error messages
}
```

## Argument Reference

* `document` (Required) - JSON or JSON5 document content to validate.
* `schema` (Required) - Path to JSON Schema file.
* `schema_version` (Optional) - Schema version override (`"draft-04"` to `"draft/2020-12"`).
* `detailed_errors` (Optional) - Enhanced error output. Defaults to provider setting.
* `error_message_template` (Optional) - Custom error template. Variables: `{error}`, `{schema}`, `{path}`, `{document}`, `{details}`, `{basic_output}`, `{detailed_output}`.

## Attributes Reference

* `validated` - The validated document in canonical JSON format. This is the `document` content parsed, validated, and re-serialized as standard JSON (even if the input was JSON5).

## Schema File Resolution

- **Local files**: Schema paths are resolved relative to the Terraform configuration directory
- **Schema references**: `$ref` URIs in schemas are resolved relative to the schema file's location
- **Relative references**: For example, if your schema is at `./schemas/main.schema.json` and contains `"$ref": "./types.json"`, it resolves to `./schemas/types.json`
- **Absolute references**: Full file paths or URLs in `$ref` are used as-is

## JSON5 Features Supported

Both `document` content and schema files support JSON5 syntax:

- **Comments**: `// single-line` and `/* multi-line */` comments
- **Trailing commas**: Arrays and objects can have trailing commas
- **Unquoted keys**: Object keys don't require quotes (when they're valid identifiers)
- **Single quotes**: Strings can use single quotes
- **Multi-line strings**: Strings can span multiple lines
- **Numeric literals**: Hexadecimal numbers, leading/trailing decimal points, numeric separators

## Error Handling

### Basic Error Format (detailed_errors = false)

Provides simplified error messages:
- Concise validation failure summary
- Schema reference location
- Easy-to-read format for quick troubleshooting

### Detailed Error Format (detailed_errors = true)

Provides comprehensive error information:
- Specific validation rule that failed
- Exact location in the document where validation failed
- Expected vs. actual values with context
- JSON Schema path references with full details
- Structured JSON output for machine processing

### Template Variables for Error Messages

When `detailed_errors = false` (default):
- `{error}` - Simplified error message
- `{schema}` - Schema file path
- `{document}` - Document content (truncated)
- `{path}` - JSON path where error occurred

When `detailed_errors = true`:
- `{error}` - Detailed error message with full context
- `{details}` - Human-readable verbose error breakdown
- `{basic_output}` - Flat JSON list of all validation errors
- `{detailed_output}` - Hierarchical JSON structure of validation errors
- All basic variables above are also available

### Structured Output Format

The structured JSON outputs (`basic_output` and `detailed_output`) are useful for:
- Automated error processing in CI/CD pipelines
- Integration with error monitoring systems
- Building custom error reporting tools
- Programmatic analysis of validation failures
