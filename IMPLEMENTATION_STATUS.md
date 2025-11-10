# CLI Tool Implementation Status

## 📚 README-Driven Development - COMPLETE

All documentation has been written first to define the specification:

### Documentation Created

- ✅ **Main README.rst** - Updated with CLI tool section and pre-commit features
- ✅ **cmd/jsonschema-validator/README.md** - Comprehensive CLI documentation
- ✅ **examples/cli/README.md** - Example configurations and usage
- ✅ **examples/cli/.jsonschema-validator.yaml** - Standalone config example
- ✅ **examples/cli/pyproject.toml.example** - Python project config
- ✅ **examples/cli/package.json.example** - Node.js project config
- ✅ **examples/cli/.pre-commit-config.yaml.example** - Pre-commit examples
- ✅ **.pre-commit-hooks.yaml** - Pre-commit hook definitions

## 🎯 Implementation Checklist

### Phase 1: Core Infrastructure

#### Package Structure
- ☐ Create `pkg/validator/` package
  - ☐ `validator.go` - Core validation logic
  - ☐ `validator_test.go` - Unit tests
  - ☐ `options.go` - Configuration options struct
  - ☐ `options_test.go` - Options tests
  - ☐ `errors.go` - Error formatting (extracted from provider)
  - ☐ `errors_test.go` - Error formatting tests

- ☐ Create `pkg/config/` package (koanf v2.3.0)
  - ☐ `config.go` - Config struct definitions
  - ☐ `loader.go` - Multi-source config loading
  - ☐ `loader_test.go` - Config loading tests
  - ☐ `yaml.go` - YAML config parser
  - ☐ `toml.go` - TOML (pyproject.toml) parser
  - ☐ `json.go` - JSON (package.json) parser

#### Extract Provider Logic
- ☐ Move JSON5 utilities to `pkg/json5/`
  - ☐ Extract from `internal/provider/json5_utils.go`
  - ☐ Keep provider wrapper for backward compatibility
  - ☐ Add tests

- ☐ Move deterministic JSON to `pkg/json/`
  - ☐ Extract from `internal/provider/deterministic_json.go`
  - ☐ Keep provider wrapper
  - ☐ Add tests

- ☐ Refactor provider to use `pkg/validator`
  - ☐ Update `data_source_jsonschema_validator.go`
  - ☐ Ensure all existing tests pass
  - ☐ Maintain backward compatibility

### Phase 2: Configuration System (koanf v2.3.0)

#### Configuration Discovery
- ☐ Implement `.jsonschema-validator.yaml` loader
- ☐ Implement `pyproject.toml` parser (`[tool.jsonschema-validator]`)
- ☐ Implement `package.json` parser (`"jsonschema-validator"` field)
- ☐ Implement environment variable support (`JSONSCHEMA_VALIDATOR_*`)
- ☐ Implement user home config (`~/.jsonschema-validator.yaml`)
- ☐ Implement config merging with priority

#### Configuration Options (1:1 with Terraform Provider)
- ☐ `schema_version` - Schema draft version
- ☐ `schemas[]` - Array of schema-document mappings
  - ☐ `path` - Schema file path
  - ☐ `documents[]` - Document file patterns (with glob support)
  - ☐ `ref_overrides{}` - Reference override map
- ☐ `error_template` - Custom error message template

#### Tests
- ☐ Test YAML config loading
- ☐ Test TOML config loading
- ☐ Test JSON config loading
- ☐ Test environment variable override
- ☐ Test configuration merging priority
- ☐ Test glob pattern expansion
- ☐ Test invalid config handling

### Phase 3: CLI Tool

#### Command-Line Interface
- ☐ Create `cmd/jsonschema-validator/main.go`
- ☐ Implement flag parsing (using `pflag`)
  - ☐ `--config, -c` - Config file path
  - ☐ `--schema, -s` - Schema file path
  - ☐ `--schema-version` - Schema draft version
  - ☐ `--ref-override` - Reference overrides (repeatable)
  - ☐ `--error-template` - Custom error template
  - ☐ `--format` - Output format (text, json)
  - ☐ `--quiet, -q` - Quiet mode
  - ☐ `--verbose, -v` - Verbose mode
  - ☐ `--version` - Version information
  - ☐ `--help, -h` - Help text

#### Core Features
- ☐ Configuration discovery and loading
- ☐ Single file validation
- ☐ Multiple file validation
- ☐ Glob pattern expansion for document paths
- ☐ Stdin support (`-` as filename)
- ☐ JSON5 document support
- ☐ JSON5 schema support
- ☐ Schema version selection
- ☐ Reference override support
- ☐ Custom error template rendering

#### Output Formats
- ☐ Text output (default, human-readable)
- ☐ JSON output (`--format json`)
- ☐ Colored output for TTY
- ☐ Plain output for non-TTY

#### Exit Codes
- ☐ Exit 0 - All validations passed
- ☐ Exit 1 - Validation errors found
- ☐ Exit 2 - Usage/configuration errors

