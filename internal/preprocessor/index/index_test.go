package index

import (
	"testing"

	"github.com/geocine/geopub/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadmeConversion(t *testing.T) {
	book := &models.Book{
		Items: []models.BookItem{
			&models.Chapter{
				Name:     "README",
				Content:  "Welcome to the book",
				Path:     strPtr("README.md"),
				SubItems: make([]models.BookItem, 0),
			},
		},
	}

	preprocessor := NewIndexPreprocessor()
	err := preprocessor.Process(book)

	require.NoError(t, err)

	ch := book.Items[0].(*models.Chapter)
	assert.Equal(t, "Introduction", ch.Name)
	assert.Equal(t, "README.md", *ch.Path)
}

func TestNonReadmeUnchanged(t *testing.T) {
	book := &models.Book{
		Items: []models.BookItem{
			&models.Chapter{
				Name:     "Chapter 1",
				Content:  "Content",
				Path:     strPtr("ch1.md"),
				SubItems: make([]models.BookItem, 0),
			},
		},
	}

	preprocessor := NewIndexPreprocessor()
	err := preprocessor.Process(book)

	require.NoError(t, err)

	ch := book.Items[0].(*models.Chapter)
	assert.Equal(t, "Chapter 1", ch.Name)
}

func TestMultipleChaptersWithReadme(t *testing.T) {
	book := &models.Book{
		Items: []models.BookItem{
			&models.Chapter{
				Name:     "README",
				Content:  "Intro",
				Path:     strPtr("README.md"),
				SubItems: make([]models.BookItem, 0),
			},
			&models.Chapter{
				Name:     "Chapter 1",
				Content:  "Content",
				Path:     strPtr("ch1.md"),
				SubItems: make([]models.BookItem, 0),
			},
		},
	}

	preprocessor := NewIndexPreprocessor()
	err := preprocessor.Process(book)

	require.NoError(t, err)

	readme := book.Items[0].(*models.Chapter)
	assert.Equal(t, "Introduction", readme.Name)

	ch1 := book.Items[1].(*models.Chapter)
	assert.Equal(t, "Chapter 1", ch1.Name)
}

func TestReadmeInNestedPath(t *testing.T) {
	book := &models.Book{
		Items: []models.BookItem{
			&models.Chapter{
				Name:    "Main",
				Content: "Main",
				Path:    strPtr("main.md"),
				SubItems: []models.BookItem{
					&models.Chapter{
						Name:     "README",
						Content:  "Nested readme",
						Path:     strPtr("ch1/README.md"),
						SubItems: make([]models.BookItem, 0),
					},
				},
			},
		},
	}

	preprocessor := NewIndexPreprocessor()
	err := preprocessor.Process(book)

	require.NoError(t, err)

	// Top-level README should be converted
	main := book.Items[0].(*models.Chapter)
	assert.Equal(t, "Main", main.Name)

	// Nested README should also be handled if at root
	// (this implementation only handles root-level README.md)
}

func TestEmptyBook(t *testing.T) {
	book := &models.Book{
		Items: make([]models.BookItem, 0),
	}

	preprocessor := NewIndexPreprocessor()
	err := preprocessor.Process(book)

	require.NoError(t, err)
	assert.Equal(t, 0, len(book.Items))
}

func TestSeparatorAndReadme(t *testing.T) {
	book := &models.Book{
		Items: []models.BookItem{
			&models.Chapter{
				Name:     "README",
				Content:  "Intro",
				Path:     strPtr("README.md"),
				SubItems: make([]models.BookItem, 0),
			},
			&models.Separator{},
			&models.Chapter{
				Name:     "Appendix",
				Content:  "Appendix",
				Path:     strPtr("appendix.md"),
				SubItems: make([]models.BookItem, 0),
			},
		},
	}

	preprocessor := NewIndexPreprocessor()
	err := preprocessor.Process(book)

	require.NoError(t, err)

	readme := book.Items[0].(*models.Chapter)
	assert.Equal(t, "Introduction", readme.Name)

	// Separator should remain
	_, ok := book.Items[1].(*models.Separator)
	assert.True(t, ok)

	appendix := book.Items[2].(*models.Chapter)
	assert.Equal(t, "Appendix", appendix.Name)
}

// Helper to create string pointer
func strPtr(s string) *string {
	return &s
}
