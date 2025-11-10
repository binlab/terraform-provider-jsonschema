# Phase 4 Completion: Pre-commit Hook Integration

## Summary

Phase 4 has been successfully completed. The `jsonschema-validator` CLI tool now has full pre-commit integration support with comprehensive documentation and tested workflows.

## What Was Implemented

### 1. Pre-commit Hook Configuration (`.pre-commit-hooks.yaml`)

Two hooks are provided:

- **`jsonschema-validator`** (main hook)
  - For production use with explicit arguments
  - Users specify schema path and file patterns in their config
  - Supports all CLI flags
  
- **`jsonschema-validator-env`** (testing/convenience hook)
  - Reads configuration from environment variables
  - Useful for CI/CD pipelines
  - Still requires explicit environment variable setup

**Key Design Decision:** Using `language: system` instead of `language: golang`

**Rationale:**
- The repository is a monorepo (Terraform provider + CLI tool)
- The CLI binary is in `cmd/jsonschema-validator/` subdirectory
- Pre-commit's `language: golang` expects a single main package at repo root
- This is the standard approach used by Go CLI tools (golangci-lint, gofmt, etc.)
- Requires users to install once with `go install`, then pre-commit invokes the installed binary

### 2. Documentation

**`examples/pre-commit/README.md`** (comprehensive guide)
- Prerequisites and installation steps
- Basic usage examples
- Advanced configurations (multiple schemas, config files, custom env prefix)
- Troubleshooting common issues
- Testing instructions
- Example project structure

**`examples/pre-commit/.pre-commit-config.yaml`** (complete example)
- 4 example configurations showing different use cases
- Inline comments explaining each pattern
- Examples for API validation, user data, product catalogs
- Shows how to use config files vs. explicit args

**README.rst** (main documentation)
- Updated with reference to examples directory
- Existing pre-commit section enhanced with pointer to examples

**docs/RELEASING.md** (release process)
- Already includes pre-commit testing in post-release verification
- Step 6: Verify pre-commit integration works after publishing

### 3. Test Files and Validation

Created test infrastructure in `examples/pre-commit/`:
- `schema.json` - Simple person schema (name, age)
- `valid.json` - Valid test document
- `invalid.json` - Invalid test document (empty name, age > 150)
- `.pre-commit-config-test.yaml` - Local testing configuration

**Tested Scenarios:**
- ✅ Validation passes on valid JSON
- ✅ Validation fails on invalid JSON with clear error messages
- ✅ Multiple files validated in one run
- ✅ File pattern matching (includes/excludes)
- ✅ Exit codes correct (0=success, 1=validation fail)
- ✅ Error messages show all validation failures

**Test Results:**
```bash
# Valid file
Test JSON Schema Validation..............................................Passed

# Invalid file
Test JSON Schema Validation..............................................Failed
- hook id: jsonschema-validator-test
- exit code: 1

document "examples/pre-commit/invalid.json": jsonschema validation failed
- at '/age': maximum: got 200, want 150
- at '/name': minLength: got 0, want 1

# Multiple files
✓ examples/pre-commit/valid.json: valid
document "examples/pre-commit/invalid.json": jsonschema validation failed...
```

## Configuration Priority

The CLI tool supports 5 configuration sources (in order):
1. CLI flags (highest priority)
2. Environment variables (customizable prefix via `--env-prefix`)
3. `.jsonschema-validator.yaml`
4. `pyproject.toml` (`[tool.jsonschema-validator]` section)
5. Defaults (lowest priority)

Pre-commit users can leverage any of these sources.

## Common Usage Patterns

### Pattern 1: Explicit Schema (Simplest)
```yaml
hooks:
  - id: jsonschema-validator
    args: [--schema, schemas/my-schema.json]
    files: ^data/.*\.json$
```

### Pattern 2: Config File (Python Projects)
```yaml
# pyproject.toml
[tool.jsonschema-validator]
schema = "schemas/my-schema.json"

# .pre-commit-config.yaml
hooks:
  - id: jsonschema-validator
    files: ^data/.*\.json$  # No args needed
```

### Pattern 3: Multiple Schemas
```yaml
hooks:
  - id: jsonschema-validator
    name: Validate API requests
    args: [--schema, schemas/api.json]
    files: ^api/requests/.*\.json$
    
  - id: jsonschema-validator
    name: Validate configs
    args: [--schema, schemas/config.json]
    files: ^config/.*\.json$
```

