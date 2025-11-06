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
- **Flexible Reference Resolution**: Configurable base URLs for resolving ``$ref`` URIs
- **Custom Error Templates**: Customize validation error messages with templating support
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
        version = "~> 0.3.1"
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
    # Default JSON Schema version (optional)
    # Supported: "draft-04", "draft-06", "draft-07", "draft/2019-09", "draft/2020-12"
    schema_version = "draft/2020-12"  # Default
    
    # Base URL for resolving $ref URIs (optional)
    base_url = "https://example.com/schemas/"
    
    # Default error message template (optional)
    error_message_template = "Validation failed: {error} in {schema}"
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

  # Remote schema with per-resource base URL
  data "jsonschema_validator" "remote_validation" {
    document = file("data.json")
    schema   = "api/v1/schema.json"
    base_url = "https://schemas.example.com/"  # Base URL for this validation
  }

  # Or use provider-level base URL as fallback
  provider "jsonschema" {
    base_url = "https://default-schemas.example.com/"
  }

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
