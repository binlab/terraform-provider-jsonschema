=============================
terraform-provider-jsonschema
=============================

|upstream_sync| |fork_releaser| |run_tests| |latest_release| |coverage_status|

.. |upstream_sync| image:: https://github.com/binlab/terraform-provider-jsonschema/actions/workflows/upstream-sync.yml/badge.svg
   :target: https://github.com/binlab/terraform-provider-jsonschema/actions/workflows/upstream-sync.yml
   :alt: Upstream Sync Status

.. |fork_releaser| image:: https://github.com/binlab/terraform-provider-jsonschema/actions/workflows/fork-releaser.yml/badge.svg
   :target: https://github.com/binlab/terraform-provider-jsonschema/actions/workflows/fork-releaser.yml
   :alt: Fork Releaser Status

.. |run_tests| image:: https://github.com/binlab/terraform-provider-jsonschema/actions/workflows/test.yml/badge.svg
   :target: https://github.com/binlab/terraform-provider-jsonschema/actions/workflows/test.yml
   :alt: Build and Test Status

.. |latest_release| image:: https://img.shields.io/github/v/release/binlab/terraform-provider-jsonschema?style=flat&color=31c653
   :target: https://github.com/binlab/terraform-provider-jsonschema/releases/latest
   :alt: GitHub Latest Release

ℹ️ **NOTE**

  This is a maintained fork of the `iilei/terraform-provider-jsonschema <https://github.com/iilei/terraform-provider-jsonschema>`_ provider.
  Originally developed by `vladarts <https://github.com/vladarts/terraform-provider-jsonschema>`_ and previously forked/maintained by `JeffAshton <https://github.com/JeffAshton/terraform-provider-jsonschema>`_ and `nekottyo <https://github.com/nekottyo/terraform-provider-jsonschema>`_.

  This fork is actively maintained to provide critical enhancements and features that are currently missing or unaddressed in the upstream repositories.

  **Key Enhancements:**

  * **In-line Content Support:** Re-introduced ``document_content`` and ``schema_content`` arguments (see `PR #29 <https://github.com/iilei/terraform-provider-jsonschema/pull/29>`_ or the `feature branch <https://github.com/binlab/terraform-provider-jsonschema/tree/feat/add-inline-document-and-schema>`_). These allow providing JSON document and schema content directly as strings within HCL, which is essential for dynamically generated data or minimizing external file dependencies.

.. |coverage_status| image:: https://codecov.io/github/binlab/terraform-provider-jsonschema/branch/release/graph/badge.svg
    :target: https://codecov.io/github/binlab/terraform-provider-jsonschema
    :alt: Coverage Status

A |terraform|_ provider for validating JSON, JSON5, YAML, and TOML documents using |json-schema|_ specifications.

.. note::
   📌 **CLI Tool Development Status**

   The standalone CLI tool (``jsonschema-validator``) and pre-commit hook features are **not actively developed** in this repository. The primary focus is the Terraform provider.

.. warning::
   ⚠️ **Version 0.x Development - Breaking Changes Expected**

   This provider is in initial development (0.x.x). Per `semantic versioning <https://semver.org/#spec-item-4>`_, **breaking changes may occur in ANY release** (minor or patch) until version 1.0.0.

   **Required actions:**

   - **Always pin to a specific version** in production
   - **Review release notes** before upgrading
   - **Test upgrades** in non-production environments first

   Stability and standard semver guarantees begin at version 1.0.0.


.. note::
   **Breaking Changes in v0.6.1**

   Version 0.6.1 introduces **breaking API changes**:

   - **document field**: Now expects a file path instead of content (remove ``file()`` wrapper)
   - **valid_json field**: Renamed from ``validated`` for clarity
   - **force_filetype field**: New optional field to override format detection
   - **Multi-format support**: Added YAML and TOML validation

   **Migration required:**

   .. code-block:: diff

     # Before (v0.5.x)
     data "jsonschema_validator" "config" {
     -  document = file("${path.module}/config.json")
     +  document = "${path.module}/config.json"
        schema   = "${path.module}/config.schema.json"
     }

     locals {
     -  config = jsondecode(data.jsonschema_validator.config.validated)
     +  config = jsondecode(data.jsonschema_validator.config.valid_json)
     }

   See the full documentation for migration details.


Features
========

