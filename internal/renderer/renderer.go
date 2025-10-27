package renderer

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	htmlutil "html"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aymerick/raymond"
	"github.com/geocine/geopub/internal/config"
	"github.com/geocine/geopub/internal/models"
	"github.com/geocine/geopub/internal/search"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	ghtml "github.com/yuin/goldmark/renderer/html"
)

// RenderContext holds context for rendering
type RenderContext struct {
	Root      string
	DestDir   string
	Book      *models.Book
	Config    *config.Config
	SourceDir string
	// If non-empty, pages inject an SSE live-reload client targeting this path.
	LiveReloadEndpointPath string
	// AssetsFS optionally provides embedded front-end assets (expects paths under "frontend/")
	AssetsFS fs.FS
	// ResourceMap provides mapping original -> fingerprinted asset paths for templates
	ResourceMap map[string]string
}

// HtmlRenderer renders a book to HTML
type HtmlRenderer struct {
	markdown goldmark.Markdown
	book     *models.Book // Store book reference for nav generation
}

// NewHtmlRenderer creates a new HTML renderer
func NewHtmlRenderer() *HtmlRenderer {
    md := goldmark.New(
        goldmark.WithExtensions(
            extension.GFM,
            extension.Footnote,
            extension.DefinitionList,
        ),
        goldmark.WithRendererOptions(
            ghtml.WithUnsafe(),
        ),
    )
    return &HtmlRenderer{markdown: md}
}

