package runner

import (
	"testing"
)

func TestTopoSort_Simple(t *testing.T) {
	// Test: A -> B -> C (simple linear order)
	names := []string{"a", "b", "c"}
	before := map[string][]string{
		"a": {"b", "c"},
		"b": {"c"},
	}
	after := map[string][]string{}

	result, err := TopoSort(names, before, after)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// a comes first, then b, then c
	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result))
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Fatalf("unexpected order: %v", result)
	}
}

func TestTopoSort_AfterConstraints(t *testing.T) {
	// Test: b after a, c after b
	names := []string{"a", "b", "c"}
	before := map[string][]string{}
	after := map[string][]string{
		"b": {"a"},
		"c": {"b"},
	}

	result, err := TopoSort(names, before, after)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result))
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Fatalf("unexpected order: %v", result)
	}
}

func TestTopoSort_Cycle(t *testing.T) {
	// Test cycle detection: a before b, b before a
	names := []string{"a", "b"}
	before := map[string][]string{
		"a": {"b"},
		"b": {"a"},
	}
	after := map[string][]string{}

	result, err := TopoSort(names, before, after)
	if err == nil {
		t.Fatal("expected error for cycle, got nil")
	}

	if result != nil {
		t.Fatalf("expected nil result on error, got %v", result)
	}

	if err.Error() == "" {
		t.Fatal("expected error message")
	}
}

func TestTopoSort_ComplexCycle(t *testing.T) {
	// Test complex cycle: a -> b -> c -> a
	names := []string{"a", "b", "c"}
	before := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"a"},
	}
	after := map[string][]string{}

	result, err := TopoSort(names, before, after)
	if err == nil {
		t.Fatal("expected error for cycle, got nil")
	}

	if result != nil {
		t.Fatalf("expected nil result on error, got %v", result)
	}
}

func TestTopoSort_NoConstraints(t *testing.T) {
	// Test with no constraints - any order should work
	names := []string{"a", "b", "c"}
	before := map[string][]string{}
	after := map[string][]string{}

	result, err := TopoSort(names, before, after)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result))
	}

	// Check that all items are present
	seen := make(map[string]bool)
	for _, item := range result {
		seen[item] = true
	}
	for _, name := range names {
		if !seen[name] {
			t.Fatalf("missing item: %s", name)
		}
	}
}

func TestTopoSort_MultipleConstraints(t *testing.T) {
	// Test: preprocessor order with multiple dependencies
	// index should run first
	// custom-a before custom-b
	// custom-c can run anytime
	names := []string{"index", "custom-a", "custom-b", "custom-c"}
	before := map[string][]string{
		"custom-a": {"custom-b"},
	}
	after := map[string][]string{}

	result, err := TopoSort(names, before, after)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 4 {
		t.Fatalf("expected 4 items, got %d", len(result))
	}

	// Find positions
	posA := -1
	posB := -1
	for i, name := range result {
		if name == "custom-a" {
			posA = i
		}
		if name == "custom-b" {
			posB = i
		}
	}

	if posA >= posB {
		t.Fatalf("custom-a should come before custom-b, got positions %d and %d", posA, posB)
	}
}

func TestResolvePreprocessorOrder(t *testing.T) {
	// Test preprocessor order resolution with defaults and customs
	defaults := []string{"index"}
	customs := map[string]struct {
		Before []string
		After  []string
	}{
		"custom-a": {
			Before: []string{"custom-b"},
		},
		"custom-b": {
			After: []string{"index"},
		},
	}

	result, err := ResolvePreprocessorOrder(defaults, customs, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d: %v", len(result), result)
	}

	// Verify order constraints
	posIndex := -1
	posA := -1
	posB := -1
	for i, name := range result {
		if name == "index" {
			posIndex = i
		}
		if name == "custom-a" {
			posA = i
		}
		if name == "custom-b" {
			posB = i
		}
	}

	if posIndex == -1 || posA == -1 || posB == -1 {
		t.Fatal("missing expected preprocessor in result")
	}

	if posB < posIndex {
		t.Fatalf("custom-b should come after index, got positions %d and %d", posB, posIndex)
	}

	if posA >= posB {
		t.Fatalf("custom-a should come before custom-b, got positions %d and %d", posA, posB)
	}
}

func TestResolvePreprocessorOrder_NoDefaults(t *testing.T) {
	// Test with includeDefaults = false
	defaults := []string{"index"}
	customs := map[string]struct {
		Before []string
		After  []string
	}{
		"custom-a": {},
	}

	result, err := ResolvePreprocessorOrder(defaults, customs, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have custom-a, not index
	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d: %v", len(result), result)
	}

	if result[0] != "custom-a" {
		t.Fatalf("expected custom-a, got %s", result[0])
	}
}