- **Multi-format Support**: Validate JSON, JSON5, YAML, and TOML documents against JSON Schema
- **Auto-detection**: Format determined from file extension (``.json``, ``.json5``, ``.yaml``, ``.yml``, ``.toml``)
- **JSON5 Support**: Full support for JSON5 schemas with comments, trailing commas, unquoted keys
- **Schema Versions**: Draft 4, 6, 7, 2019-09, and 2020-12 support
- **External References**: Resolves ``$ref`` URIs including JSON5 files
- **Reference Overrides**: Redirect remote ``$ref`` URLs to local files for offline validation
- **Enhanced Error Templating**: Flexible error formatting with Go templates
- **Deterministic Output**: Consistent JSON for stable Terraform state

.. note::
   🔒 **Security Requirement**: All commits to this repository must be GPG signed. Pull requests with unsigned commits will be rejected by CI.

See the `full documentation <docs/index.md>`_ for detailed usage, advanced features, and examples. **Hands-on examples** demonstrating schema traversal, error templating, and reference overrides are available in the `examples/ <examples/>`_ directory.

Installation
============

Terraform Provider
------------------

On |terraform|_ versions 0.13+ use:

.. code-block:: terraform

  terraform {
    required_providers {
      jsonschema = {
        source  = "binlab/jsonschema"
        version = "0.6.2"  # Pin to specific version
      }
    }
  }

For |terraform|_ versions 0.12 or lower, see |terraform-install-plugin|_.

Standalone CLI Tool
-------------------

Install the ``jsonschema-validator`` CLI for use outside Terraform (Python, Node.js, or any project):

**Via Go:**

.. code-block:: bash

  go install github.com/binlab/terraform-provider-jsonschema/cmd/jsonschema-validator@latest

**Via Release Binary:**

Download pre-built binaries from `GitHub Releases <https://github.com/binlab/terraform-provider-jsonschema/releases>`_


Quick Start
===========

Terraform Provider
------------------

.. code-block:: terraform

  provider "jsonschema" {
    schema_version = "draft/2020-12"  # Optional
  }

  # JSON validation
  data "jsonschema_validator" "config" {
    document = "${path.module}/config.json"
    schema   = "${path.module}/config.schema.json"
  }

  # YAML validation (auto-detected from .yaml extension)
  data "jsonschema_validator" "k8s_manifest" {
    document = "${path.module}/deployment.yaml"
    schema   = "${path.module}/k8s-schema.json"
  }

  # TOML validation (auto-detected from .toml extension)
  data "jsonschema_validator" "app_config" {
    document = "${path.module}/config.toml"
    schema   = "${path.module}/config-schema.json"
  }

  # Use the validated document
  locals {
    config = jsondecode(data.jsonschema_validator.config.valid_json)
  }

Inline Content Validation
---------------------------------

Validate documents and schemas provided directly as strings in your HCL code. This is useful for dynamically generated content or when you want to avoid managing separate files.

.. code-block:: terraform

  locals {
    # Example Terraform object to validate
    user_data = {
      name  = "John Doe"
      email = "john.doe@example.com"
      age   = 30
    }
  }

  data "jsonschema_validator" "inline_user_validation" {
    # Document content from a Terraform local variable, encoded as JSON
    document_content = jsonencode(local.user_data)

    # Schema content defined inline using a heredoc
    schema_content = <<-EOT
      {
        "type": "object",
        "properties": {
          "name": {"type": "string"},
          "email": {"type": "string", "format": "email"},
          "age": {"type": "integer", "minimum": 0}
        },
        "required": ["name", "email"]
      }
    EOT

    # Optionally force content type if auto-detection is not sufficient
    force_filetype = "json"
  }

  output "validated_user" {
    value = jsondecode(data.jsonschema_validator.inline_user_validation.valid_json)
  }

Standalone CLI
--------------

.. code-block:: bash

  # Validate JSON file
  jsonschema-validator --schema config.schema.json config.json

  # Validate YAML file (auto-detected)
  jsonschema-validator --schema k8s.schema.json deployment.yaml

  # Validate TOML file (auto-detected)
  jsonschema-validator --schema app.schema.json config.toml

  # Force file type override
  jsonschema-validator --schema api.schema.json --force-filetype yaml data.txt

  # With JSON5 support
  jsonschema-validator --schema app.schema.json5 app.json5

  # Use configuration file (zero-config mode)
  jsonschema-validator  # Reads .jsonschema-validator.yaml

