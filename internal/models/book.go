package models

import (
	"path/filepath"
)

// SectionNumber represents a chapter's section number (e.g., "1.2.3")
type SectionNumber struct {
	Parts []int
}

// String returns the string representation of a section number
func (sn *SectionNumber) String() string {
	if sn == nil || len(sn.Parts) == 0 {
		return ""
	}
	result := ""
	for i, part := range sn.Parts {
		if i > 0 {
			result += "."
		}
		result += string(rune(part + 48)) // Convert to digit
	}
	return result
}

// Book represents a collection of chapters/items
type Book struct {
	Items []BookItem
}

// NewBook creates an empty book
func NewBook() *Book {
	return &Book{
		Items: make([]BookItem, 0),
	}
}

// NewBookWithItems creates a book with initial items
func NewBookWithItems(items []BookItem) *Book {
	return &Book{
		Items: items,
	}
}

// PushItem appends a BookItem to the book
func (b *Book) PushItem(item BookItem) {
	b.Items = append(b.Items, item)
}

// Chapters returns only non-draft chapters from the book
func (b *Book) Chapters() []*Chapter {
	var chapters []*Chapter
	for _, item := range b.Items {
		if ch, ok := item.(*Chapter); ok && !ch.IsDraftChapter() {
			chapters = append(chapters, ch)
		}
		// Recursively check sub-items for chapters
		if ch, ok := item.(*Chapter); ok {
			chapters = append(chapters, ch.GetAllChapters()...)
		}
	}
	return chapters
}

// IterAll iterates over all items depth-first (including nested)
func (b *Book) IterAll() []BookItem {
	var result []BookItem
	for _, item := range b.Items {
		result = append(result, item)
		if ch, ok := item.(*Chapter); ok {
			result = append(result, ch.IterSubItems()...)
		}
	}
	return result
}

// BookItemType represents the type of a BookItem
type BookItemType int

const (
	ChapterItem BookItemType = iota
	SeparatorItem
	PartTitleItem
)

// BookItem is an interface for different types of book items
type BookItem interface {
	Type() BookItemType
}

// Chapter represents a single chapter/section
type Chapter struct {
	Name        string         // Chapter name/title
	Content     string         // Markdown content
	Number      *SectionNumber // Section number (e.g., 1.2.3)
	SubItems    []BookItem     // Nested items
	Path        *string        // Relative path to the markdown file (relative to src/)
	SourcePath  *string        // Actual path on disk
	ParentNames []string       // Names of parent chapters
	IsDraft     bool           // Is this a draft chapter?
}

// NewChapter creates a new chapter with content
func NewChapter(name, content string, path string, parentNames []string) *Chapter {
	return &Chapter{
		Name:        name,
		Content:     content,
		Path:        &path,
		SubItems:    make([]BookItem, 0),
		ParentNames: parentNames,
		IsDraft:     false,
	}
}

// NewDraftChapter creates a draft chapter (no file)
func NewDraftChapter(name string, parentNames []string) *Chapter {
	return &Chapter{
		Name:        name,
		Content:     "",
		SubItems:    make([]BookItem, 0),
		ParentNames: parentNames,
		IsDraft:     true,
	}
}

// Type returns the BookItem type
func (c *Chapter) Type() BookItemType {
	return ChapterItem
}

// IsDraftChapter returns true if this is a draft chapter
func (c *Chapter) IsDraftChapter() bool {
	return c.IsDraft
}

// GetAllChapters recursively returns all chapters including nested ones
func (c *Chapter) GetAllChapters() []*Chapter {
	var chapters []*Chapter
	for _, item := range c.SubItems {
		if ch, ok := item.(*Chapter); ok {
			chapters = append(chapters, ch)
			chapters = append(chapters, ch.GetAllChapters()...)
		}
	}
	return chapters
}

// IterSubItems recursively iterates over sub-items
func (c *Chapter) IterSubItems() []BookItem {
	var result []BookItem
	for _, item := range c.SubItems {
		result = append(result, item)
		if ch, ok := item.(*Chapter); ok {
			result = append(result, ch.IterSubItems()...)
		}
	}
	return result
}

// Separator represents a separator/divider between sections
type Separator struct{}

// Type returns the BookItem type
func (s *Separator) Type() BookItemType {
	return SeparatorItem
}

// PartTitle represents a part title for grouping chapters
type PartTitle struct {
	Title string
}

// Type returns the BookItem type
func (p *PartTitle) Type() BookItemType {
	return PartTitleItem
}

// PathToRoot calculates relative path from the given path back to root
// E.g., "some/relative/path" -> "../../"
func PathToRoot(path string) string {
	dir := filepath.Dir(path)
	if dir == "." {
		return ""
	}

	depth := 0
	for dir != "." && dir != "" {
		depth++
		dir = filepath.Dir(dir)
	}

	result := ""
	for i := 0; i < depth; i++ {
		result += "../"
	}
	return result
}