### Pattern 4: Custom Environment Prefix
```yaml
hooks:
  - id: jsonschema-validator
    args:
      - --env-prefix
      - MY_APP_
      - --schema
      - schemas/my-schema.json
    files: ^data/.*\.json$
```

## Installation Flow

**For Hook Users:**
1. Install CLI tool: `go install github.com/iilei/terraform-provider-jsonschema/cmd/jsonschema-validator@latest`
2. Verify: `jsonschema-validator --version`
3. Add to `.pre-commit-config.yaml`
4. Install hooks: `pre-commit install`
5. Run: `pre-commit run jsonschema-validator --hook-stage manual --all-files`

**Why Manual Stage by Default:**
- Avoids interfering with normal git workflow
- Users can opt-in to automatic validation by adding `stages: [commit]`
- Gives users control over when validation runs
- Consistent with other validation tools (mypy, pylint, etc.)

## Troubleshooting Guide

Common issues and solutions documented in `examples/pre-commit/README.md`:

1. **"Executable not found"** → Run `go install`
2. **Hook not running** → Check stage, file patterns, hook ID
3. **Schema not found** → Use relative paths from project root
4. **Need debugging** → Add `--verbose` to args

## Testing Workflow

Documented complete testing workflow:

```bash
# Test locally without committing
pre-commit run jsonschema-validator --hook-stage manual --all-files

# Test specific files
pre-commit run jsonschema-validator --hook-stage manual --files data/test.json

# Test with verbose output
pre-commit run jsonschema-validator --hook-stage manual --all-files --verbose

# Test with try-repo (during development)
pre-commit try-repo . jsonschema-validator --files examples/pre-commit/valid.json
```

## What Changed from Initial Attempts

**Initial Approach:** `language: golang` with various `entry` and `additional_dependencies` configurations

**Problems Encountered:**
- "Executable not found" errors
- "no Go files in .../cmd/jsonschema-validator" errors
- "no required module provides package" errors
- Pre-commit's golang support doesn't handle monorepo structure well

**Final Solution:** `language: system`
- Requires manual installation but works reliably
- Matches pattern used by all major Go CLI tools (golangci-lint, gofmt, goimports, etc.)
- Better user experience (install once, use everywhere)
- More flexible (users can pin versions independently)

## Files Added/Modified

**Added:**
- `examples/pre-commit/README.md` - Comprehensive guide (200+ lines)
- `examples/pre-commit/.pre-commit-config.yaml` - Complete example (60 lines)
- `examples/pre-commit/schema.json` - Test schema
- `examples/pre-commit/valid.json` - Valid test data
- `examples/pre-commit/invalid.json` - Invalid test data
- `.pre-commit-config-test.yaml` - Local testing config

**Modified:**
- `.pre-commit-hooks.yaml` - Added two hooks with `language: system`
- `README.rst` - Added reference to examples directory

## Next Steps (Phase 5)

Phase 4 is complete. Ready to proceed with Phase 5: Build & Distribution

**Phase 5 Tasks:**
1. ✅ Goreleaser configuration (already complete)
2. ✅ Release documentation (already complete in `docs/RELEASING.md`)
3. ☐ Test local build: `goreleaser build --snapshot --clean`
4. ☐ Verify both binaries build correctly (provider + CLI)
5. ☐ Test CLI installation from built artifacts
6. ☐ Test end-to-end release flow (tag → build → publish)
7. ☐ Update CI/CD workflows if needed
8. ☐ Document any CI/CD changes

## Metrics

- **Documentation:** 250+ lines of examples and guides
- **Test Coverage:** All major scenarios tested (valid/invalid/multiple files)
- **Configuration Options:** 5 sources supported
- **Example Configs:** 4 different usage patterns documented
- **Troubleshooting:** 4 common issues documented with solutions

## Success Criteria Met

✅ Pre-commit hooks defined and tested  
✅ Documentation comprehensive and clear  
✅ Installation process documented  
✅ Testing workflow documented  
✅ Troubleshooting guide provided  
✅ Multiple usage patterns demonstrated  
✅ Files validated correctly (pass/fail with proper exit codes)  
✅ Error messages clear and actionable  

## Phase 4 Status: ✅ COMPLETE
