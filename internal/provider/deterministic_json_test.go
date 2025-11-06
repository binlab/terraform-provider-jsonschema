package provider

import (
	"testing"
)

func TestMarshalDeterministic(t *testing.T) {
	// Test that deterministic marshaling produces consistent output
	testData := map[string]interface{}{
		"zebra": "last",
		"alpha": "first",
		"nested": map[string]interface{}{
			"charlie": 3,
			"bravo":   2,
			"alpha":   1,
		},
	}

	// Marshal multiple times to ensure deterministic output
	result1, err1 := MarshalDeterministic(testData)
	if err1 != nil {
		t.Fatalf("First marshal failed: %v", err1)
	}

	result2, err2 := MarshalDeterministic(testData)
	if err2 != nil {
		t.Fatalf("Second marshal failed: %v", err2)
	}

	if string(result1) != string(result2) {
		t.Errorf("Non-deterministic results:\nFirst:  %s\nSecond: %s", result1, result2)
	}

	// Verify keys are sorted alphabetically
	expected := `{"alpha":"first","nested":{"alpha":1,"bravo":2,"charlie":3},"zebra":"last"}`
	if string(result1) != expected {
		t.Errorf("Keys not sorted correctly:\nExpected: %s\nGot: %s", expected, string(result1))
	}
}