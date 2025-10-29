package runner

import (
	"fmt"
)

// TopoSort performs a topological sort on preprocessor names based on before/after constraints.
// Returns ordered list of preprocessor names, or an error if a cycle is detected.
func TopoSort(names []string, before map[string][]string, after map[string][]string) ([]string, error) {
	// Build adjacency list from before/after constraints
	// before[A] = [B, C] means A must come after B and C
	// after[A] = [B, C] means A must come before B and C

	adjacency := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize all nodes
	for _, name := range names {
		if _, ok := adjacency[name]; !ok {
			adjacency[name] = []string{}
		}
		if _, ok := inDegree[name]; !ok {
			inDegree[name] = 0
		}
	}

	// Add edges from after constraints (A after [B, C] means B -> A, C -> A)
	for name, afterList := range after {
		for _, after := range afterList {
			// after comes before name
			if _, exists := adjacency[after]; !exists {
				adjacency[after] = []string{}
			}
			adjacency[after] = append(adjacency[after], name)
			inDegree[name]++
		}
	}

	// Add edges from before constraints (A before [B, C] means A -> B, A -> C)
	for name, beforeList := range before {
		for _, before := range beforeList {
			// name comes before before
			if _, exists := adjacency[name]; !exists {
				adjacency[name] = []string{}
			}
			adjacency[name] = append(adjacency[name], before)
			inDegree[before]++
		}
	}

	// Kahn's algorithm for topological sort
	queue := []string{}
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	result := []string{}
	for len(queue) > 0 {
		// Pop from front
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Process neighbors
		for _, neighbor := range adjacency[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for cycles
	if len(result) != len(names) {
		// Find nodes still in a cycle
		cycleNodes := []string{}
		for name, degree := range inDegree {
			if degree > 0 {
				cycleNodes = append(cycleNodes, name)
			}
		}
		return nil, fmt.Errorf("cycle detected in preprocessor ordering constraints involving: %v", cycleNodes)
	}

	return result, nil
}

// ResolvePreprocessorOrder resolves the execution order of preprocessors based on their before/after constraints.
// defaultPreprocessors: list of built-in preprocessor names
// customPreprocessors: map of custom preprocessor name -> (before list, after list)
// includeDefaults: whether to include default preprocessors
// Returns: ordered list of preprocessor names to execute
func ResolvePreprocessorOrder(defaultPreprocessors []string, customPreprocessors map[string]struct {
	Before []string
	After  []string
}, includeDefaults bool) ([]string, error) {
	// Build combined list
	allNames := []string{}
	if includeDefaults {
		allNames = append(allNames, defaultPreprocessors...)
	}

	before := make(map[string][]string)
	after := make(map[string][]string)

	// Add custom preprocessors
	for name, config := range customPreprocessors {
		allNames = append(allNames, name)
		if len(config.Before) > 0 {
			before[name] = config.Before
		}
		if len(config.After) > 0 {
			after[name] = config.After
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	unique := []string{}
	for _, name := range allNames {
		if !seen[name] {
			unique = append(unique, name)
			seen[name] = true
		}
	}

	// Perform topological sort
	return TopoSort(unique, before, after)
}
