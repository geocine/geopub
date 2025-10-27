package renderer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/aymerick/raymond"
)

// pageData is the context passed to the Handlebars templates for pages
type pageData struct {
	Language           string                 `json:"language"`
	DefaultTheme       string                 `json:"default_theme"`
	PreferredDarkTheme string                 `json:"preferred_dark_theme"`
	TextDirection      string                 `json:"text_direction"`
	Title              string                 `json:"title"`
	BaseUrl            string                 `json:"base_url"`
	Description        string                 `json:"description"`
	FaviconSvg         bool                   `json:"favicon_svg"`
	FaviconPng         bool                   `json:"favicon_png"`
	PrintEnable        bool                   `json:"print_enable"`
	AdditionalCSS      []string               `json:"additional_css"`
	AdditionalJS       []string               `json:"additional_js"`
	MathJaxSupport     bool                   `json:"mathjax_support"`
	SearchJS           bool                   `json:"search_js"`
	SearchEnabled      bool                   `json:"search_enabled"`
	PathToRoot         string                 `json:"path_to_root"`
	BookTitle          string                 `json:"book_title"`
	Previous           *struct{ Link string } `json:"previous"`
	Next               *struct{ Link string } `json:"next"`
	LiveReloadEndpoint string                 `json:"live_reload_endpoint"`
	Content            raymond.SafeString     `json:"content"`
	IsPrint            bool                   `json:"is_print"`
	FragmentMap        string                 `json:"fragment_map"`
}

// registerCommonHelpers registers helpers used by the templates.
func registerCommonHelpers(resolve func(string) string) {
	// Register helpers - they're global in raymond, so we only need to register once
	// The first time this is called per rendering session, they will be registered
	// If they're already registered, we catch the panic and continue

	safeRegisterHelper := func(name string, helper interface{}) {
		defer func() {
			if r := recover(); r != nil {
				// Helper already registered, that's OK
			}
		}()
		raymond.RegisterHelper(name, helper)
	}

	// equality test
	safeRegisterHelper("eq", func(a interface{}, b interface{}) bool {
		return fmt.Sprint(a) == fmt.Sprint(b)
	})
	// resource passthrough (no hashing yet)
	safeRegisterHelper("resource", func(name string) string {
		if resolve == nil {
			return name
		}
		return resolve(name)
	})

	// FontAwesome icon helper
	// Icon map for name mapping
	iconMap := map[string]string{
		"bars":              "bars",
		"paintbrush":        "paint-brush",
		"magnifying-glass":  "search",
		"print":             "print",
		"pencil":            "pencil",
		"spinner":           "spinner",
		"angle-right":       "angle-right",
		"angle-left":        "angle-left",
		"eye":               "eye",
		"eye-slash":         "eye-slash",
		"copy":              "copy",
		"play":              "play",
		"clock-rotate-left": "history",
	}

	// Helper for fa with just icon name: {{fa "bars"}}
	safeRegisterHelper("fa", func(iconName string, options *raymond.Options) raymond.SafeString {
		faName := iconName
		if mapped, ok := iconMap[iconName]; ok {
			faName = mapped
		}
		return raymond.SafeString("<i class=\"fa fa-" + faName + "\" aria-hidden=\"true\"></i>")
	})
}

// renderPageWithHbs renders a page using Handlebars template engine
func renderPageWithHbs(ctx *RenderContext, data *pageData) (string, error) {
	// Register helpers before rendering
	registerCommonHelpers(func(name string) string {
		// This is just a no-op since we're not using asset fingerprinting
		return name
	})

	// Determine template source and register partials
	var indexData []byte
	var err error
	var tmplFS fs.FS
	var base string

	if ctx.AssetsFS != nil {
		// Use embedded FS
		tmplFS = ctx.AssetsFS
		base = "frontend/templates/"
		indexData, err = fs.ReadFile(ctx.AssetsFS, "frontend/templates/index.hbs")
	} else {
		// Fallback to disk
		tmplDir := filepath.Join("frontend", "templates")
		if _, err := os.Stat(tmplDir); err != nil {
			return "", fmt.Errorf("templates directory not found at %s", tmplDir)
		}
		tmplFS = os.DirFS(tmplDir)
		base = ""
		indexData, err = os.ReadFile(filepath.Join("frontend", "templates", "index.hbs"))
	}

	if err != nil {
		return "", fmt.Errorf("failed to read index.hbs: %w", err)
	}

	// Register partials
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

	// Register head partial (used for head section)
	if b, err := fs.ReadFile(tmplFS, base+"head.hbs"); err == nil {
		safeRegisterPartial("head", b)
	}

	// Register header partial (used for page header)
	if b, err := fs.ReadFile(tmplFS, base+"header.hbs"); err == nil {
		safeRegisterPartial("header", b)
	}

	// Register footer partial
	if b, err := fs.ReadFile(tmplFS, base+"footer.hbs"); err == nil {
		safeRegisterPartial("footer", b)
	}

	// Parse and render template
	tpl, err := raymond.Parse(string(indexData))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Convert struct to map for proper field name resolution in template
	dataMap := map[string]interface{}{
		"language":             data.Language,
		"default_theme":        data.DefaultTheme,
		"preferred_dark_theme": data.PreferredDarkTheme,
		"text_direction":       data.TextDirection,
		"title":                data.Title,
		"base_url":             data.BaseUrl,
		"description":          data.Description,
		"favicon_svg":          data.FaviconSvg,
		"favicon_png":          data.FaviconPng,
		"print_enable":         data.PrintEnable,
		"additional_css":       data.AdditionalCSS,
		"additional_js":        data.AdditionalJS,
		"mathjax_support":      data.MathJaxSupport,
		"search_js":            data.SearchJS,
		"search_enabled":       data.SearchEnabled,
		"path_to_root":         data.PathToRoot,
		"book_title":           data.BookTitle,
		"previous":             data.Previous,
		"next":                 data.Next,
		"live_reload_endpoint": data.LiveReloadEndpoint,
		"content":              data.Content,
		"is_print":             data.IsPrint,
		"fragment_map":         data.FragmentMap,
	}

	result, err := tpl.Exec(dataMap)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}
	return result, nil
}
