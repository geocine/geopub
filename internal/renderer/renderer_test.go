package renderer

import (
    "testing"

    "github.com/geocine/geopub/internal/models"
    "github.com/stretchr/testify/assert"
)

func TestSlugify(t *testing.T) {
    cases := map[string]string{
        "Hello, World!":      "hello-world",
        "  Trim -- me  ":     "trim-me",
        "Café au lait!":      "café-au-lait",
        "Rust_2024 Edition":  "rust_2024-edition",
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
    sub.Number = &models.SectionNumber{Parts: []int{1,1}}
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

