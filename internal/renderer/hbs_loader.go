package renderer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/aymerick/raymond"
)

// renderTocHTMLWithHbs renders toc.html using the geopub Handlebars template and the provided
// prebuilt toc list markup. It registers minimal helpers and partials to satisfy the template.
func renderTocHTMLWithHbs(ctx *RenderContext, tocListHTML string) (string, error) {
	// Locate templates directory from embedded FS if available
	var tmplFS fs.FS
	var base string
	if ctx.AssetsFS != nil {
		tmplFS = ctx.AssetsFS
		base = "frontend/templates/"
	} else {
		// Fallback to disk
		tmplDir := filepath.Join("frontend", "templates")
		if _, err := os.Stat(tmplDir); err != nil {
			return "", fmt.Errorf("templates directory not found at %s", tmplDir)
		}
		tmplFS = os.DirFS(tmplDir)
		base = ""
	}

	// Register partials used by toc.renderer.hbs
	safeRegisterPartial := func(name string, content []byte) {
		defer func() {
			if r := recover(); r != nil {
				// Partial already registered, that's OK
			}
		}()
		if len(content) > 0 {
			if tmpl, err := raymond.Parse(string(content)); err == nil {
				raymond.RegisterPartialTemplate(name, tmpl)
			}
		}
	}

	if b, err := fs.ReadFile(tmplFS, base+"head.hbs"); err == nil {
		safeRegisterPartial("head", b)
	}

	// Register helpers using ctx.ResourceMap
	safeRegisterHelper := func(name string, helper interface{}) {
		defer func() {
			if r := recover(); r != nil {
				// Helper already registered, that's OK
			}
		}()
		raymond.RegisterHelper(name, helper)
	}

	safeRegisterHelper("resource", func(name string) string {
		if ctx != nil && ctx.ResourceMap != nil {
			if v, ok := ctx.ResourceMap[name]; ok {
				return v
			}
		}
		return name
	})

	// The template expects a block helper named 'toc' that returns the list markup
	safeRegisterHelper("toc", func(options *raymond.Options) raymond.SafeString {
		return raymond.SafeString(tocListHTML)
	})

	// Load toc.renderer.hbs
	layout, err := fs.ReadFile(tmplFS, base+"toc.html.hbs")
	if err != nil {
		return "", fmt.Errorf("failed to read toc.html.hbs: %w", err)
	}

	// Build minimal context
	language := ctx.Config.Book.Language
	if language == "" {
		language = "en"
	}

	data := map[string]interface{}{
		"language":       language,
		"default_theme":  "light",
		"text_direction": "ltr",
		"print_enable":   true,
		"additional_css": []string{},
		"base_url":       "",
	}

	out, err := raymond.Render(string(layout), data)
	if err != nil {
		return "", fmt.Errorf("failed to render toc.renderer.hbs: %w", err)
	}
	return out, nil
}

// renderTocJSWithHbs renders toc.js using the Handlebars template and the provided TOC list markup
func renderTocJSWithHbs(ctx *RenderContext, tocListHTML string) (string, error) {
	// Locate templates directory from embedded FS if available
	var tmplFS fs.FS
	var base string
	if ctx.AssetsFS != nil {
		tmplFS = ctx.AssetsFS
		base = "frontend/templates/"
	} else {
		// Fallback to disk
		tmplDir := filepath.Join("frontend", "templates")
		if _, err := os.Stat(tmplDir); err != nil {
			return "", fmt.Errorf("templates directory not found at %s", tmplDir)
		}
		tmplFS = os.DirFS(tmplDir)
		base = ""
	}

	// Register helpers
	safeRegisterHelper := func(name string, helper interface{}) {
		defer func() {
			if r := recover(); r != nil {
				// Helper already registered, that's OK
			}
		}()
		raymond.RegisterHelper(name, helper)
	}

	// The template expects a block helper named 'toc' that returns the list markup
	safeRegisterHelper("toc", func(options *raymond.Options) raymond.SafeString {
		return raymond.SafeString(tocListHTML)
	})

	// Load toc.js.hbs template
	layout, err := fs.ReadFile(tmplFS, base+"toc.js.hbs")
	if err != nil {
		return "", fmt.Errorf("failed to read toc.js.hbs: %w", err)
	}

	// Build context - toc.js doesn't need much context since it's mostly hardcoded logic
	data := map[string]interface{}{
		"sidebar_header_nav": ctx.Config.GetBool("output.html.sidebar-header-nav", false),
	}

	out, err := raymond.Render(string(layout), data)
	if err != nil {
		return "", fmt.Errorf("failed to render toc.js.hbs: %w", err)
	}
	return out, nil
}
