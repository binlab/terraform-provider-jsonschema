package provider

import (
	"fmt"

	"github.com/titanous/json5"
)

// ParseJSON5 parses JSON5 content and returns standard JSON data
func ParseJSON5(content []byte) (interface{}, error) {
	var result interface{}
	if err := json5.Unmarshal(content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON5: %w", err)
	}
	return result, nil
}

// ParseJSON5String parses JSON5 string content
func ParseJSON5String(content string) (interface{}, error) {
	return ParseJSON5([]byte(content))
}

// JSON5ToJSON converts JSON5 content to deterministic standard JSON bytes
func JSON5ToJSON(content []byte) ([]byte, error) {
	data, err := ParseJSON5(content)
	if err != nil {
		return nil, err
	}
	
	return MarshalDeterministic(data)
}

// JSON5StringToJSON converts JSON5 string to deterministic standard JSON bytes
func JSON5StringToJSON(content string) ([]byte, error) {
	return JSON5ToJSON([]byte(content))
}