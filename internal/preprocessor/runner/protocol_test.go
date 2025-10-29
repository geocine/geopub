package runner

import (
	"encoding/json"
	"testing"

	"github.com/geocine/geopub/internal/models"
)

func TestBookToJson(t *testing.T) {
	// Create a test book
	ch1 := models.NewChapter("Chapter 1", "Content 1", "ch1.md", []string{})
	ch2 := models.NewChapter("Chapter 2", "Content 2", "ch2.md", []string{})
	sep := &models.Separator{}

	book := models.NewBook()
	book.PushItem(ch1)
	book.PushItem(sep)
	book.PushItem(ch2)

	// Convert to JSON
	jsonBook := BookToJson(book)

	if jsonBook == nil {
		t.Fatal("BookToJson returned nil")
	}

	if len(jsonBook.Sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(jsonBook.Sections))
	}

	// Check first chapter
	if jsonBook.Sections[0].Chapter == nil {
		t.Fatal("expected first section to be a chapter")
	}
	if jsonBook.Sections[0].Chapter.Name != "Chapter 1" {
		t.Fatalf("expected 'Chapter 1', got '%s'", jsonBook.Sections[0].Chapter.Name)
	}
	if jsonBook.Sections[0].Chapter.Content != "Content 1" {
		t.Fatalf("expected 'Content 1', got '%s'", jsonBook.Sections[0].Chapter.Content)
	}

	// Check separator
	if !jsonBook.Sections[1].IsSeparator {
		t.Fatal("expected second section to be a separator")
	}

	// Check second chapter
	if jsonBook.Sections[2].Chapter == nil {
		t.Fatal("expected third section to be a chapter")
	}
	if jsonBook.Sections[2].Chapter.Name != "Chapter 2" {
		t.Fatalf("expected 'Chapter 2', got '%s'", jsonBook.Sections[2].Chapter.Name)
	}
}

func TestJsonToBook(t *testing.T) {
	// Create a JSON book
	jsonBook := &JsonBook{
		Sections: []JsonSection{
			{
				Chapter: &JsonChapter{
					Name:     "Chapter 1",
					Content:  "Content 1",
					Path:     "ch1.md",
					SubItems: []JsonSection{},
				},
			},
			{
				IsSeparator: true,
			},
			{
				Chapter: &JsonChapter{
					Name:     "Chapter 2",
					Content:  "Content 2",
					Path:     "ch2.md",
					SubItems: []JsonSection{},
				},
			},
		},
	}

	// Convert to book
	book := models.NewBook()
	err := JsonToBook(jsonBook, book)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(book.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(book.Items))
	}

	// Check first chapter
	ch1, ok := book.Items[0].(*models.Chapter)
	if !ok {
		t.Fatal("expected first item to be a chapter")
	}
	if ch1.Name != "Chapter 1" {
		t.Fatalf("expected 'Chapter 1', got '%s'", ch1.Name)
	}
	if ch1.Content != "Content 1" {
		t.Fatalf("expected 'Content 1', got '%s'", ch1.Content)
	}

	// Check separator
	_, ok = book.Items[1].(*models.Separator)
	if !ok {
		t.Fatal("expected second item to be a separator")
	}

	// Check second chapter
	ch2, ok := book.Items[2].(*models.Chapter)
	if !ok {
		t.Fatal("expected third item to be a chapter")
	}
	if ch2.Name != "Chapter 2" {
		t.Fatalf("expected 'Chapter 2', got '%s'", ch2.Name)
	}
}

func TestJsonRoundTrip(t *testing.T) {
	// Create original book
	ch1 := models.NewChapter("Chapter 1", "Content 1", "ch1.md", []string{})
	ch2 := models.NewChapter("Chapter 2", "Modified content", "ch2.md", []string{})

	originalBook := models.NewBook()
	originalBook.PushItem(ch1)
	originalBook.PushItem(ch2)

	// Convert to JSON
	jsonBook := BookToJson(originalBook)

	// Modify the JSON (simulate what a preprocessor would do)
	if jsonBook.Sections[1].Chapter != nil {
		jsonBook.Sections[1].Chapter.Content = "Preprocessor modified content"
	}

	// Convert back to book
	resultBook := models.NewBook()
	err := JsonToBook(jsonBook, resultBook)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify changes were applied
	if resultBook.Items[1].(*models.Chapter).Content != "Preprocessor modified content" {
		t.Fatal("preprocessor modifications not applied")
	}
}

func TestMarshalUnmarshalContext(t *testing.T) {
	// Create a context
	ch := &JsonChapter{
		Name:     "Test Chapter",
		Content:  "Test content",
		SubItems: []JsonSection{},
	}

	ctx := &PreprocessorContext{
		Book: &JsonBook{
			Sections: []JsonSection{
				{
					Chapter: ch,
				},
			},
		},
		Config: map[string]interface{}{
			"book": map[string]interface{}{
				"title": "Test Book",
			},
		},
		Renderer: "html",
		Version:  "0.1",
	}

	// Marshal to JSON
	data, err := json.Marshal(ctx)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var ctx2 PreprocessorContext
	err = json.Unmarshal(data, &ctx2)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify
	if ctx2.Renderer != "html" {
		t.Fatalf("expected renderer 'html', got '%s'", ctx2.Renderer)
	}

	if ctx2.Book.Sections[0].Chapter.Name != "Test Chapter" {
		t.Fatalf("expected chapter name 'Test Chapter', got '%s'", ctx2.Book.Sections[0].Chapter.Name)
	}
}

func TestNestedChapters(t *testing.T) {
	// Create book with nested chapters
	parent := models.NewChapter("Parent", "Parent content", "parent.md", []string{})
	child := models.NewChapter("Child", "Child content", "child.md", []string{"Parent"})
	parent.SubItems = append(parent.SubItems, child)

	book := models.NewBook()
	book.PushItem(parent)

	// Convert to JSON
	jsonBook := BookToJson(book)

	if len(jsonBook.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(jsonBook.Sections))
	}

	if len(jsonBook.Sections[0].Chapter.SubItems) != 1 {
		t.Fatalf("expected 1 sub-item, got %d", len(jsonBook.Sections[0].Chapter.SubItems))
	}

	if jsonBook.Sections[0].Chapter.SubItems[0].Chapter.Name != "Child" {
		t.Fatal("nested chapter not found")
	}

	// Round-trip
	resultBook := models.NewBook()
	err := JsonToBook(jsonBook, resultBook)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resultBook.Items) != 1 {
		t.Fatalf("expected 1 item in result, got %d", len(resultBook.Items))
	}

	resultParent := resultBook.Items[0].(*models.Chapter)
	if len(resultParent.SubItems) != 1 {
		t.Fatalf("expected 1 sub-item in result, got %d", len(resultParent.SubItems))
	}

	resultChild := resultParent.SubItems[0].(*models.Chapter)
	if resultChild.Name != "Child" {
		t.Fatalf("expected 'Child', got '%s'", resultChild.Name)
	}
}