#### Tests
- ☐ Test flag parsing
- ☐ Test config discovery
- ☐ Test single file validation
- ☐ Test multiple file validation
- ☐ Test glob pattern expansion
- ☐ Test stdin input
- ☐ Test JSON5 support
- ☐ Test reference overrides
- ☐ Test error templates
- ☐ Test output formats
- ☐ Test exit codes
- ☐ Integration tests with real schemas

### Phase 4: Pre-commit Integration

#### Pre-commit Hook Configuration
- ✅ Create `.pre-commit-hooks.yaml`
  - ✅ `jsonschema-validator` - Main hook
  - ✅ `jsonschema-validator-json5` - JSON5-specific hook

#### Pre-commit Features
- ☐ Pass filenames to validator
- ☐ Respect file filters from pre-commit
- ☐ Proper exit codes for pre-commit
- ☐ Colorized output in pre-commit

#### Tests
- ☐ Test pre-commit hook execution
- ☐ Test with `.jsonschema-validator.yaml`
- ☐ Test with `pyproject.toml`
- ☐ Test with `package.json`
- ☐ Test with inline args
- ☐ Test file filtering

### Phase 5: Build & Distribution

#### Build Configuration
- ☐ Update `.goreleaser.yml`
  - ☐ Add CLI binary build target
  - ☐ Multi-platform builds (linux, darwin, windows)
  - ☐ Multi-arch builds (amd64, arm64)
  - ☐ Archive generation
  - ☐ Checksum generation

#### Installation Methods
- ☐ `go install` support
- ☐ GitHub Releases with binaries
- ☐ Homebrew tap (future)
- ☐ Installation documentation

#### Tests
- ☐ Test builds on all platforms
- ☐ Test installation from release
- ☐ Test `go install` command

### Phase 6: Documentation & Examples

#### Examples
- ✅ Basic examples in `examples/cli/`
- ☐ Working example with actual schema files
- ☐ Pre-commit example repository
- ☐ GitHub Actions workflow example
- ☐ GitLab CI example

#### Documentation
- ✅ Main README update
- ✅ CLI-specific README
- ✅ Configuration examples
- ✅ Pre-commit examples
- ☐ Migration guide from other validators
- ☐ Troubleshooting guide

## 📋 Pre-commit Feature Checklist

From README.rst documentation:

```
☐ Configuration discovery from .jsonschema-validator.yaml
☐ Configuration discovery from pyproject.toml [tool.jsonschema-validator]
☐ Configuration discovery from package.json "jsonschema-validator"
☐ Environment variable support (JSONSCHEMA_VALIDATOR_*)
☐ Command-line flag parsing (matching Terraform provider options)
☐ JSON5 document validation
☐ JSON5 schema validation
☐ Schema version selection (draft 4/6/7/2019-09/2020-12)
☐ Reference override support (ref_overrides)
☐ Custom error message templates
☐ Batch file validation (multiple documents per schema)
☐ Glob pattern support for document paths
☐ Exit codes (0=success, 1=validation error, 2=usage error)
☐ Colored output for TTY
☐ JSON output format (--format json)
☐ Quiet mode (--quiet)
☐ Verbose mode (--verbose)
☐ Stdin support (validate from pipe)
☐ Pre-commit hooks.yaml configuration
☐ GitHub Actions integration example
☐ GitLab CI integration example
```

## 🎯 Next Steps

1. **Start with Phase 1** - Extract core validation logic
2. **Implement Phase 2** - Configuration system with koanf
3. **Build Phase 3** - CLI tool
4. **Test Phase 4** - Pre-commit integration
5. **Complete Phase 5** - Build and distribution
6. **Finalize Phase 6** - Documentation polish

## 📦 Dependencies to Add

```go
// go.mod additions
require (
    github.com/knadh/koanf/v2 v2.3.0
    github.com/knadh/koanf/parsers/yaml v0.1.0
    github.com/knadh/koanf/parsers/toml v0.1.0
    github.com/knadh/koanf/parsers/json v0.1.0
    github.com/knadh/koanf/providers/file v0.1.0
    github.com/knadh/koanf/providers/env v0.1.0
    github.com/knadh/koanf/providers/posflag v0.1.0
    github.com/spf13/pflag v1.0.5
    github.com/fatih/color v1.16.0  // For colored output
)
```

## 🧪 Testing Strategy

- **Unit Tests** - All packages at 90%+ coverage
- **Integration Tests** - End-to-end validation scenarios
- **CLI Tests** - Command-line interface testing
- **Pre-commit Tests** - Hook execution testing
- **Platform Tests** - Cross-platform compatibility

## 📊 Success Metrics

- [ ] All Terraform provider tests still pass
- [ ] CLI tool has 90%+ test coverage
- [ ] Pre-commit hook works in all documented scenarios
- [ ] Documentation is complete and accurate
- [ ] Builds successfully on all platforms
- [ ] Examples work out of the box