// Render renders the book to HTML
func (r *HtmlRenderer) Render(ctx *RenderContext) error {
	r.book = ctx.Book // Store for nav generation

	// Create output directory
	if err := os.MkdirAll(ctx.DestDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Prepare and copy assets first, building the resource mapping for fingerprinting
	if err := r.copyAssets(ctx); err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	// Collect all chapters for TOC and navigation
	allChapters := r.collectChapters(ctx.Book)

	// Render all chapters and their nested items
	for _, item := range ctx.Book.Items {
		if ch, ok := item.(*models.Chapter); ok {
			if err := r.renderChapterRecursive(ctx, ch, allChapters); err != nil {
				return fmt.Errorf("failed to render chapter: %w", err)
			}
		}
	}

	// Create index.html
	if err := r.renderIndex(ctx); err != nil {
		return fmt.Errorf("failed to render index: %w", err)
	}

	// Create other static pages
	if err := r.renderExtraPages(ctx); err != nil {
		return fmt.Errorf("failed to render extra pages: %w", err)
	}

	// Generate search index
	if err := r.generateSearchIndex(ctx); err != nil {
		return fmt.Errorf("failed to generate search index: %w", err)
	}

	// Copy non-Markdown files from source directory into output directory
	if err := r.copyNonMarkdown(ctx); err != nil {
		return fmt.Errorf("failed to copy source assets: %w", err)
	}

	return nil
}

// renderChapterRecursive renders a chapter and its sub-chapters
func (r *HtmlRenderer) renderChapterRecursive(ctx *RenderContext, chapter *models.Chapter, allChapters []*models.Chapter) error {
	// Render this chapter
	if err := r.renderChapter(ctx, chapter, allChapters); err != nil {
		return err
	}

	// Recursively render sub-chapters
	for _, item := range chapter.SubItems {
		if subCh, ok := item.(*models.Chapter); ok {
			if err := r.renderChapterRecursive(ctx, subCh, allChapters); err != nil {
				return err
			}
		}
	}

	return nil
}

// collectChapters flattens all chapters for navigation
func (r *HtmlRenderer) collectChapters(book *models.Book) []*models.Chapter {
	var chapters []*models.Chapter
	for _, item := range book.Items {
		if ch, ok := item.(*models.Chapter); ok {
			chapters = append(chapters, ch)
			chapters = append(chapters, r.collectSubChapters(ch)...)
		}
	}
	return chapters
}

// collectSubChapters recursively collects nested chapters
func (r *HtmlRenderer) collectSubChapters(ch *models.Chapter) []*models.Chapter {
	var chapters []*models.Chapter
	for _, item := range ch.SubItems {
		if subCh, ok := item.(*models.Chapter); ok {
			chapters = append(chapters, subCh)
			chapters = append(chapters, r.collectSubChapters(subCh)...)
		}
	}
	return chapters
}

// convertMarkdown converts markdown to HTML and extracts headings for TOC
func (r *HtmlRenderer) convertMarkdown(content string) (string, []HeadingInfo) {
	// Pre-scan footnote refs to map goldmark numbers to labels (for later footnote transform)
	numToLabel := map[string]string{}
	seen := map[string]bool{}
	idx := 1
    // capture [^label] that are not definitions ([^label]:)
    for i := 0; i+2 < len(content); {
        j := strings.Index(content[i:], "[^")
        if j == -1 { break }
        i += j + 2
        // find closing ]
        k := strings.IndexByte(content[i:], ']')
        if k == -1 { break }
        label := content[i : i+k]
        // next char after ]
        nextIdx := i + k + 1
        isDef := false
        if nextIdx < len(content) && content[nextIdx] == ':' { isDef = true }
        if !isDef {
            if !seen[label] {
                numToLabel[fmt.Sprintf("%d", idx)] = label
                seen[label] = true
                idx++
            }
        }
        i = nextIdx
    }

	var buf bytes.Buffer
	if err := r.markdown.Convert([]byte(content), &buf); err != nil { return "", nil }
	html := buf.String()

	// Extract headings and add unique IDs
	var headings []HeadingInfo
	headingRegex := regexp.MustCompile(`<h([1-6])>(.*?)</h[1-6]>`)
	used := map[string]bool{}
	nextSuffix := map[string]int{}
	html = headingRegex.ReplaceAllStringFunc(html, func(match string) string {
		parts := headingRegex.FindStringSubmatch(match)
		if len(parts) < 3 { return match }
		level := parts[1]
		text := parts[2]
		plain := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(text, "")
		plain = htmlutil.UnescapeString(plain)
		base := slugify(plain)
		id := base
		if used[id] {
			for {
				nextSuffix[base]++
				cand := base
				if cand != "" { cand = cand + "-" + fmt.Sprintf("%d", nextSuffix[base]) } else { cand = "-" + fmt.Sprintf("%d", nextSuffix[base]) }
				if !used[cand] { id = cand; break }
			}
		}
		used[id] = true
		if level != "1" {
			headings = append(headings, HeadingInfo{ Level: level, Text: plain, ID: id })
		}
		return fmt.Sprintf(`<h%s id="%s"><a class="header" href="#%s">%s</a></h%s>`, level, id, id, text, level)
	})

	// Post-process
	html = transformFootnotes(html, numToLabel)
	html = transformImages(html)
	html = transformLinkTermParagraphsToDL(html)
	html = transformDefinitionLists(html)
	html = transformLinksMdToHtml(html)
	html = transformCollapseBrWhitespace(html)
	html = transformAdmonitions(html)
	return html, headings
}

// HeadingInfo represents a heading in the document
type HeadingInfo struct {
	Level string
	Text  string
	ID    string
}

// slugify converts text to URL slug using geopub-compatible algorithm
func slugify(text string) string {
    s := strings.ToLower(text)
    // Remove characters that are not unicode letters, numbers, whitespace, hyphen, or underscore
    s = regexp.MustCompile(`[^\p{L}\p{N}\s_-]`).ReplaceAllString(s, "")
    // Collapse whitespace to single hyphens; preserve underscores and hyphens
    s = regexp.MustCompile(`\s+`).ReplaceAllString(s, "-")
    // Collapse multiple hyphens
    s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")
    s = strings.Trim(s, "-")
    return s
}

// renderTOCItemAbsolute recursively renders TOC items with absolute paths
func (r *HtmlRenderer) renderTOCItemAbsolute(buf *strings.Builder, ch *models.Chapter, depth int) {
	if ch.Path == nil {
		return
	}

	// Use absolute path from root (path_to_root will be prepended by JS)
	path := strings.TrimSuffix(*ch.Path, ".md") + ".html"
	path = strings.ReplaceAll(path, "\\", "/") // Normalize Windows paths
	fmt.Fprintf(buf, `<li><a href="%s">%s</a>`, path, htmlEscape(ch.Name))

	if len(ch.SubItems) > 0 {
		buf.WriteString(`<ul>`)
		for _, item := range ch.SubItems {
			if subCh, ok := item.(*models.Chapter); ok {
				r.renderTOCItemAbsolute(buf, subCh, depth+1)
			}
		}
		buf.WriteString(`</ul>`)
	}

	buf.WriteString(`</li>`)
}

// generateTOCAbsolute generates table of contents with absolute paths
func (r *HtmlRenderer) generateTOCAbsolute(book *models.Book) string {
	var buf strings.Builder
	buf.WriteString(`<nav class="sidebar-nav"><ul>`)

	for _, item := range book.Items {
		if ch, ok := item.(*models.Chapter); ok {
			r.renderTOCItemAbsolute(&buf, ch, 0)
		} else if _, ok := item.(*models.Separator); ok {
			buf.WriteString(`</ul><hr/><ul>`)
		}
	}

	buf.WriteString(`</ul></nav>`)
	return buf.String()
}

// renderChapter renders a single chapter to an HTML file
func (r *HtmlRenderer) renderChapter(ctx *RenderContext, chapter *models.Chapter, allChapters []*models.Chapter) error {
	// Convert markdown to HTML with heading anchors
	htmlContent, _ := r.convertMarkdown(chapter.Content)

	// Generate filename preserving nested structure
	path := ""
	if chapter.Path != nil {
		path = *chapter.Path
	}
	path = strings.TrimSuffix(path, ".md")
	outPath := filepath.Join(ctx.DestDir, path+".html")

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Find prev/next chapters
	var prevCh, nextCh *models.Chapter
	for i, ch := range allChapters {
		if ch == chapter {
			if i > 0 {
				prevCh = allChapters[i-1]
			}
			if i < len(allChapters)-1 {
				nextCh = allChapters[i+1]
			}
			break
		}
	}

	// Previously used for static TOC; now sidebar JS handles active links
	_ = r.generateTOCAbsolute(ctx.Book)

	// Calculate depth for path_to_root
	// Normalize backslashes to forward slashes for consistent counting
	normalizedPath := strings.ReplaceAll(path, "\\", "/")
	depth := strings.Count(normalizedPath, "/")

	// Render with Handlebars page template
	var prevData, nextData *struct{ Link string }
	if prevCh != nil && prevCh.Path != nil {
		prevData = &struct{ Link string }{Link: strings.ReplaceAll(strings.TrimSuffix(*prevCh.Path, ".md")+".html", "\\", "/")}
	}
	if nextCh != nil && nextCh.Path != nil {
		nextData = &struct{ Link string }{Link: strings.ReplaceAll(strings.TrimSuffix(*nextCh.Path, ".md")+".html", "\\", "/")}
	}

	pd := &pageData{
		Language:           ctx.Config.Book.Language,
		DefaultTheme:       "light",
		PreferredDarkTheme: "navy",
		TextDirection:      "ltr",
		Title:              fmt.Sprintf("%s - %s", chapter.Name, ctx.Config.Book.Title),
		Description:        ctx.Config.Book.Description,
		FaviconSvg:         true,
		FaviconPng:         true,
		PrintEnable:        true,
		AdditionalCSS:      []string{},
		AdditionalJS:       []string{},
		MathJaxSupport:     false,
		SearchJS:           true,
		SearchEnabled:      true,
		PathToRoot:         strings.Repeat("../", depth),
		BookTitle:          ctx.Config.Book.Title,
		Previous:           prevData,
		Next:               nextData,
		LiveReloadEndpoint: ctx.LiveReloadEndpointPath,
		Content:            raymond.SafeString(htmlContent),
		IsPrint:            false,
	}
	pageHTML, err := renderPageWithHbs(ctx, pd)
	if err != nil {
		return fmt.Errorf("failed to render HBS page: %w", err)
	}

	// Write file
	if err := os.WriteFile(outPath, []byte(pageHTML), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// renderIndex renders the index.html - now shows introduction content
func (r *HtmlRenderer) renderIndex(ctx *RenderContext) error {
	// Find first chapter (introduction)
	var firstCh *models.Chapter
	for _, item := range ctx.Book.Items {
		if ch, ok := item.(*models.Chapter); ok {
			firstCh = ch
			break
		}
	}

	var htmlContent string
	if firstCh != nil && firstCh.Path != nil {
		var headings []HeadingInfo
		htmlContent, headings = r.convertMarkdown(firstCh.Content)
		_ = headings // unused for index
	} else {
		htmlContent = `<h1 id="introduction"><a class="header" href="#introduction">Introduction</a></h1>
<p>Select a chapter to begin reading.</p>`
	}

	// Find next chapter after first
	var nextCh *models.Chapter
	allChapters := r.collectChapters(ctx.Book)
	if len(allChapters) > 1 {
		nextCh = allChapters[1]
	}

	pd := &pageData{
		Language:           ctx.Config.Book.Language,
		DefaultTheme:       "light",
		PreferredDarkTheme: "navy",
		TextDirection:      "ltr",
		Title:              fmt.Sprintf("%s - %s", "Introduction", ctx.Config.Book.Title),
		Description:        ctx.Config.Book.Description,
		FaviconSvg:         true,
		FaviconPng:         true,
		PrintEnable:        true,
		SearchJS:           true,
		SearchEnabled:      true,
		PathToRoot:         "",
		BookTitle:          ctx.Config.Book.Title,
		Previous:           nil,
		Next: func() *struct{ Link string } {
			if nextCh != nil && nextCh.Path != nil {
				return &struct{ Link string }{Link: strings.ReplaceAll(strings.TrimSuffix(*nextCh.Path, ".md")+".html", "\\", "/")}
			}
			return nil
		}(),
		LiveReloadEndpoint: ctx.LiveReloadEndpointPath,
		Content:            raymond.SafeString(htmlContent),
		IsPrint:            false,
	}
	pageHTML, err := renderPageWithHbs(ctx, pd)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(ctx.DestDir, "index.html"), []byte(pageHTML), 0644)
}

// renderExtraPages generates print.html, 404.html, etc.
func (r *HtmlRenderer) renderExtraPages(ctx *RenderContext) error {
	// .nojekyll - tells GitHub Pages to serve the site as-is
	nojekyllContent := "This file makes sure that Github Pages doesn't process geopub's output.\n"
	if err := os.WriteFile(filepath.Join(ctx.DestDir, ".nojekyll"), []byte(nojekyllContent), 0644); err != nil {
		return err
	}

	// 404.html - fallback page with full template
	notFoundContent := `<h1 id="document-not-found-404"><a class="header" href="#document-not-found-404">Document not found (404)</a></h1>
<p>This URL is invalid, sorry. Please use the navigation bar or search to continue.</p>`

	pd404 := &pageData{
		Language:           ctx.Config.Book.Language,
		DefaultTheme:       "light",
		PreferredDarkTheme: "navy",
		TextDirection:      "ltr",
		Title:              fmt.Sprintf("%s - %s", "Page not found", ctx.Config.Book.Title),
		BaseUrl:            "/",
		Description:        ctx.Config.Book.Description,
		FaviconSvg:         true,
		FaviconPng:         true,
		PrintEnable:        true,
		SearchJS:           true,
		SearchEnabled:      true,
		PathToRoot:         "",
		BookTitle:          ctx.Config.Book.Title,
		LiveReloadEndpoint: ctx.LiveReloadEndpointPath,
		Content:            raymond.SafeString(notFoundContent),
		IsPrint:            false,
	}
	notFoundHTML, err := renderPageWithHbs(ctx, pd404)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(ctx.DestDir, "404.html"), []byte(notFoundHTML), 0644); err != nil {
		return err
	}

	// print.html - full book on one page
	if err := r.renderPrintPage(ctx); err != nil {
		return err
	}

	// toc.html - sidebar fallback for no-JS browsers
	if err := r.renderTocPage(ctx); err != nil {
		return err
	}

	// toc.js - dynamic sidebar population script (rendered from current TOC)
	if err := r.renderTocJS(ctx); err != nil {
		return err
	}

	// CNAME support
	cname := ctx.Config.GetString("output.renderer.cname", "")
	if cname != "" {
		if err := os.WriteFile(filepath.Join(ctx.DestDir, "CNAME"), []byte(cname), 0644); err != nil {
			return err
		}
	}

	return nil
}

// renderTocPage generates toc.html with table of contents for noscript fallback
func (r *HtmlRenderer) renderTocPage(ctx *RenderContext) error {
	tocList := r.generateTocListHTML(ctx.Book)
	rendered, err := renderTocHTMLWithHbs(ctx, tocList)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(ctx.DestDir, "toc.html"), []byte(rendered), 0644)
}

