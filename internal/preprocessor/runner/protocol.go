package runner

import (
	"encoding/json"

	"github.com/geocine/geopub/internal/config"
	"github.com/geocine/geopub/internal/models"
)

// PreprocessorContext is the JSON structure sent to and received from preprocessors
// It matches the mdBook preprocessor protocol
type PreprocessorContext struct {
	Book     *JsonBook              `json:"book"`
	Config   map[string]interface{} `json:"config"`
	Renderer string                 `json:"renderer"`
	Version  string                 `json:"version"`
}

// JsonBook represents a book in the preprocessor protocol
type JsonBook struct {
	Sections []JsonSection `json:"sections"`
	// Root path to the book (for context)
	Root string `json:"root,omitempty"`
}

// JsonSection represents a section (chapter or separator) in the preprocessor protocol
type JsonSection struct {
	Chapter *JsonChapter `json:"chapter,omitempty"`
	// Separator appears as { "separator": {} } in mdBook
	IsSeparator bool `json:"separator,omitempty"`
}

// JsonChapter represents a chapter in the preprocessor protocol
type JsonChapter struct {
	Name     string        `json:"name"`
	Content  string        `json:"content"`
	Number   []int         `json:"number,omitempty"`
	SubItems []JsonSection `json:"sub_items"`
	Path     string        `json:"path,omitempty"`
}

// BookToJson converts a GeoPub book to the JSON representation for preprocessors
func BookToJson(book *models.Book) *JsonBook {
	jsonBook := &JsonBook{
		Sections: []JsonSection{},
	}

	for _, item := range book.Items {
		switch v := item.(type) {
		case *models.Chapter:
			jsonBook.Sections = append(jsonBook.Sections, chapterToJsonSection(v))
		case *models.Separator:
			jsonBook.Sections = append(jsonBook.Sections, JsonSection{IsSeparator: true})
		case *models.PartTitle:
			// PartTitle doesn't map directly; skip it for now
		}
	}

	return jsonBook
}

// chapterToJsonSection converts a chapter to a JSON section
func chapterToJsonSection(ch *models.Chapter) JsonSection {
	jsonCh := &JsonChapter{
		Name:     ch.Name,
		Content:  ch.Content,
		SubItems: []JsonSection{},
	}

	// Convert section number
	if ch.Number != nil && len(ch.Number.Parts) > 0 {
		jsonCh.Number = ch.Number.Parts
	}

	// Set path if available
	if ch.Path != nil {
		jsonCh.Path = *ch.Path
	}

	// Convert sub-items
	for _, subItem := range ch.SubItems {
		switch v := subItem.(type) {
		case *models.Chapter:
			jsonCh.SubItems = append(jsonCh.SubItems, chapterToJsonSection(v))
		case *models.Separator:
			jsonCh.SubItems = append(jsonCh.SubItems, JsonSection{IsSeparator: true})
		}
	}

	return JsonSection{Chapter: jsonCh}
}

// JsonToBook converts the JSON representation back to a GeoPub book, applying mutations
func JsonToBook(jsonBook *JsonBook, originalBook *models.Book) error {
	// Create a new book from the JSON structure
	newItems := []models.BookItem{}

	for _, section := range jsonBook.Sections {
		if section.IsSeparator {
			newItems = append(newItems, &models.Separator{})
		} else if section.Chapter != nil {
			ch, err := jsonChapterToChapter(section.Chapter)
			if err != nil {
				return err
			}
			newItems = append(newItems, ch)
		}
	}

	// Replace the book's items
	originalBook.Items = newItems
	return nil
}

// jsonChapterToChapter converts a JSON chapter to a GeoPub chapter
func jsonChapterToChapter(jsonCh *JsonChapter) (*models.Chapter, error) {
	ch := &models.Chapter{
		Name:     jsonCh.Name,
		Content:  jsonCh.Content,
		SubItems: []models.BookItem{},
	}

	// Set path if available
	if jsonCh.Path != "" {
		ch.Path = &jsonCh.Path
	}

	// Convert number
	if len(jsonCh.Number) > 0 {
		ch.Number = &models.SectionNumber{Parts: jsonCh.Number}
	}

	// Convert sub-items
	for _, subSection := range jsonCh.SubItems {
		if subSection.IsSeparator {
			ch.SubItems = append(ch.SubItems, &models.Separator{})
		} else if subSection.Chapter != nil {
			subCh, err := jsonChapterToChapter(subSection.Chapter)
			if err != nil {
				return nil, err
			}
			ch.SubItems = append(ch.SubItems, subCh)
		}
	}

	return ch, nil
}

// NewPreprocessorContext creates a context for passing to a preprocessor
func NewPreprocessorContext(book *models.Book, cfg *config.Config, renderer string) *PreprocessorContext {
	// Convert config to map for JSON marshaling
	configMap := make(map[string]interface{})
	if cfg != nil {
		// Include book config
		configMap["book"] = map[string]interface{}{
			"title":       cfg.Book.Title,
			"authors":     cfg.Book.Authors,
			"description": cfg.Book.Description,
			"language":    cfg.Book.Language,
		}
		// Include build config
		configMap["build"] = map[string]interface{}{
			"build-dir": cfg.Build.BuildDir,
		}
		// Include preprocessor configs
		configMap["preprocessor"] = cfg.Preprocessor
	}

	return &PreprocessorContext{
		Book:     BookToJson(book),
		Config:   configMap,
		Renderer: renderer,
		Version:  "0.1",
	}
}

// UnmarshalContext unmarshals a context from JSON
func UnmarshalContext(data []byte) (*PreprocessorContext, error) {
	var ctx PreprocessorContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, err
	}
	return &ctx, nil
}
