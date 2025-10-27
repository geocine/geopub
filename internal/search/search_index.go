// Package search implements full-text search indexing with inverted index and stemming.
// It creates search indexes that are compatible with the elasticlunr.js format.
package search

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"unicode"
)

const (
	ElasticlunrVersion   = "0.9.5"
	maxWordLengthToIndex = 80
)

// TermFrequency represents the term frequency in the inverted index
type TermFrequency struct {
	TF float64 `json:"tf"`
}

// IndexItem is a node in the trie-like inverted index structure
type IndexItem struct {
	Docs     map[string]TermFrequency `json:"docs"`
	DF       int64                    `json:"df"`
	Children map[string]*IndexItem    `json:""`
}

// MarshalJSON custom marshaling for IndexItem
func (ii *IndexItem) MarshalJSON() ([]byte, error) {
	// We need to manually marshal to include children at the same level
	data := make(map[string]interface{})

	if len(ii.Docs) > 0 {
		data["docs"] = ii.Docs
	}
	if ii.DF > 0 {
		data["df"] = ii.DF
	}

	// Add children at the same level
	for key, child := range ii.Children {
		childData, err := child.MarshalJSON()
		if err != nil {
			return nil, err
		}
		var childMap map[string]interface{}
		if err := json.Unmarshal(childData, &childMap); err != nil {
			return nil, err
		}
		data[key] = childMap
	}

	return json.Marshal(data)
}

// UnmarshalJSON custom unmarshaling for IndexItem
func (ii *IndexItem) UnmarshalJSON(data []byte) error {
	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return err
	}

	ii.Docs = make(map[string]TermFrequency)
	ii.Children = make(map[string]*IndexItem)

	for key, val := range dataMap {
		switch key {
		case "docs":
			// Parse docs
			if docsData, ok := val.(map[string]interface{}); ok {
				for docID, tfVal := range docsData {
					if tfMap, ok := tfVal.(map[string]interface{}); ok {
						if tf, ok := tfMap["tf"].(float64); ok {
							ii.Docs[docID] = TermFrequency{TF: tf}
						}
					}
				}
			}
		case "df":
			if df, ok := val.(float64); ok {
				ii.DF = int64(df)
			}
		default:
			// This is a child node
			if childData, err := json.Marshal(val); err == nil {
				var child IndexItem
				if err := child.UnmarshalJSON(childData); err == nil {
					ii.Children[key] = &child
				}
			}
		}
	}

	return nil
}

// InvertedIndex represents an inverted index for a field
type InvertedIndex struct {
	Root *IndexItem
}

// NewInvertedIndex creates a new inverted index
func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		Root: &IndexItem{
			Docs:     make(map[string]TermFrequency),
			Children: make(map[string]*IndexItem),
		},
	}
}

// AddToken adds a token to the inverted index
func (ii *InvertedIndex) AddToken(docRef, token string, termFreq float64) {
	if len(token) == 0 {
		return
	}

	current := ii.Root
	for _, ch := range token {
		key := string(ch)
		if _, exists := current.Children[key]; !exists {
			current.Children[key] = &IndexItem{
				Docs:     make(map[string]TermFrequency),
				Children: make(map[string]*IndexItem),
			}
		}
		current = current.Children[key]
	}

	if _, exists := current.Docs[docRef]; !exists {
		current.DF++
	}
	current.Docs[docRef] = TermFrequency{TF: termFreq}
}

// GetNode retrieves a node for a given token
func (ii *InvertedIndex) GetNode(token string) *IndexItem {
	current := ii.Root
	for _, ch := range token {
		key := string(ch)
		if child, exists := current.Children[key]; exists {
			current = child
		} else {
			return nil
		}
	}
	return current
}

// HasToken checks if a token exists in the index
func (ii *InvertedIndex) HasToken(token string) bool {
	return ii.GetNode(token) != nil
}

// GetDocs returns the documents containing a token
func (ii *InvertedIndex) GetDocs(token string) map[string]float64 {
	node := ii.GetNode(token)
	if node == nil {
		return nil
	}

	result := make(map[string]float64)
	for docID, tf := range node.Docs {
		result[docID] = tf.TF
	}
	return result
}

// GetDocFrequency returns the document frequency of a token
func (ii *InvertedIndex) GetDocFrequency(token string) int64 {
	node := ii.GetNode(token)
	if node == nil {
		return 0
	}
	return node.DF
}

// DocumentStore stores document data
type DocumentStore struct {
	Save    bool                              `json:"save"`
	Docs    map[string]map[string]interface{} `json:"docs"`
	DocInfo map[string]map[string]int         `json:"docInfo"`
	Length  int                               `json:"length"`
}

// NewDocumentStore creates a new document store
func NewDocumentStore(save bool) *DocumentStore {
	return &DocumentStore{
		Save:    save,
		Docs:    make(map[string]map[string]interface{}),
		DocInfo: make(map[string]map[string]int),
		Length:  0,
	}
}