// generateTocListHTML builds the <ol class="chapter"> list used by toc.html and toc.js
func (r *HtmlRenderer) generateTocListHTML(book *models.Book) string {
	var buf strings.Builder
	buf.WriteString(`<ol class="chapter">`)
	for _, item := range book.Items {
		if ch, ok := item.(*models.Chapter); ok {
			r.renderTocItemForPage(&buf, ch, 0)
		} else if _, ok := item.(*models.Separator); ok {
			buf.WriteString(`<li class="chapter-item expanded "><li class="spacer"></li>`)
		}
	}
	buf.WriteString(`</ol>`)
	return buf.String()
}

// generateTocListForJS builds the <ol class="chapter"> list for toc.js (without target="_parent")
func (r *HtmlRenderer) generateTocListForJS(book *models.Book) string {
	var buf strings.Builder
	buf.WriteString(`<ol class="chapter">`)
	for _, item := range book.Items {
		if ch, ok := item.(*models.Chapter); ok {
			r.renderTocItemForJS(&buf, ch, 0)
		} else if _, ok := item.(*models.Separator); ok {
			buf.WriteString(`<li class="chapter-item expanded "><li class="spacer"></li>`)
		}
	}
	buf.WriteString(`</ol>`)
	return buf.String()
}

// renderTocJS writes toc.js using the Handlebars template
func (r *HtmlRenderer) renderTocJS(ctx *RenderContext) error {
	tocList := r.generateTocListForJS(ctx.Book)
	rendered, err := renderTocJSWithHbs(ctx, tocList)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(ctx.DestDir, "toc.js"), []byte(rendered), 0644)
}

