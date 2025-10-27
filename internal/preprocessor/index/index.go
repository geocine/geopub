package index

import (
	"github.com/geocine/geopub/internal/models"
)

// IndexPreprocessor handles converting README files and creating indexes
type IndexPreprocessor struct {
}

// NewIndexPreprocessor creates a new index preprocessor
func NewIndexPreprocessor() *IndexPreprocessor {
	return &IndexPreprocessor{}
}

// Name returns the preprocessor name
func (i *IndexPreprocessor) Name() string {
	return "index"
}

// Process processes all chapters in the book
func (i *IndexPreprocessor) Process(book *models.Book) error {
	// Rename README to introduction if it exists
	for j, item := range book.Items {
		if ch, ok := item.(*models.Chapter); ok {
			if ch.Path != nil && *ch.Path == "README.md" {
				// Convert README to introduction-like chapter
				ch.Name = "Introduction"
				// Keep original path for reference
				book.Items[j] = ch
			}
		}
	}
	return nil
}
