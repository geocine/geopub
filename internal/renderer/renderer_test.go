package renderer

import (
	"strings"
	"testing"

	"github.com/geocine/geopub/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Hello, World!":     "hello-world",
		"  Trim -- me  ":    "trim-me",
		"Café au lait!":     "café-au-lait",
		"Rust_2024 Edition": "rust_2024-edition",
	}
	for in, want := range cases {
		got := slugify(in)
		assert.Equal(t, want, got)
	}
}

func TestHtmlEscape(t *testing.T) {
	in := `Tom & Jerry " <>`
	got := htmlEscape(in)
	assert.Equal(t, "Tom &amp; Jerry &quot; &lt;&gt;", got)
}

func TestConvertMarkdownHeadings(t *testing.T) {
	r := NewHtmlRenderer()
	md := "# Title\n\n## Section <em>One</em>\n\nParagraph.\n"
	html, headings := r.convertMarkdown(md)

	// h1 becomes anchor but not included in headings slice
	assert.Contains(t, html, `<h1 id="title"><a class="header" href="#title">Title</a></h1>`)

	// h2 appears with id and anchor link; text is plain without tags in headings
	assert.Contains(t, html, `id="section-one"`)
	assert.Len(t, headings, 1)
	assert.Equal(t, "2", headings[0].Level)
	assert.Equal(t, "Section One", headings[0].Text)
	assert.Equal(t, "section-one", headings[0].ID)
}

func TestGenerateTocListHTML(t *testing.T) {
	// Build a simple book structure
	ch1 := models.NewChapter("Chapter 1", "# Chapter 1", "ch1.md", nil)
	ch1.Number = &models.SectionNumber{Parts: []int{1}}
	sub := models.NewChapter("Sub", "# Sub", "ch1/sec.md", nil)
	sub.Number = &models.SectionNumber{Parts: []int{1, 1}}
	ch1.SubItems = append(ch1.SubItems, sub)

	book := models.NewBookWithItems([]models.BookItem{ch1})

	r := NewHtmlRenderer()
	toc := r.generateTocListHTML(book)

	// Top-level entry with numbering and href
	assert.Contains(t, toc, `<ol class="chapter">`)
	assert.Contains(t, toc, `href="ch1.html"`)
	assert.Contains(t, toc, `>1.</strong> Chapter 1</a>`)
	// Sub-entry contains sec path
	assert.Contains(t, toc, `href="ch1/sec.html"`)
}

func TestGenerateTocWithPartTitles(t *testing.T) {
	// Build a book with part titles
	part1 := &models.PartTitle{Title: "Getting Started"}
	ch1 := models.NewChapter("Chapter 1", "# Ch1", "ch1.md", nil)
	ch1.Number = &models.SectionNumber{Parts: []int{1}}

	part2 := &models.PartTitle{Title: "Advanced"}
	ch2 := models.NewChapter("Chapter 2", "# Ch2", "ch2.md", nil)
	ch2.Number = &models.SectionNumber{Parts: []int{2}}

	book := models.NewBookWithItems([]models.BookItem{part1, ch1, part2, ch2})

	r := NewHtmlRenderer()
	toc := r.generateTocListHTML(book)

	// Check part titles are rendered
	assert.Contains(t, toc, `<li class="part-title">Getting Started</li>`)
	assert.Contains(t, toc, `<li class="part-title">Advanced</li>`)
	assert.Contains(t, toc, `affix`)

	// Check chapters are numbered
	assert.Contains(t, toc, `>1.</strong> Chapter 1</a>`)
	assert.Contains(t, toc, `>2.</strong> Chapter 2</a>`)
}

func TestGenerateTocWithPrefixChapters(t *testing.T) {
	// Prefix chapter (no number)
	intro := models.NewChapter("Introduction", "# Intro", "intro.md", nil)
	// intro.Number is nil - prefix chapters have no numbers

	// Numbered chapter
	ch1 := models.NewChapter("Chapter 1", "# Ch1", "ch1.md", nil)
	ch1.Number = &models.SectionNumber{Parts: []int{1}}

	book := models.NewBookWithItems([]models.BookItem{intro, ch1})

	r := NewHtmlRenderer()
	toc := r.generateTocListHTML(book)

	// Introduction should have affix class and no number
	assert.Contains(t, toc, `class="chapter-item expanded affix ">`)
	assert.Contains(t, toc, `href="intro.html"`)
	assert.Contains(t, toc, `>Introduction</a>`)
	// Should not have a strong tag with number for intro
	assert.NotContains(t, toc, `>1.</strong> Introduction`)

	// Chapter 1 should have number and no affix class
	assert.Contains(t, toc, `>1.</strong> Chapter 1</a>`)
}

func TestRenderTocItemForJSWithoutNumber(t *testing.T) {
	// Test that chapters without numbers render correctly (no empty strong tag)
	intro := models.NewChapter("Introduction", "# Intro", "intro.md", nil)
	// intro.Number is nil

	r := NewHtmlRenderer()
	var buf strings.Builder
	r.renderTocItemForJS(&buf, intro, 0)

	result := buf.String()

	// Should have affix class
	assert.Contains(t, result, `class="chapter-item expanded affix "`)
	// Should NOT have strong tag
	assert.NotContains(t, result, `<strong`)
	// Should have the link
	assert.Contains(t, result, `href="intro.html"`)
	assert.Contains(t, result, `>Introduction</a>`)
}

func TestRenderTocItemForJSWithNumber(t *testing.T) {
	// Test that numbered chapters render with strong tag
	ch1 := models.NewChapter("Chapter 1", "# Ch1", "ch1.md", nil)
	ch1.Number = &models.SectionNumber{Parts: []int{1}}

	r := NewHtmlRenderer()
	var buf strings.Builder
	r.renderTocItemForJS(&buf, ch1, 0)

	result := buf.String()

	// Should NOT have affix class
	assert.NotContains(t, result, `affix`)
	// Should have strong tag with number
	assert.Contains(t, result, `<strong aria-hidden="true">1.</strong>`)
	// Should have the link
	assert.Contains(t, result, `href="ch1.html"`)
	assert.Contains(t, result, `Chapter 1</a>`)
}
