package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	flag "github.com/spf13/pflag"
)

// Custom unmarshaler for handling both camelCase (JSON) and snake_case (YAML/TOML)
var unmarshalConf = koanf.UnmarshalConf{
	Tag: "koanf",
	// Use flatMap to handle nested structures
	FlatPaths: false,
}

// Loader handles configuration loading from multiple sources
type Loader struct {
	k *koanf.Koanf
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{
		k: koanf.New("."),
	}
}

// Load loads configuration from all available sources in priority order:
// 1. Command-line flags (highest priority)
// 2. Environment variables (JSONSCHEMA_VALIDATOR_*)
// 3. .jsonschema-validator.yaml in current directory
// 4. pyproject.toml section [tool.jsonschema-validator]
// 5. package.json field "jsonschema-validator"
// 6. ~/.jsonschema-validator.yaml in user home
// 7. Default values (lowest priority)
func (l *Loader) Load(flags *flag.FlagSet) (*Config, error) {
	// 1. Load defaults (lowest priority)
	if err := l.loadDefaults(); err != nil {
		return nil, fmt.Errorf("loading defaults: %w", err)
	}

	// 2. Load from user home config (if exists)
	if err := l.loadUserConfig(); err != nil {
		// User config is optional, don't fail if not found
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("loading user config: %w", err)
		}
	}

	// 3. Load from package.json (if exists)
	if err := l.loadPackageJSON(); err != nil {
		// Optional, don't fail if not found
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("loading package.json config: %w", err)
		}
	}

	// 4. Load from pyproject.toml (if exists)
	if err := l.loadPyprojectTOML(); err != nil {
		// Optional, don't fail if not found
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("loading pyproject.toml config: %w", err)
		}
	}

	// 5. Load from .jsonschema-validator.yaml (if exists)
	if err := l.loadProjectConfig(); err != nil {
		// Optional, don't fail if not found
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("loading project config: %w", err)
		}
	}

	// 6. Load from environment variables
	if err := l.loadEnvVars(); err != nil {
		return nil, fmt.Errorf("loading environment variables: %w", err)
	}

	// 7. Load from command-line flags (highest priority)
	if flags != nil {
		if err := l.loadFlags(flags); err != nil {
			return nil, fmt.Errorf("loading flags: %w", err)
		}
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := l.k.UnmarshalWithConf("", &cfg, unmarshalConf); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &cfg, nil
}

// LoadFromFile loads configuration from a specific file
func (l *Loader) LoadFromFile(path string) (*Config, error) {
	// Load defaults first
	if err := l.loadDefaults(); err != nil {
		return nil, fmt.Errorf("loading defaults: %w", err)
	}

	// Determine parser based on file extension
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".yaml", ".yml":
		// YAML uses snake_case (same as Terraform)
		if err := l.k.Load(file.Provider(path), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("loading config file %q: %w", path, err)
		}
	case ".toml":
		// TOML uses snake_case (same as Terraform)
		if err := l.k.Load(file.Provider(path), toml.Parser()); err != nil {
			return nil, fmt.Errorf("loading config file %q: %w", path, err)
		}
	case ".json":
		// JSON uses camelCase (package.json convention)
		// Need to normalize keys to snake_case
		tempK := koanf.New(".")
		if err := tempK.Load(file.Provider(path), json.Parser()); err != nil {
			return nil, fmt.Errorf("loading config file %q: %w", path, err)
		}
		// Convert camelCase to snake_case
		normalized := normalizeKeys(tempK.Raw())
		if err := l.k.Load(confmap.Provider(normalized, "."), nil); err != nil {
			return nil, fmt.Errorf("loading normalized config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := l.k.UnmarshalWithConf("", &cfg, unmarshalConf); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &cfg, nil
}

// loadDefaults loads default configuration values
func (l *Loader) loadDefaults() error {
	defaults := map[string]interface{}{
		"schema_version": "", // Empty means use schema's $schema field
		"schemas":        []interface{}{},
		"error_template": "", // Empty means use default formatting
	}

	return l.k.Load(confmap.Provider(defaults, "."), nil)
}

// loadUserConfig loads configuration from ~/.jsonschema-validator.yaml
func (l *Loader) loadUserConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".jsonschema-validator.yaml")
	if _, err := os.Stat(configPath); err != nil {
		return err
	}

	return l.k.Load(file.Provider(configPath), yaml.Parser())
}

// loadProjectConfig loads configuration from .jsonschema-validator.yaml in current directory
func (l *Loader) loadProjectConfig() error {
	// Try multiple file names in order of preference
	candidates := []string{
		".jsonschema-validator.yaml",
		".jsonschema-validator.yml",
		".jsonschema.yaml",
		".jsonschema.yml",
		"jsonschema-validator.yaml",
		"jsonschema-validator.yml",
	}

	for _, name := range candidates {
		if _, err := os.Stat(name); err == nil {
			return l.k.Load(file.Provider(name), yaml.Parser())
		}
	}

	return os.ErrNotExist
}