// renderTocItemForPage renders a TOC item with proper numbering and nesting for toc.html
func (r *HtmlRenderer) renderTocItemForPage(buf *strings.Builder, ch *models.Chapter, level int) {
	if ch.Path == nil {
		return
	}

	path := strings.TrimSuffix(*ch.Path, ".md") + ".html"
	path = strings.ReplaceAll(path, "\\", "/")

	// Calculate section number display
	numStr := ""
	if ch.Number != nil && len(ch.Number.Parts) > 0 {
		parts := make([]string, len(ch.Number.Parts))
		for i, p := range ch.Number.Parts {
			parts[i] = fmt.Sprintf("%d", p)
		}
		numStr = strings.Join(parts, ".") + "."
	}

	fmt.Fprintf(buf, `<li class="chapter-item expanded "><a href="%s" target="_parent"><strong aria-hidden="true">%s</strong> %s</a></li>`, path, numStr, htmlEscape(ch.Name))

	if len(ch.SubItems) > 0 {
		buf.WriteString(`<li><ol class="section">`)
		for _, item := range ch.SubItems {
			if subCh, ok := item.(*models.Chapter); ok {
				r.renderTocItemForPage(buf, subCh, level+1)
			}
		}
		buf.WriteString(`</ol></li>`)
	}
}