Documentation
=============

- **Provider Documentation**: `docs/index.md <docs/index.md>`_
- **Data Source Reference**: `docs/data-sources/jsonschema_validator.md <docs/data-sources/jsonschema_validator.md>`_
- **Examples**: `examples/ <examples/>`_ directory
- **Registry Documentation**: |user-docs|_

Key Features
============

Multi-format Support
--------------------

Validate documents in multiple formats against JSON Schema:

.. code-block:: terraform

  # YAML document validation
  data "jsonschema_validator" "k8s_deployment" {
    document = "${path.module}/deployment.yaml"  # Auto-detected from extension
    schema   = "${path.module}/k8s-schema.json"
  }

  # TOML configuration validation
  data "jsonschema_validator" "app_config" {
    document = "${path.module}/config.toml"  # Auto-detected from extension
    schema   = "${path.module}/config-schema.json"
  }

  # Force file type override
  data "jsonschema_validator" "custom" {
    document       = "${path.module}/data.txt"  # File without standard extension
    schema         = "${path.module}/schema.json"
    force_filetype = "yaml"  # Override auto-detection
  }

JSON5 Support
-------------

.. code-block:: terraform

  # Create a JSON5 document file
  resource "local_file" "json5_config" {
    filename = "${path.module}/config.json5"
    content  = <<-EOT
      {
        // JSON5 comments supported
        "ports": [8080, 8081,], // Trailing commas
        config: { enabled: true } // Unquoted keys
      }
    EOT
  }

  data "jsonschema_validator" "json5_config" {
    document = local_file.json5_config.filename
    schema   = "${path.module}/service.schema.json5"
  }

Reference Overrides
-------------------

Redirect remote ``$ref`` URLs to local files for offline validation:

.. code-block:: terraform

  data "jsonschema_validator" "api_request" {
    document = "${path.module}/api-request.json"
    schema   = "${path.module}/schemas/api-request.schema.json"

    ref_overrides = {
      "https://api.example.com/schemas/user.schema.json" = "${path.module}/schemas/user.schema.json"
    }
  }

**Benefits:**

- **Offline validation** - No internet connection required
- **Air-gapped environments** - Works in restricted networks
- **Deterministic builds** - Same inputs = same results
- **No HTTP complexity** - No proxy settings, authentication, or TLS configuration needed

See `examples/ref_overrides/ <examples/ref_overrides/>`_ for a complete example.

Error Templating
----------------

Customize error output with Go templates:

.. code-block:: terraform

  data "jsonschema_validator" "config" {
    document = "${path.module}/config.yaml"
    schema   = "${path.module}/config.schema.json"
    error_message_template = "{{range .Errors}}{{.DocumentPath}}: {{.Message}}\n{{end}}"
  }

Available template variables: ``{{.FullMessage}}``, ``{{.ErrorCount}}``, ``{{.Errors}}``, ``{{.SchemaFile}}``, ``{{.Document}}``

See the `full documentation <docs/index.md>`_ for advanced templating examples.

Development
===========

Requirements: |go|_ 1.25+

.. code-block:: bash

  # Run tests
  go test ./internal/provider/ -v

  # Run acceptance tests
  TF_ACC=1 go test ./internal/provider/ -v

  # View coverage
  go test ./internal/provider/ -coverprofile=coverage.out
  go tool cover -html=coverage.out


.. |terraform| replace:: Terraform
.. _terraform: https://www.terraform.io/

.. |terraform-install-plugin| replace:: install a terraform plugin
.. _terraform-install-plugin: https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin

.. |user-docs| replace:: User Documentation
.. _user-docs: https://registry.terraform.io/providers/binlab/jsonschema/latest/docs

.. |json-schema| replace:: json-schema
.. _json-schema: https://json-schema.org/

.. |terraform-provider-scaffolding| replace:: terraform-provider-scaffolding
.. _terraform-provider-scaffolding: https://github.com/hashicorp/terraform-provider-scaffolding

.. |terraform-publishing-provider| replace:: Publishing Providers
.. _terraform-publishing-provider: https://www.terraform.io/docs/registry/providers/publishing.html

.. |go| replace:: Go
.. _go: https://golang.org/doc/install
