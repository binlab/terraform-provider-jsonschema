# jsonschema_validator Data Source

The `jsonschema_validator` data source validates JSON or JSON5 documents using [JSON Schema](https://json-schema.org/) specifications with enhanced error templating capabilities.

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
# Default full message
data "jsonschema_validator" "default_errors" {
  document = file("config.json")
  schema   = "config.schema.json"
  error_message_template = "{{.FullMessage}}"
}

# Individual error iteration
data "jsonschema_validator" "individual_errors" {
  document = file("config.json")
  schema   = "config.schema.json"
  error_message_template = "{{range .Errors}}{{.Path}}: {{.Message}}\n{{end}}"
}

# Custom format with error count
data "jsonschema_validator" "counted_errors" {
  document = file("config.json")
  schema   = "config.schema.json"
  error_message_template = "Found {{.ErrorCount}} errors:\n{{range .Errors}}• {{.Path}}: {{.Message}}\n{{end}}"
}

# CI/CD integration format
data "jsonschema_validator" "ci_errors" {
  document = file("deployment.json")
  schema   = "deployment.schema.json"
  error_message_template = "{{range .Errors}}::error file={{$.Schema}}::{{.Message}}{{if .Path}} at {{.Path}}{{end}}\n{{end}}"
}
```

### Advanced Error Templates

```hcl-terraform
# Detailed format with schema information
data "jsonschema_validator" "detailed_format" {
  document = file("config.json")
  schema   = "config.schema.json"
  error_message_template = <<-EOT
    Schema: {{.Schema}}
    Errors: {{.ErrorCount}}
    {{range .Errors}}• {{.Path}}: {{.Message}}
    {{end}}
  EOT
}

# JSON format for structured logging
data "jsonschema_validator" "json_format" {
  document = file("config.json")
  schema   = "config.schema.json"
  error_message_template = jsonencode({
    "validation_failed": true,
    "schema": "{{.Schema}}",
    "error_count": "{{.ErrorCount}}",
    "errors": "{{range $i, $e := .Errors}}{{if $i}},{{end}}{\"path\":\"{{.Path}}\",\"message\":\"{{.Message}}\"}{{end}}"
  })
}
```

## Argument Reference

* `document` (Required) - JSON or JSON5 document content to validate.
* `schema` (Required) - Path to JSON or JSON5 schema file.
* `schema_version` (Optional) - Schema version override (`"draft-04"` to `"draft/2020-12"`).
* `error_message_template` (Optional) - Custom Go template for error messages. Available variables: `{{.Schema}}`, `{{.Document}}`, `{{.FullMessage}}`, `{{.Errors}}`, `{{.ErrorCount}}`.

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

## Error Templating

### Template Variables

Available in `error_message_template`:

- `{{.FullMessage}}` - Complete validation error message from jsonschema library
- `{{.ErrorCount}}` - Number of individual validation errors
- `{{.Errors}}` - Array of individual validation errors (for iteration)
- `{{.Document}}` - The document content (truncated if long)
- `{{.Schema}}` - Path to the schema file

### Individual Error Details

Each error in `{{.Errors}}` contains:

- `{{.Message}}` - Human-readable error message
- `{{.Path}}` - JSON path where the error occurred (e.g., `/user/name`)
- `{{.SchemaPath}}` - Schema path of the failing constraint
- `{{.Value}}` - The actual value that failed validation (if available)

### Template Examples

```hcl-terraform
# Simple full message (default)
error_message_template = "{{.FullMessage}}"

# Individual error iteration
error_message_template = "{{range .Errors}}{{.Path}}: {{.Message}}\n{{end}}"

# Numbered list format (one-based)
error_message_template = "{{.ErrorCount}} validation errors:\n{{range $i, $e := .Errors}}{{add $i 1}}. {{.Message}} (at {{.Path}})\n{{end}}"

# CI/CD format
error_message_template = "{{range .Errors}}::error file={{$.Schema}}::{{.Message}}{{if .Path}} at {{.Path}}{{end}}\n{{end}}"
```

### Template Functions

The following template function is available:

- `add` - Add two integers: `{{add $i 1}}` (useful for one-based indexing)

Templates support standard Go template syntax for formatting error messages.