// renderTocItemForJS renders a TOC item for toc.js (without target="_parent")
func (r *HtmlRenderer) renderTocItemForJS(buf *strings.Builder, ch *models.Chapter, level int) {
	if ch.Path == nil {
		return
	}

	path := strings.TrimSuffix(*ch.Path, ".md") + ".html"
	path = strings.ReplaceAll(path, "\\", "/")

	// Calculate section number display
	numStr := ""
	if ch.Number != nil && len(ch.Number.Parts) > 0 {
		parts := make([]string, len(ch.Number.Parts))
		for i, p := range ch.Number.Parts {
			parts[i] = fmt.Sprintf("%d", p)
		}
		numStr = strings.Join(parts, ".") + "."
	}

	fmt.Fprintf(buf, `<li class="chapter-item expanded "><a href="%s"><strong aria-hidden="true">%s</strong> %s</a></li>`, path, numStr, htmlEscape(ch.Name))

	if len(ch.SubItems) > 0 {
		buf.WriteString(`<li><ol class="section">`)
		for _, item := range ch.SubItems {
			if subCh, ok := item.(*models.Chapter); ok {
				r.renderTocItemForJS(buf, subCh, level+1)
			}
		}
		buf.WriteString(`</ol></li>`)
	}
}

// renderPrintPage generates a single-page printable version of the book
func (r *HtmlRenderer) renderPrintPage(ctx *RenderContext) error {
	// Collect all chapters in reading order
	chapters := r.collectChapters(ctx.Book)

	var combined strings.Builder
	isFirst := true
	for _, ch := range chapters {
		if !isFirst {
			combined.WriteString(`<div style="break-before: page; page-break-before: always;"></div>`)
		}
		isFirst = false

		htmlContent, _ := r.convertMarkdown(ch.Content)
		combined.WriteString(htmlContent)
	}

	// Render with HBS in print mode
	pd := &pageData{
		Language:           ctx.Config.Book.Language,
		DefaultTheme:       "light",
		PreferredDarkTheme: "navy",
		TextDirection:      "ltr",
		Title:              ctx.Config.Book.Title,
		Description:        ctx.Config.Book.Description,
		FaviconSvg:         true,
		FaviconPng:         true,
		PrintEnable:        true,
		SearchJS:           true,
		SearchEnabled:      true,
		PathToRoot:         "",
		BookTitle:          ctx.Config.Book.Title,
		Content:            raymond.SafeString(combined.String()),
		IsPrint:            true,
	}
	pageHTML, err := renderPageWithHbs(ctx, pd)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(ctx.DestDir, "print.html"), []byte(pageHTML), 0644)
}