// AddDoc adds a document to the store
func (ds *DocumentStore) AddDoc(docRef string, doc map[string]interface{}) {
	if _, exists := ds.Docs[docRef]; !exists {
		ds.Length++
	}

	if ds.Save {
		ds.Docs[docRef] = doc
	} else {
		ds.Docs[docRef] = make(map[string]interface{})
	}
}

// GetDoc retrieves a document
func (ds *DocumentStore) GetDoc(docRef string) (map[string]interface{}, bool) {
	doc, exists := ds.Docs[docRef]
	return doc, exists
}

// HasDoc checks if a document exists
func (ds *DocumentStore) HasDoc(docRef string) bool {
	_, exists := ds.Docs[docRef]
	return exists
}

// AddFieldLength adds field length information
func (ds *DocumentStore) AddFieldLength(docRef, field string, length int) {
	if _, exists := ds.DocInfo[docRef]; !exists {
		ds.DocInfo[docRef] = make(map[string]int)
	}
	ds.DocInfo[docRef][field] = length
}

// GetFieldLength retrieves field length information
func (ds *DocumentStore) GetFieldLength(docRef, field string) int {
	if docInfo, exists := ds.DocInfo[docRef]; exists {
		if length, exists := docInfo[field]; exists {
			return length
		}
	}
	return 0
}

// Index is the main search index
type Index struct {
	Fields        []string                  `json:"fields"`
	FieldIndexes  map[string]*InvertedIndex `json:"-"`
	Ref           string                    `json:"ref"`
	Version       string                    `json:"version"`
	Pipeline      []string                  `json:"pipeline"`
	Lang          string                    `json:"lang"`
	DocumentStore *DocumentStore            `json:"documentStore"`
}

// NewIndex creates a new index with the given fields
func NewIndex(fields []string) *Index {
	fieldIndexes := make(map[string]*InvertedIndex)
	for _, field := range fields {
		fieldIndexes[field] = NewInvertedIndex()
	}

	return &Index{
		Fields:        fields,
		FieldIndexes:  fieldIndexes,
		Ref:           "id",
		Version:       ElasticlunrVersion,
		Pipeline:      []string{"trimmer", "stopWordFilter", "stemmer"},
		Lang:          "English",
		DocumentStore: NewDocumentStore(true),
	}
}

// AddDoc adds a document to the index
func (idx *Index) AddDoc(doc map[string]interface{}) {
	docRef := fmt.Sprintf("%v", doc[idx.Ref])

	// Ensure doc reference is stored as string
	docCopy := make(map[string]interface{})
	for key, val := range doc {
		if key == idx.Ref {
			docCopy[key] = docRef
		} else {
			docCopy[key] = val
		}
	}

	idx.DocumentStore.AddDoc(docRef, docCopy)

	// Process each field
	tokenFreq := make(map[string]map[string]int)

	for _, field := range idx.Fields {
		if field == idx.Ref {
			continue
		}

		if fieldVal, exists := doc[field]; exists {
			fieldStr := fmt.Sprintf("%v", fieldVal)

			// Tokenize
			tokens := tokenize(fieldStr)

			// Apply pipeline: trimmer, stopWordFilter, stemmer
			processedTokens := make([]string, 0, len(tokens))
			for _, token := range tokens {
				// Trimmer: remove non-alphanumeric characters (already done in tokenize)

				// StopWordFilter: skip stop words
				if stopWords[token] {
					continue
				}

				// Stemmer: reduce to root form
				stemmed := stem(token)
				if stemmed != "" {
					processedTokens = append(processedTokens, stemmed)
				}
			}

			// Count unique stemmed tokens for field length
			uniqueTokens := make(map[string]bool)
			for _, token := range processedTokens {
				uniqueTokens[token] = true
			}
			idx.DocumentStore.AddFieldLength(docRef, field, len(uniqueTokens))

			// Calculate token frequencies
			if _, exists := tokenFreq[field]; !exists {
				tokenFreq[field] = make(map[string]int)
			}
			for _, token := range processedTokens {
				tokenFreq[field][token]++
			}

			// Add tokens to inverted index
			for token, count := range tokenFreq[field] {
				freq := math.Sqrt(float64(count))
				idx.FieldIndexes[field].AddToken(docRef, token, freq)
			}
		}
	}
}

// ToMap converts the index to a map suitable for JSON serialization
func (idx *Index) ToMap() map[string]interface{} {
	nestedIndex := make(map[string]interface{})
	for fieldName, fieldIndex := range idx.FieldIndexes {
		// Wrap the root node in a "root" key to maintain elasticlunr-compatible format
		nestedIndex[fieldName] = map[string]interface{}{
			"root": fieldIndex.Root,
		}
	}

	return map[string]interface{}{
		"documentStore": idx.DocumentStore,
		"index":         nestedIndex,
		"lang":          idx.Lang,
		"pipeline":      idx.Pipeline,
		"ref":           idx.Ref,
		"version":       idx.Version,
		"fields":        idx.Fields,
	}
}