// loadPyprojectTOML loads configuration from pyproject.toml [tool.jsonschema-validator]
func (l *Loader) loadPyprojectTOML() error {
	const configFile = "pyproject.toml"

	if _, err := os.Stat(configFile); err != nil {
		return err
	}

	// Load the entire TOML file
	tempK := koanf.New(".")
	if err := tempK.Load(file.Provider(configFile), toml.Parser()); err != nil {
		return err
	}

	// Extract only the [tool.jsonschema-validator] section
	toolConfig := tempK.Cut("tool.jsonschema-validator")
	if toolConfig == nil || toolConfig.Raw() == nil {
		// No jsonschema-validator section found
		return os.ErrNotExist
	}

	// Merge the tool section into main config
	return l.k.Merge(toolConfig)
}

// loadPackageJSON loads configuration from package.json "jsonschema-validator" field
func (l *Loader) loadPackageJSON() error {
	const configFile = "package.json"

	if _, err := os.Stat(configFile); err != nil {
		return err
	}

	// Load the entire JSON file
	tempK := koanf.New(".")
	if err := tempK.Load(file.Provider(configFile), json.Parser()); err != nil {
		return err
	}

	// Extract only the "jsonschema-validator" field
	jsConfig := tempK.Cut("jsonschema-validator")
	if jsConfig == nil || jsConfig.Raw() == nil {
		// No jsonschema-validator field found
		return os.ErrNotExist
	}

	// Convert camelCase keys to snake_case for consistency
	normalized := normalizeKeys(jsConfig.Raw())
	
	// Load normalized config
	return l.k.Load(confmap.Provider(normalized, "."), nil)
}

// normalizeKeys recursively converts camelCase keys to snake_case
func normalizeKeys(data interface{}) map[string]interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			snakeKey := camelToSnake(key)
			result[snakeKey] = normalizeValue(value)
		}
		return result
	default:
		// If not a map, return empty map
		return make(map[string]interface{})
	}
}

// normalizeValue recursively processes values
func normalizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			snakeKey := camelToSnake(key)
			result[snakeKey] = normalizeValue(val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = normalizeValue(val)
		}
		return result
	default:
		return v
	}
}

// camelToSnake converts camelCase to snake_case
func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// loadEnvVars loads configuration from environment variables
// Environment variables use SCREAMING_SNAKE_CASE (no camelCase!)
// Prefixed with JSONSCHEMA_VALIDATOR_
// Examples:
//   JSONSCHEMA_VALIDATOR_SCHEMA_VERSION=draft/2020-12
//   JSONSCHEMA_VALIDATOR_ERROR_TEMPLATE="..."
func (l *Loader) loadEnvVars() error {
	return l.k.Load(env.Provider("JSONSCHEMA_VALIDATOR_", ".", func(s string) string {
		// Convert JSONSCHEMA_VALIDATOR_SCHEMA_VERSION to schema_version
		// Just remove prefix and lowercase - underscores stay as-is
		s = strings.TrimPrefix(s, "JSONSCHEMA_VALIDATOR_")
		s = strings.ToLower(s)
		return s
	}), nil)
}

// loadFlags loads configuration from command-line flags
func (l *Loader) loadFlags(flags *flag.FlagSet) error {
	return l.k.Load(posflag.Provider(flags, ".", l.k), nil)
}

// ParseRefOverridesFromString parses ref_override from string format
// Format: "url1=path1,url2=path2" or repeated --ref-override flags
func ParseRefOverridesFromString(s string) map[string]string {
	overrides := make(map[string]string)

	if s == "" {
		return overrides
	}

	// Split by comma
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		// Split by equals sign
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			url := strings.TrimSpace(parts[0])
			path := strings.TrimSpace(parts[1])
			if url != "" && path != "" {
				overrides[url] = path
			}
		}
	}

	return overrides
}

// ParseRefOverridesFromSlice parses ref_override from slice format
// Each element is in format "url=path"
func ParseRefOverridesFromSlice(slice []string) map[string]string {
	overrides := make(map[string]string)

	for _, item := range slice {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) == 2 {
			url := strings.TrimSpace(parts[0])
			path := strings.TrimSpace(parts[1])
			if url != "" && path != "" {
				overrides[url] = path
			}
		}
	}

	return overrides
}

// GetKoanf returns the underlying koanf instance for advanced usage
func (l *Loader) GetKoanf() *koanf.Koanf {
	return l.k
}