// htmlEscape is a minimal HTML escaper for titles
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

// copyAssets copies CSS, JS, fonts, and other assets to the output directory
func (r *HtmlRenderer) copyAssets(ctx *RenderContext) error {
	type asset struct {
		key     string
		src     string // relative under frontend/
		destRel string // pre-hash destination relative path
		data    []byte
		hash    bool
	}
	var assets []asset
	read := func(rel string) ([]byte, error) {
		if ctx.AssetsFS != nil {
			return fs.ReadFile(ctx.AssetsFS, filepath.ToSlash(filepath.Join("frontend", rel)))
		}
		return os.ReadFile(filepath.Join("frontend", rel))
	}
	add := func(key, src, dest string, hash bool) {
		if b, err := read(src); err == nil {
			assets = append(assets, asset{key: key, src: filepath.ToSlash(src), destRel: filepath.ToSlash(dest), data: b, hash: hash})
		}
	}

	// css folder
	add("css/variables.css", "css/variables.css", "css/variables.css", false)
	add("css/general.css", "css/general.css", "css/general.css", false)
	add("css/chrome.css", "css/chrome.css", "css/chrome.css", false)
	add("css/print.css", "css/print.css", "css/print.css", false)
	// root-level css
	add("highlight.css", "css/highlight.css", "highlight.css", false)
	add("tomorrow-night.css", "css/tomorrow-night.css", "tomorrow-night.css", false)
	add("ayu-highlight.css", "css/ayu-highlight.css", "ayu-highlight.css", false)
	// fonts
	add("fonts/fonts.css", "fonts/fonts.css", "fonts/fonts.css", false)
	fontFiles := []string{
		"fonts/open-sans-v17-all-charsets-300.woff2",
		"fonts/open-sans-v17-all-charsets-300italic.woff2",
		"fonts/open-sans-v17-all-charsets-regular.woff2",
		"fonts/open-sans-v17-all-charsets-italic.woff2",
		"fonts/open-sans-v17-all-charsets-600.woff2",
		"fonts/open-sans-v17-all-charsets-600italic.woff2",
		"fonts/open-sans-v17-all-charsets-700.woff2",
		"fonts/open-sans-v17-all-charsets-700italic.woff2",
		"fonts/open-sans-v17-all-charsets-800.woff2",
		"fonts/open-sans-v17-all-charsets-800italic.woff2",
		"fonts/source-code-pro-v11-all-charsets-500.woff2",
		"fonts/OPEN-SANS-LICENSE.txt",
		"fonts/SOURCE-CODE-PRO-LICENSE.txt",
	}
	for _, f := range fontFiles {
		add(f, f, f, false)
	}
	// favicons
	add("favicon.svg", "images/favicon.svg", "favicon.svg", false)
	add("favicon.png", "images/favicon.png", "favicon.png", false)
	// root-level js
	add("book.js", "js/book.js", "book.js", false)
	add("clipboard.min.js", "js/clipboard.min.js", "clipboard.min.js", false)
	add("highlight.js", "js/highlight.js", "highlight.js", false)
	// search js
	add("elasticlunr.min.js", "searcher/elasticlunr.min.js", "elasticlunr.min.js", false)
	add("mark.min.js", "searcher/mark.min.js", "mark.min.js", false)
	add("searcher.js", "searcher/searcher.js", "searcher.js", false)

	// Copy FontAwesome directory
	add("FontAwesome/css/font-awesome.css", "FontAwesome/css/font-awesome.css", "FontAwesome/css/font-awesome.css", false)
	fontAwesomeFonts := []string{
		"FontAwesome/fonts/fontawesome-webfont.eot",
		"FontAwesome/fonts/fontawesome-webfont.svg",
		"FontAwesome/fonts/fontawesome-webfont.ttf",
		"FontAwesome/fonts/fontawesome-webfont.woff",
		"FontAwesome/fonts/fontawesome-webfont.woff2",
	}
	for _, f := range fontAwesomeFonts {
		add(f, f, f, false)
	}

	// Compute mapping
	mapping := map[string]string{}
	hashName := func(name string, data []byte) string {
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		sum := sha256.Sum256(data)
		short := fmt.Sprintf("%x", sum)[:8]
		return base + "-" + short + ext
	}
	for _, a := range assets {
		dest := a.destRel
		if a.hash {
			dir := filepath.Dir(dest)
			hn := hashName(filepath.Base(dest), a.data)
			if dir == "." || dir == "/" {
				dest = hn
			} else {
				dest = filepath.ToSlash(filepath.Join(dir, hn))
			}
		}
		mapping[a.key] = dest
	}

	// Write assets with placeholder rewrite
	re := regexp.MustCompile(`\{\{\s*resource\s+["']([^"']+)["']\s*\}\}`)
	for _, a := range assets {
		content := a.data
		ext := strings.ToLower(filepath.Ext(a.destRel))
		if ext == ".css" || ext == ".js" {
			content = re.ReplaceAllFunc(content, func(m []byte) []byte {
				sub := re.FindSubmatch(m)
				if len(sub) == 2 {
					key := string(sub[1])
					if v, ok := mapping[key]; ok {
						// Convert absolute path to relative path based on current asset location
						assetDir := filepath.Dir(a.destRel)
						relPath, err := filepath.Rel(assetDir, v)
						if err == nil {
							// Normalize to forward slashes for URLs
							return []byte(filepath.ToSlash(relPath))
						}
						return []byte(v)
					}
					return []byte(key)
				}
				return m
			})
		}
		dest := mapping[a.key]
		out := filepath.Join(ctx.DestDir, filepath.FromSlash(dest))
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(out, content, 0o644); err != nil {
			return err
		}
	}

	ctx.ResourceMap = mapping
	return nil
}