// Tokenizer implementation
func tokenize(text string) []string {
	tokens := make([]string, 0)
	var word strings.Builder

	for _, r := range text {
		if unicode.IsSpace(r) || r == '-' {
			if word.Len() > 0 {
				token := strings.ToLower(strings.TrimSpace(word.String()))
				if token != "" && len(token) <= maxWordLengthToIndex {
					tokens = append(tokens, token)
				}
				word.Reset()
			}
		} else {
			word.WriteRune(r)
		}
	}

	if word.Len() > 0 {
		token := strings.ToLower(strings.TrimSpace(word.String()))
		if token != "" && len(token) <= maxWordLengthToIndex {
			tokens = append(tokens, token)
		}
	}

	return tokens
}

// Stop word list
var stopWords = map[string]bool{
	"":        true,
	"a":       true,
	"able":    true,
	"about":   true,
	"across":  true,
	"after":   true,
	"all":     true,
	"almost":  true,
	"also":    true,
	"am":      true,
	"among":   true,
	"an":      true,
	"and":     true,
	"any":     true,
	"are":     true,
	"as":      true,
	"at":      true,
	"be":      true,
	"because": true,
	"been":    true,
	"but":     true,
	"by":      true,
	"can":     true,
	"cannot":  true,
	"could":   true,
	"dear":    true,
	"did":     true,
	"do":      true,
	"does":    true,
	"either":  true,
	"else":    true,
	"ever":    true,
	"every":   true,
	"for":     true,
	"from":    true,
	"get":     true,
	"got":     true,
	"had":     true,
	"has":     true,
	"have":    true,
	"he":      true,
	"her":     true,
	"hers":    true,
	"him":     true,
	"his":     true,
	"how":     true,
	"however": true,
	"i":       true,
	"if":      true,
	"in":      true,
	"into":    true,
	"is":      true,
	"it":      true,
	"its":     true,
	"just":    true,
	"least":   true,
	"let":     true,
	"like":    true,
	"likely":  true,
	"may":     true,
	"me":      true,
	"might":   true,
	"most":    true,
	"must":    true,
	"my":      true,
	"neither": true,
	"no":      true,
	"nor":     true,
	"not":     true,
	"of":      true,
	"off":     true,
	"often":   true,
	"on":      true,
	"only":    true,
	"or":      true,
	"other":   true,
	"our":     true,
	"own":     true,
	"rather":  true,
	"said":    true,
	"say":     true,
	"says":    true,
	"she":     true,
	"should":  true,
	"since":   true,
	"so":      true,
	"some":    true,
	"than":    true,
	"that":    true,
	"the":     true,
	"their":   true,
	"them":    true,
	"then":    true,
	"there":   true,
	"these":   true,
	"they":    true,
	"this":    true,
	"tis":     true,
	"to":      true,
	"too":     true,
	"twas":    true,
	"us":      true,
	"wants":   true,
	"was":     true,
	"we":      true,
	"were":    true,
	"what":    true,
	"when":    true,
	"where":   true,
	"which":   true,
	"while":   true,
	"who":     true,
	"whom":    true,
	"why":     true,
	"will":    true,
	"with":    true,
	"would":   true,
	"yet":     true,
	"you":     true,
	"your":    true,
}

// Simple Porter-like stemmer that removes common suffixes aggressively
func stem(word string) string {
	if len(word) <= 2 {
		return word
	}

	word = strings.ToLower(word)

	// Remove common plural and past tense suffixes
	// Try longest suffixes first to avoid over-stemming
	step1Suffixes := []struct {
		suffix string
		minLen int
	}{
		{"ies", 3},
		{"es", 2},
		{"s", 1},
	}

	for _, s := range step1Suffixes {
		if strings.HasSuffix(word, s.suffix) && len(word) > len(s.suffix)+s.minLen {
			word = word[:len(word)-len(s.suffix)]
			break
		}
	}

	// Remove -ed and -ing
	if strings.HasSuffix(word, "ed") && len(word) > 4 {
		word = word[:len(word)-2]
	} else if strings.HasSuffix(word, "ing") && len(word) > 5 {
		word = word[:len(word)-3]
	}

	// Remove common suffixes
	step3Suffixes := []string{
		"tion", "sion", "ment", "ness", "ful", "less", "ity",
		"ous", "ive", "ent", "ant", "able", "ible", "ence", "ance",
	}

	for _, suffix := range step3Suffixes {
		if strings.HasSuffix(word, suffix) && len(word) > len(suffix)+2 {
			word = word[:len(word)-len(suffix)]
			break
		}
	}

	// Remove common endings
	step4Suffixes := []string{"ly", "er", "est"}
	for _, suffix := range step4Suffixes {
		if strings.HasSuffix(word, suffix) && len(word) > len(suffix)+2 {
			word = word[:len(word)-len(suffix)]
			break
		}
	}

	return word
}
