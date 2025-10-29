package search

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenizeAndStem(t *testing.T) {
	toks := tokenize("Hello, world! running-runner's studies")
	// tokenize splits on whitespace and hyphen; punctuation (except hyphen) is preserved
	assert.Equal(t, []string{"hello,", "world!", "running", "runner's", "studies"}, toks)

	// Stem a few forms
	assert.Equal(t, "runn", stem("running"))
	assert.Equal(t, "stud", stem("studies"))
	assert.Equal(t, "happines", stem("happiness"))
}

func TestIndexAddDocAndToMap(t *testing.T) {
	idx := NewIndex([]string{"title", "body", "breadcrumbs"})

	doc := map[string]interface{}{
		"title":       "Quick Brown Foxes",
		"body":        "Running jumped runner's",
		"breadcrumbs": "Intro",
	}
	idx.AddDoc(doc)

	m := idx.ToMap()
	// Validate top-level keys
	fields, ok := m["fields"].([]string)
	require.True(t, ok)
	assert.ElementsMatch(t, []string{"title", "body", "breadcrumbs"}, fields)

	nested, ok := m["index"].(map[string]interface{})
	require.True(t, ok)
	// Each field should have a root node structure
	for _, f := range fields {
		v, exists := nested[f]
		require.True(t, exists)
		_, isMap := v.(map[string]interface{})
		assert.True(t, isMap)
	}

	// Ensure JSON marshals without error (shape compatible downstream)
	_, err := json.Marshal(m)
	require.NoError(t, err)
}

func TestIndexLongTokenOmittedFromTrie(t *testing.T) {
	idx := NewIndex([]string{"title", "body"})
	long := "ThisLongWordIsIncludedSoWeCanCheckThatSufficientlyLongWordsAreOmittedFromTheSearchIndex."
	doc := map[string]interface{}{
		"title": "No Headers",
		"body":  long,
	}
	idx.AddDoc(doc)

	// Ensure the long token is not present in the body field trie
	token := strings.ToLower(long)
	// punctuation is preserved by tokenize, but length filter omits the token entirely
	// Try both with punctuation and stripped period
	if idx.FieldIndexes["body"].HasToken(token) {
		t.Fatalf("long token unexpectedly present in trie")
	}
	if idx.FieldIndexes["body"].HasToken(strings.TrimSuffix(token, ".")) {
		t.Fatalf("long token (trimmed) unexpectedly present in trie")
	}
}
