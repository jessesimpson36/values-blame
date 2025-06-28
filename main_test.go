package main


import (
	"testing"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	util "helm.sh/helm/v4/pkg/chart/v2/util"
	"fmt"
)

func TestGetListOfKeys(t *testing.T) {
	v1 := map[string]interface{}{
		"key1": "value1",
		"key2": map[string]interface{}{
			"subkey1": "subvalue1",
			"subkey2": "subvalue2",
		},
		"key3": map[string]interface{}{
			"deeply": map[string]interface{}{
				"nested": "value3",
			},
		},
	}
	expectedKeys := []string{
		"key1",
		"key2",
		"key2.subkey1",
		"key2.subkey2",
		"key3",
		"key3.deeply",
		"key3.deeply.nested",
	}
	keys := getListOfKeys(v1)
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}
	for _, expectedKey := range expectedKeys {
		found := false
		for _, key := range keys {
			if key == expectedKey {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected key '%s' not found in keys", expectedKey)
		}
	}


}

func TestIsOptionWithinAppliedValues(t *testing.T) {
	v1 := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": map[string]interface{}{
			"deeply": map[string]interface{}{
				"nested": "value3",
			},
		},
	}

	if !isOptionWithinAppliedValues("key1", v1) {
		t.Errorf("Expected true for key1 in v1")
	}
	if !isOptionWithinAppliedValues("key2", v1) {
		t.Errorf("Expected true for key2 in v1")
	}
	if isOptionWithinAppliedValues("jesse", v1) {
		t.Errorf("Expected false for non-existent key in v1")
	}
	if isOptionWithinAppliedValues("key1.deeply.nested", v1) {
		t.Errorf("Expected false for deeply nested key in v1")
	}
	if !isOptionWithinAppliedValues("key3.deeply.nested", v1) {
		t.Errorf("Expected true for deeply nested key in v1")
	}
}

func TestTraversal(t *testing.T) {
	v1 := map[string]interface{}{
		"key1": "value1",
		"key2": map[string]interface{}{
			"subkey1": "subvalue1",
			"subkey2": "subvalue2",
		},
	}

	v2 := map[string]interface{}{
		"key3": "value3",
	}

	options := make(map[string]ValuesYAMLLine)

	c := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "moby",
			Version: "1.2.3",
		},
		Templates: []*chart.File{},
		Values: v1,
	}

	valuesCoalesced, err := util.CoalesceValues(c, v1)
	if err != nil {
		t.Fatalf("Error coalescing values: %v", err)
	}
	traverseValues(valuesCoalesced, v1, "test-1.yaml", &options)
	valuesCoalesced, err = util.CoalesceValues(c, v2)
	if err != nil {
		t.Fatalf("Error coalescing values: %v", err)
	}
	traverseValues(valuesCoalesced, v2, "test-2.yaml", &options)

	fmt.Printf("Options: %+v\n", options)
	if len(options) != 5 {
		t.Errorf("Expected 5 option, got %d", len(options))
	}
	if options["key1"].Option != "key1" || options["key1"].Line != "key1: value1" {
		t.Errorf("Unexpected option or line for key1: %v", options["key1"])
	}
}

