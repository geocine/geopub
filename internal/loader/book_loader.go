package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/geocine/geopub/internal/config"
	"github.com/geocine/geopub/internal/models"
	"github.com/geocine/geopub/internal/parser"
)

// BookLoader handles loading books from disk
type BookLoader struct {
	rootDir string
	srcDir  string
	config  *config.Config
}

// NewBookLoader creates a new book loader
func NewBookLoader(rootDir string, cfg *config.Config) *BookLoader {
	srcDir := filepath.Join(rootDir, cfg.Book.Src)
	return &BookLoader{
		rootDir: rootDir,
		srcDir:  srcDir,
		config:  cfg,
	}
}

// Load loads a complete book from disk
func (bl *BookLoader) Load() (*models.Book, error) {
	summaryPath := filepath.Join(bl.srcDir, "SUMMARY.md")

	// Read SUMMARY.md
	summaryContent, err := bl.readFile(summaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SUMMARY.md: %w", err)
	}

	// Parse SUMMARY.md
	summary, err := parser.ParseSummary(summaryContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SUMMARY.md: %w", err)
	}

	// Assign section numbers to chapters (geopub parity)
	summary.AssignSectionNumbers()

	// Validate
	if err := parser.ValidateSummaryStructure(summary); err != nil {
		return nil, err
	}

	// Create missing files if configured
	if bl.config.Build.CreateMissing {
		if err := bl.createMissingChapters(summary); err != nil {
			return nil, fmt.Errorf("failed to create missing chapters: %w", err)
		}
	}

	// Load book from disk
	book, err := bl.loadFromDisk(summary)
	if err != nil {
		return nil, err
	}

	return book, nil
}

// LoadFromDisk loads a book using a provided Summary
func (bl *BookLoader) LoadFromDisk(summary *parser.Summary) (*models.Book, error) {
	return bl.loadFromDisk(summary)
}

func (bl *BookLoader) loadFromDisk(summary *parser.Summary) (*models.Book, error) {
	items := make([]models.BookItem, 0)

	// Load all chapters from summary
	allItems := summary.FlattenSummary()
	for _, summaryItem := range allItems {
		bookItem, err := bl.loadSummaryItem(summaryItem, nil)
		if err != nil {
			return nil, err
		}
		if bookItem != nil {
			items = append(items, bookItem)
		}
	}

	book := models.NewBookWithItems(items)
	return book, nil
}

func (bl *BookLoader) loadSummaryItem(item *parser.SummaryItem, parentNames []string) (models.BookItem, error) {
	if parentNames == nil {
		parentNames = []string{}
	}

	switch item.Type {
	case "separator":
		return &models.Separator{}, nil
	case "part-title":
		return &models.PartTitle{Title: item.Title}, nil
	case "link":
		return bl.loadChapter(item, parentNames)
	default:
		return nil, fmt.Errorf("unknown summary item type: %s", item.Type)
	}
}

func (bl *BookLoader) loadChapter(item *parser.SummaryItem, parentNames []string) (models.BookItem, error) {
	var ch *models.Chapter

	if item.Location != nil {
		// Load from file
		location := *item.Location

		// Resolve absolute or relative path
		var filePath string
		if filepath.IsAbs(location) {
			filePath = location
		} else {
			filePath = filepath.Join(bl.srcDir, location)
		}

		// Read file content
		content, err := bl.readFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read chapter '%s': %w", item.Title, err)
		}

		// Strip BOM if present
		if len(content) >= 3 && content[0] == 0xef && content[1] == 0xbb && content[2] == 0xbf {
			content = content[3:]
		}

		// Get relative path
		relPath, err := filepath.Rel(bl.srcDir, filePath)
		if err != nil {
			relPath = location
		}

		ch = models.NewChapter(item.Title, content, relPath, parentNames)
		ch.SourcePath = &filePath
	} else {
		// Draft chapter
		ch = models.NewDraftChapter(item.Title, parentNames)
	}

	// Assign section number if present
	if item.Number != nil {
		ch.Number = bl.convertSectionNumber(item.Number)
	}

	// Load nested items
	newParentNames := append(parentNames, item.Title)
	for _, nestedItem := range item.NestedItems {
		nestedBookItem, err := bl.loadSummaryItem(nestedItem, newParentNames)
		if err != nil {
			return nil, err
		}
		if nestedBookItem != nil {
			ch.SubItems = append(ch.SubItems, nestedBookItem)
		}
	}

	return ch, nil
}

func (bl *BookLoader) convertSectionNumber(num *parser.SectionNumber) *models.SectionNumber {
	if num == nil {
		return nil
	}
	return &models.SectionNumber{Parts: num.Parts}
}

func (bl *BookLoader) createMissingChapters(summary *parser.Summary) error {
	items := summary.FlattenSummary()
	return bl.createMissingRecursive(items)
}

func (bl *BookLoader) createMissingRecursive(items []*parser.SummaryItem) error {
	for _, item := range items {
		if item.Type == "link" && item.Location != nil {
			location := *item.Location
			filePath := filepath.Join(bl.srcDir, location)

			// Check if file exists
			if _, err := os.Stat(filePath); err != nil {
				if os.IsNotExist(err) {
					// Create parent directories
					parentDir := filepath.Dir(filePath)
					if err := os.MkdirAll(parentDir, 0o755); err != nil {
						return fmt.Errorf("failed to create directory '%s': %w", parentDir, err)
					}

					// Create file with title as heading
					title := escapeHtml(item.Title)
					content := fmt.Sprintf("# %s\n", title)

					if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
						return fmt.Errorf("failed to create file '%s': %w", filePath, err)
					}
				} else {
					return err
				}
			}
		}

		// Recursively create for nested items
		if err := bl.createMissingRecursive(item.NestedItems); err != nil {
			return err
		}
	}

	return nil
}

func (bl *BookLoader) readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read '%s': %w", path, err)
	}
	return string(data), nil
}

// Helper function to escape HTML entities
func escapeHtml(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(s)
}

// LoadBook is a convenience function to load a book with default configuration
func LoadBook(rootDir string, cfg *config.Config) (*models.Book, error) {
	loader := NewBookLoader(rootDir, cfg)
	return loader.Load()
}
