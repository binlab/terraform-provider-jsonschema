# Pre-commit Integration Examples

This directory contains examples of how to use `jsonschema-validator` with pre-commit.

## Prerequisites

Install the CLI tool first:

```bash
go install github.com/binlab/terraform-provider-jsonschema/cmd/jsonschema-validator@latest
```

Verify installation:

```bash
jsonschema-validator --version
```

## Basic Usage

1. Create a `.pre-commit-config.yaml` file in your project root:

```yaml
repos:
  - repo: https://github.com/binlab/terraform-provider-jsonschema
    rev: v0.1.0 # Use latest release
    hooks:
      - id: jsonschema-validator
        args:
          - --schema
          - schemas/my-schema.json
        files: ^data/.*\.json$
```

2. Install pre-commit hooks:

```bash
pre-commit install
```

3. Run manually:

```bash
# Run on all files
pre-commit run jsonschema-validator --hook-stage manual --all-files

# Run on specific files
pre-commit run jsonschema-validator --hook-stage manual --files data/example.json

# Run on staged files (if configured for commit stage)
pre-commit run
```

## Configuration Options

### Using Config File

If you have a `.jsonschema-validator.yaml` or `pyproject.toml` in your project:

```yaml
repos:
  - repo: https://github.com/binlab/terraform-provider-jsonschema
    rev: v0.1.0
    hooks:
      - id: jsonschema-validator
        files: ^data/.*\.json$
        # No args - reads from config file
```

### Multiple Schemas

Validate different file patterns with different schemas:

```yaml
repos:
  - repo: https://github.com/binlab/terraform-provider-jsonschema
    rev: v0.1.0
    hooks:
      - id: jsonschema-validator
        name: Validate API requests
        args: [--schema, schemas/api-request.schema.json]
        files: ^data/api-requests/.*\.json$

      - id: jsonschema-validator
        name: Validate user data
        args: [--schema, schemas/user.schema.json]
        files: ^data/users/.*\.json$
```

### Custom Environment Prefix

Use a custom environment variable prefix:

```yaml
repos:
  - repo: https://github.com/binlab/terraform-provider-jsonschema
    rev: v0.1.0
    hooks:
      - id: jsonschema-validator
        args:
          - --env-prefix
          - MY_APP_
          - --schema
          - schemas/config.schema.json
        files: ^config/.*\.json$
```

Then you can set `MY_APP_SCHEMA`, `MY_APP_VERBOSE`, etc. in your environment.

### Run on Git Commit

By default, the hook uses `stages: [manual]` to avoid interfering with normal commits. To run on commit:

```yaml
repos:
  - repo: https://github.com/binlab/terraform-provider-jsonschema
    rev: v0.1.0
    hooks:
      - id: jsonschema-validator
        args: [--schema, schemas/my-schema.json]
        files: ^data/.*\.json$
        stages: [commit] # Run on every commit
```

## Troubleshooting

### "Executable `jsonschema-validator` not found"

The CLI tool is not installed or not in PATH. Install it:

```bash
go install github.com/binlab/terraform-provider-jsonschema/cmd/jsonschema-validator@latest
```

### Hook Not Running

Check that:

1. The hook ID matches (use `jsonschema-validator` not `jsonschema-validator-env`)
2. The stage is correct (use `--hook-stage manual` or configure `stages: [commit]`)
3. File patterns match your target files

### Schema Not Found

Use relative paths from your project root:

```yaml
args: [--schema, schemas/my-schema.json] # ✓ Correct
# Not: args: [--schema, ./schemas/my-schema.json]  # ✗ May fail
```

### Verbose Output

Add `--verbose` for debugging:

```yaml
hooks:
  - id: jsonschema-validator
    args:
      - --schema
      - schemas/my-schema.json
      - --verbose
    files: ^data/.*\.json$
```

## Testing Locally

Test without committing:

```bash
# Test on all matching files
pre-commit run jsonschema-validator --hook-stage manual --all-files

# Test on specific files
pre-commit run jsonschema-validator --hook-stage manual --files data/test.json data/test2.json

# Test with verbose output
pre-commit run jsonschema-validator --hook-stage manual --all-files --verbose
```

## Example Project Structure

```
my-project/
├── .pre-commit-config.yaml
├── .jsonschema-validator.yaml  # Optional: default config
├── schemas/
│   ├── api-request.schema.json
│   ├── user.schema.json
│   └── product.schema.json
└── data/
    ├── api-requests/
    │   ├── create-user.json
    │   └── update-product.json
    ├── users/
    │   ├── alice.json
    │   └── bob.json
    └── products/
        ├── widget.json
        └── gadget.json
```

See `.pre-commit-config.yaml` in this directory for a complete working example.
