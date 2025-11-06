# JSON Schema Provider

Terraform provider for validating JSON and JSON5 documents using [JSON Schema](https://json-schema.org/) specifications.

## Features

- **JSON5 Support**: Parse and validate both JSON and JSON5 format documents and schemas
- **Multiple Schema Versions**: Support for JSON Schema Draft 4, 6, 7, 2019-09, and 2020-12
- **Flexible Reference Resolution**: Configurable base URLs for resolving `$ref` URIs
- **Robust Validation**: Powered by `santhosh-tekuri/jsonschema/v5` for comprehensive validation

## Provider Configuration

```hcl-terraform
provider "jsonschema" {
  # Default JSON Schema version (optional)
  # Supported: "draft-04", "draft-06", "draft-07", "draft/2019-09", "draft/2020-12"
  schema_version = "draft/2020-12"  # Default value
  
  # Base URL for resolving $ref URIs (optional)
  base_url = "https://example.com/schemas/"
}
```

### Configuration Arguments

- `schema_version` (Optional) - Default JSON Schema version to use when not specified in the schema document. Supported values: `"draft-04"`, `"draft-06"`, `"draft-07"`, `"draft/2019-09"`, `"draft/2020-12"`. Defaults to `"draft/2020-12"`.
- `base_url` (Optional) - Default base URL for resolving relative `$ref` URIs in schemas. This serves as a fallback when the data source doesn't specify its own `base_url`.

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

# Remote schema with per-resource base URL
data "jsonschema_validator" "remote_validation" {
  document = file("data.json")
  schema   = "api/v1/schema.json"
  base_url = "https://schemas.example.com/"  # Base URL for this validation
}

# Multiple schemas from different sources
data "jsonschema_validator" "api_validation" {
  document = file("api-config.json")
  schema   = "openapi/v3.1/config.schema.json"
  base_url = "https://api-schemas.company.com/"
}

data "jsonschema_validator" "internal_validation" {
  document = file("internal-config.json") 
  schema   = "internal/service.schema.json"
  base_url = "https://internal-schemas.company.com/"
}

# Provider-level base URL as fallback (optional)
provider "jsonschema" {
  base_url = "https://default-schemas.example.com/"  # Used when data source base_url not specified
}
```
