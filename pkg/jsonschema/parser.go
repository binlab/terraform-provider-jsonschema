package jsonschema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// FileType represents the format of a configuration file
type FileType string

const (
	FileTypeJSON  FileType = "json"
	FileTypeJSON5 FileType = "json5"
	FileTypeYAML  FileType = "yaml"
	FileTypeTOML  FileType = "toml"
	FileTypeAuto  FileType = "auto"
)

// ParseFile reads and parses a file based on its extension or forced type.
// Supports JSON, JSON5, YAML, and TOML formats.
func ParseFile(path string, forceType FileType) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	fileType := forceType
	if fileType == FileTypeAuto || fileType == "" {
		fileType = DetectFileType(path)
	}

	switch fileType {
	case FileTypeJSON:
		return ParseJSON(data)
	case FileTypeJSON5:
		return ParseJSON5(data)
	case FileTypeYAML:
		return ParseYAML(data)
	case FileTypeTOML:
		return ParseTOML(data)
	default:
		// Try JSON5 as fallback (most permissive)
		return ParseJSON5(data)
	}
}

// DetectFileType determines file type from extension.
// Returns FileTypeJSON5 as fallback for unknown extensions (most permissive).
func DetectFileType(path string) FileType {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".json":
		return FileTypeJSON
	case ".json5":
		return FileTypeJSON5
	case ".yaml", ".yml":
		return FileTypeYAML
	case ".toml":
		return FileTypeTOML
	default:
		return FileTypeJSON5 // Most permissive fallback
	}
}

// ParseJSON parses standard JSON data
func ParseJSON(data []byte) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	return result, nil
}

// ParseYAML parses YAML data
func ParseYAML(data []byte) (interface{}, error) {
	var result interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}
	return result, nil
}

// ParseTOML parses TOML data
func ParseTOML(data []byte) (interface{}, error) {
	var result interface{}
	if err := toml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing TOML: %w", err)
	}
	return result, nil
}
