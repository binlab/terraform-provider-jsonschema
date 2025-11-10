package jsonschema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFileType(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected FileType
	}{
		{"JSON file", "config.json", FileTypeJSON},
		{"JSON5 file", "config.json5", FileTypeJSON5},
		{"YAML file", "config.yaml", FileTypeYAML},
		{"YML file", "config.yml", FileTypeYAML},
		{"TOML file", "config.toml", FileTypeTOML},
		{"Uppercase extension", "CONFIG.JSON", FileTypeJSON},
		{"Mixed case YAML", "Config.YaML", FileTypeYAML},
		{"No extension", "config", FileTypeJSON5},
		{"Unknown extension", "config.xml", FileTypeJSON5},
		{"Path with slashes", "/path/to/config.yaml", FileTypeYAML},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectFileType(tt.path)
			if result != tt.expected {
				t.Errorf("DetectFileType(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "valid JSON object",
			input:   []byte(`{"name": "test", "count": 42}`),
			wantErr: false,
		},
		{
			name:    "valid JSON array",
			input:   []byte(`[1, 2, 3]`),
			wantErr: false,
		},
		{
			name:    "valid JSON string",
			input:   []byte(`"hello"`),
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   []byte(`{invalid}`),
			wantErr: true,
		},
		{
			name:    "JSON with trailing comma (should fail)",
			input:   []byte(`{"name": "test",}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("ParseJSON() returned nil result without error")
			}
		})
	}
}

func TestParseYAML(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name: "valid YAML object",
			input: []byte(`
name: test
count: 42
enabled: true
`),
			wantErr: false,
		},
		{
			name: "valid YAML array",
			input: []byte(`
- item1
- item2
- item3
`),
			wantErr: false,
		},
		{
			name:    "valid YAML string",
			input:   []byte(`hello world`),
			wantErr: false,
		},
		{
			name: "YAML with comments",
			input: []byte(`
# This is a comment
name: test  # inline comment
count: 42
`),
			wantErr: false,
		},
		{
			name: "invalid YAML",
			input: []byte(`
{{{invalid yaml
`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseYAML(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("ParseYAML() returned nil result without error")
			}
		})
	}
}

func TestParseTOML(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name: "valid TOML",
			input: []byte(`
name = "test"
count = 42
enabled = true
`),
			wantErr: false,
		},
		{
			name: "TOML with section",
			input: []byte(`
[section]
name = "test"
count = 42
`),
			wantErr: false,
		},
		{
			name: "TOML with array",
			input: []byte(`
items = [1, 2, 3]
`),
			wantErr: false,
		},
		{
			name: "invalid TOML",
			input: []byte(`
name = test  # missing quotes
`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTOML(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTOML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("ParseTOML() returned nil result without error")
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"test.json":  `{"name": "json-test", "count": 42}`,
		"test.json5": `{name: "json5-test", count: 42, /* comment */ }`,
		"test.yaml": `
name: yaml-test
count: 42
`,
		"test.toml": `
name = "toml-test"
count = 42
`,
		"no-ext": `{"name": "no-extension"}`,
	}

	for filename, content := range files {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	tests := []struct {
		name      string
		filename  string
		forceType FileType
		wantErr   bool
	}{
		{"JSON auto-detect", "test.json", FileTypeAuto, false},
		{"JSON5 auto-detect", "test.json5", FileTypeAuto, false},
		{"YAML auto-detect", "test.yaml", FileTypeAuto, false},
		{"TOML auto-detect", "test.toml", FileTypeAuto, false},
		{"No extension with auto (JSON5 fallback)", "no-ext", FileTypeAuto, false},
		{"Force JSON on YAML file", "test.yaml", FileTypeJSON, true}, // YAML is not valid JSON
		{"Force JSON5 on no extension", "no-ext", FileTypeJSON5, false},
		{"Non-existent file", "missing.json", FileTypeAuto, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.filename)
			result, err := ParseFile(path, tt.forceType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile(%q, %v) error = %v, wantErr %v", tt.filename, tt.forceType, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("ParseFile() returned nil result without error")
			}
		})
	}
}

func TestParseFile_ContentValidation(t *testing.T) {
	tmpDir := t.TempDir()

	// Test that each format correctly parses its content
	tests := []struct {
		name     string
		filename string
		content  string
		fileType FileType
		validate func(t *testing.T, result interface{})
	}{
		{
			name:     "JSON object parsing",
			filename: "test.json",
			content:  `{"name": "test", "count": 42}`,
			fileType: FileTypeAuto,
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("Expected map[string]interface{}")
				}
				if m["name"] != "test" {
					t.Errorf("Expected name=test, got %v", m["name"])
				}
				if m["count"].(float64) != 42 {
					t.Errorf("Expected count=42, got %v", m["count"])
				}
			},
		},
		{
			name:     "YAML object parsing",
			filename: "test.yaml",
			content:  "name: test\ncount: 42\n",
			fileType: FileTypeAuto,
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("Expected map[string]interface{}")
				}
				if m["name"] != "test" {
					t.Errorf("Expected name=test, got %v", m["name"])
				}
				if m["count"].(int) != 42 {
					t.Errorf("Expected count=42, got %v", m["count"])
				}
			},
		},
		{
			name:     "TOML object parsing",
			filename: "test.toml",
			content:  "name = \"test\"\ncount = 42\n",
			fileType: FileTypeAuto,
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("Expected map[string]interface{}")
				}
				if m["name"] != "test" {
					t.Errorf("Expected name=test, got %v", m["name"])
				}
				if m["count"].(int64) != 42 {
					t.Errorf("Expected count=42, got %v", m["count"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.filename)
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result, err := ParseFile(path, tt.fileType)
			if err != nil {
				t.Fatalf("ParseFile() error = %v", err)
			}

			tt.validate(t, result)
		})
	}
}