// copyNonMarkdown copies all non-Markdown files from the source directory to the destination,
// preserving the directory structure relative to the source root. It skips copying into itself
// if the destination lies within the source tree.
func (r *HtmlRenderer) copyNonMarkdown(ctx *RenderContext) error {
	srcRoot := ctx.SourceDir
	dstRoot := ctx.DestDir

	// Avoid copying the build dir back into itself if misconfigured
	srcClean := filepath.Clean(srcRoot)
	dstClean := filepath.Clean(dstRoot)
	if strings.HasPrefix(dstClean, srcClean) {
		// Best-effort: do nothing to avoid recursion
		return nil
	}

	return filepath.Walk(srcRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		// Skip Markdown files
		nameLower := strings.ToLower(info.Name())
		if strings.HasSuffix(nameLower, ".md") {
			return nil
		}
		// Compute relative path
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return nil
		}
		// Destination path
		dst := filepath.Join(dstRoot, rel)
		// Ensure directory
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0o644)
	})
}

// generateSearchIndex generates a searchindex.js file from the book chapters
func (r *HtmlRenderer) generateSearchIndex(ctx *RenderContext) error {
	// Create elasticlunr index with fields
	idx := search.NewIndex([]string{"title", "body", "breadcrumbs"})

	// Build search index data with breadcrumbs
	docURLs := make([]string, 0)
	docID := 0

	// Recursively process chapters to maintain hierarchy for breadcrumbs
	var processChaptersRecursive func(*models.Chapter, string)
	processChaptersRecursive = func(ch *models.Chapter, parentBreadcrumb string) {
		if ch.Path == nil {
			return
		}

		docPath := strings.TrimSuffix(*ch.Path, ".md") + ".html"
		docPath = strings.ReplaceAll(docPath, "\\", "/")

		// Build breadcrumb for this chapter
		var breadcrumb string
		if parentBreadcrumb == "" {
			breadcrumb = ch.Name
		} else {
			breadcrumb = parentBreadcrumb + " » " + ch.Name
		}

		// Convert markdown to HTML first to extract headings
		htmlContent, headings := r.convertMarkdown(ch.Content)
		plainText := r.stripHTML(htmlContent)

		// Add the main chapter document
		doc := map[string]interface{}{
			"body":        plainText,
			"breadcrumbs": breadcrumb,
			"id":          docID,
			"title":       ch.Name,
		}
		docURLs = append(docURLs, docPath)
		idx.AddDoc(doc)
		docID++

		// Add each heading as a separate searchable document
		for _, h := range headings {
			headingDoc := map[string]interface{}{
				"body":        h.Text,
				"breadcrumbs": breadcrumb + " » " + h.Text,
				"id":          docID,
				"title":       h.Text,
			}
			docURLs = append(docURLs, docPath+"#"+h.ID)
			idx.AddDoc(headingDoc)
			docID++
		}

		// Process sub-chapters recursively
		for _, item := range ch.SubItems {
			if subCh, ok := item.(*models.Chapter); ok {
				processChaptersRecursive(subCh, breadcrumb)
			}
		}
	}

	// Process all top-level chapters
	for _, item := range ctx.Book.Items {
		if ch, ok := item.(*models.Chapter); ok {
			processChaptersRecursive(ch, "")
		}
	}

	// Create search index in geopub format with the built index
	indexMap := idx.ToMap()

	// Build the structure: { doc_urls, index: {...}, results_options, search_options }
	// where index contains: { documentStore, fields, index, lang, pipeline, ref, version }
	searchIndex := map[string]interface{}{
		"doc_urls": docURLs,
		"index":    indexMap, // This now contains all the elasticlunr index fields
		"results_options": map[string]interface{}{
			"limit_results":     30,
			"teaser_word_count": 30,
		},
		"search_options": map[string]interface{}{
			"bool":   "OR",
			"expand": true,
			"fields": map[string]interface{}{
				"title":       map[string]interface{}{"boost": 2},
				"body":        map[string]interface{}{"boost": 1},
				"breadcrumbs": map[string]interface{}{"boost": 1},
			},
		},
	}

	// Marshal to JSON (compact, no indentation for production)
	indexJSON, err := json.Marshal(searchIndex)
	if err != nil {
		return fmt.Errorf("failed to marshal search index: %w", err)
	}

	// Escape backslashes and single quotes for safe embedding into a JS string literal
	jsonStr := string(indexJSON)
	jsonStr = strings.ReplaceAll(jsonStr, "\\", "\\\\")
	jsonStr = strings.ReplaceAll(jsonStr, "'", "\\'")

	// Write searchindex.js in geopub format (index is pre-built, no runtime fallback needed)
	searchIndexPath := filepath.Join(ctx.DestDir, "searchindex.js")
	content := fmt.Sprintf("window.search = Object.assign(window.search, JSON.parse('%s'));", jsonStr)

	if err := os.WriteFile(searchIndexPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write searchindex.js: %w", err)
	}

	// Generate redirect pages (if configured)
	if err := r.generateRedirects(ctx); err != nil {
		return err
	}

	return nil
}

// stripHTML removes HTML tags from text
func (r *HtmlRenderer) stripHTML(content string) string {
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	return htmlRegex.ReplaceAllString(content, "")
}





