=============================
terraform-provider-jsonschema
=============================

.. image:: https://codecov.io/github/iilei/terraform-provider-jsonschema/branch/master/graph/badge.svg
    :target: https://codecov.io/github/iilei/terraform-provider-jsonschema
    :alt: Coverage Status

.. contents::
    :local:
    :depth: 2


Abstract
========

A |terraform|_ provider for validating JSON and JSON5 documents using |json-schema|_ specifications.

Features
========

- **JSON5 Support**: Parse and validate both JSON and JSON5 format documents and schemas
- **Multiple Schema Versions**: Support for JSON Schema Draft 4, 6, 7, 2019-09, and 2020-12
- **Automatic Reference Resolution**: Resolves ``$ref`` URIs relative to schema file location
- **Custom Error Templates**: Customize validation error messages with templating support
- **Detailed Error Output**: Enhanced error reporting with structured JSON output for debugging
- **Flexible Error Control**: Configure error detail level at provider and resource level
- **Robust Validation**: Powered by ``santhosh-tekuri/jsonschema/v5`` for comprehensive validation
- **Deterministic Output**: Consistent JSON marshaling for stable resource IDs

Installation
============

On |terraform|_ versions 0.13+ use:

.. code-block:: terraform

  terraform {
    required_providers {
      jsonschema = {
        source  = "iilei/jsonschema"
        version = "~> 0.3.2"  // Use the latest stable version
      }
    }
  }

For |terraform|_ versions 0.12 or lower use instructions: |terraform-install-plugin|_

Usage
=====

See |user-docs|_ for details.

Provider Configuration
======================

.. code-block:: terraform

  provider "jsonschema" {
    schema_version = "draft/2020-12"  # Optional: JSON Schema version
    detailed_errors = true            # Optional: Enhanced error output (default)
    error_message_template = "Validation failed: {error}"  # Optional: Custom template
  }

Basic Example
=============

.. code-block:: terraform

  # Configure the provider
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

JSON5 Support Example
====================

.. code-block:: terraform

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

Advanced Configuration
=====================

.. code-block:: terraform

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

  # Enable detailed error output for specific validation
  data "jsonschema_validator" "debug_validation" {
    document        = file("complex-config.json")
    schema          = "complex.schema.json"
    detailed_errors = true  # Override provider default for detailed debugging
  }

  # Schema references are resolved relative to schema file location
  # For example, if schema is at "/path/to/schemas/main.schema.json"
  # then "$ref": "./types.json" resolves to "/path/to/schemas/types.json"
  data "jsonschema_validator" "with_refs" {
    document = file("document.json")
    schema   = "${path.module}/schemas/main.schema.json"  # Contains $ref references
  }

Detailed Error Output
====================

The provider supports enhanced error reporting with structured output for better debugging and automated error handling.

Basic vs Detailed Errors
------------------------

.. code-block:: terraform

  # Basic error format (default)
  data "jsonschema_validator" "basic" {
    detailed_errors = false
    document = jsonencode({"name": "Jo"})  # Too short
    schema = jsonencode({
      "type": "object",
      "properties": {
        "name": {"type": "string", "minLength": 3}
      }
    })
  }
  # Error: "doesn't validate with schema"

  # Detailed error format 
  data "jsonschema_validator" "detailed" {
    detailed_errors = true
    document = jsonencode({"name": "Jo"})  # Too short
    schema = jsonencode({
      "type": "object", 
      "properties": {
        "name": {"type": "string", "minLength": 3}
      }
    })
  }
  # Error: "jsonschema: '/name' does not validate with schema#/properties/name/minLength: length must be >= 3, but got 2"

Structured Error Output
----------------------

When ``detailed_errors = true``, additional template variables become available:

.. code-block:: terraform

  provider "jsonschema" {
    detailed_errors = true
    error_message_template = <<-EOT
      Validation failed: {error}
      
      Basic Output (JSON):
      {basic_output}
      
      Detailed Output (JSON):
      {detailed_output}
    EOT
  }

Available Template Variables
---------------------------

- ``{error}`` - The validation error message (simple or detailed based on ``detailed_errors``)
- ``{schema}`` - Path to the schema file
- ``{document}`` - The document content (truncated if long)  
- ``{path}`` - JSON path where validation failed
- ``{details}`` - Human-readable verbose error information (when detailed_errors = true)
- ``{basic_output}`` - Flat list of all errors in JSON format (when detailed_errors = true)
- ``{detailed_output}`` - Hierarchical error structure in JSON format (when detailed_errors = true)

Predefined Error Templates
-------------------------

The provider includes several predefined templates accessible via ``GetCommonTemplate()``:

- ``simple`` - "Validation failed: {error}"
- ``detailed`` - Multi-line format with error, schema, and path
- ``compact`` - "[{schema}] {error} at {path}"
- ``json`` - JSON formatted error information
- ``verbose`` - Includes detailed human-readable error breakdown
- ``structured_basic`` - Uses flat JSON error list
- ``structured_full`` - Uses hierarchical JSON error structure

Example with Structured Output
------------------------------

.. code-block:: terraform

  # Configure provider for machine-processable errors
  provider "jsonschema" {
    detailed_errors = true
    error_message_template = "structured_basic"  # Use predefined template
  }
  
  # This will output JSON formatted error information suitable for
  # processing by external tools, CI/CD systems, or error aggregators

Development
===========

This repository follows structure of |terraform-provider-scaffolding|_ template
recommended by |terraform|_ developers (see |terraform-publishing-provider|_).

For publishing it uses Gitlab Actions.

Environment requirements:

- |go|_ 1.24+ (to build the provider plugin)

Running tests:

.. code-block:: bash

  TF_ACC=1 go test ./internal/provider -v -timeout 5m


.. |terraform| replace:: Terraform
.. _terraform: https://www.terraform.io/

.. |terraform-install-plugin| replace:: install a terraform plugin
.. _terraform-install-plugin: https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin

.. |user-docs| replace:: User Documentation  
.. _user-docs: https://registry.terraform.io/providers/iilei/jsonschema/latest/docs

.. |json-schema| replace:: json-schema
.. _json-schema: https://json-schema.org/

.. |terraform-provider-scaffolding| replace:: terraform-provider-scaffolding
.. _terraform-provider-scaffolding: https://github.com/hashicorp/terraform-provider-scaffolding

.. |terraform-publishing-provider| replace:: Publishing Providers
.. _terraform-publishing-provider: https://www.terraform.io/docs/registry/providers/publishing.html

.. |go| replace:: Go
.. _go: https://golang.org/doc/install
