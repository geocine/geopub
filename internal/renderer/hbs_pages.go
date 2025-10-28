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
	// We need to pass path_to_root to the resource helper, but since helpers are global
	// and registered once, we use a different approach: use the template's path_to_root variable
	registerCommonHelpers(func(name string) string {
		// Just return the name; the template will prepend path_to_root
		return name
	})

	// Determine template source and register partials
	var indexData []byte
	var err error
	var tmplFS fs.FS
	var base string

	// Check for theme override first
	themeIndexPath := filepath.Join("theme", "index.hbs")
	if data, err := os.ReadFile(themeIndexPath); err == nil {
		// Use custom theme template and replace mdbook- with geopub-
		content := string(data)
		content = strings.ReplaceAll(content, "mdbook-", "geopub-")
		content = strings.ReplaceAll(content, "MDBook", "GeoPub")
		content = strings.ReplaceAll(content, "<mdbook-", "<geopub-")
		content = strings.ReplaceAll(content, "</mdbook-", "</geopub-")

		// Fix resource paths: add path_to_root before resource helper calls for proper relative paths
		// Replace {{ resource "..." }} with {{ path_to_root }}{{ resource "..." }}
		// but only if path_to_root is not already there
		// First, replace existing patterns that already have path_to_root with a marker
		content = strings.ReplaceAll(content, "{{ path_to_root }}{{ resource ", "{{MARKER_path_to_root_resource ")
		content = strings.ReplaceAll(content, "{{path_to_root}}{{ resource ", "{{MARKER_path_to_root_resource ")
		content = strings.ReplaceAll(content, "{{ path_to_root}}{{resource ", "{{MARKER_path_to_root_resource ")
		content = strings.ReplaceAll(content, "{{path_to_root}}{{resource ", "{{MARKER_path_to_root_resource ")
		// Now add path_to_root to all remaining {{ resource calls
		resourceRegex := regexp.MustCompile(`(\{\{\s*resource\s+)`)
		content = resourceRegex.ReplaceAllString(content, "{{ path_to_root }}$1")
		// Restore the marked ones
		content = strings.ReplaceAll(content, "{{MARKER_path_to_root_resource ", "{{ path_to_root }}{{ resource ")

		indexData = []byte(content)
	} else {
		// Use default template
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
	}

	// If default tmplFS not yet set (happens when we didn't use default template), set it now for partials
	if tmplFS == nil {
		if ctx.AssetsFS != nil {
			tmplFS = ctx.AssetsFS
			base = "frontend/templates/"
		} else {
			tmplDir := filepath.Join("frontend", "templates")
			if _, err := os.Stat(tmplDir); err == nil {
				tmplFS = os.DirFS(tmplDir)
				base = ""
			}
		}
	}

	// Helper to read file with theme override support
	readTemplateFile := func(filename string) ([]byte, error) {
		// Try theme directory first
		themePath := filepath.Join("theme", filename)
		if data, err := os.ReadFile(themePath); err == nil {
			// Replace mdbook- prefixes with geopub- for compatibility
			ext := strings.ToLower(filepath.Ext(filename))
			if ext == ".hbs" || ext == ".html" {
				content := string(data)
				content = strings.ReplaceAll(content, "mdbook-", "geopub-")
				content = strings.ReplaceAll(content, "mdBook", "GeoPub")
				content = strings.ReplaceAll(content, "<mdbook-", "<geopub-")
				content = strings.ReplaceAll(content, "</mdbook-", "</geopub-")
				data = []byte(content)
			}
			return data, nil
		}
		// Fall back to default templates
		if tmplFS != nil {
			return fs.ReadFile(tmplFS, base+filename)
		}
		return nil, fmt.Errorf("file not found: %s", filename)
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
	if b, err := readTemplateFile("head.hbs"); err == nil {
		safeRegisterPartial("head", b)
	}

	// Register header partial (used for page header)
	if b, err := readTemplateFile("header.hbs"); err == nil {
		safeRegisterPartial("header", b)
	}

	// Register footer partial
	if b, err := readTemplateFile("footer.hbs"); err == nil {
		safeRegisterPartial("footer", b)
	}

	// Parse and render template
	tpl, err := raymond.Parse(string(indexData))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Convert struct to map for proper field name resolution in template
	dataMap := map[string]interface{}{
		"language":                  data.Language,
		"default_theme":             data.DefaultTheme,
		"preferred_dark_theme":      data.PreferredDarkTheme,
		"text_direction":            data.TextDirection,
		"title":                     data.Title,
		"base_url":                  data.BaseUrl,
		"description":               data.Description,
		"favicon_svg":               data.FaviconSvg,
		"favicon_png":               data.FaviconPng,
		"copy_fonts":                data.CopyFonts,
		"print_enable":              data.PrintEnable,
		"additional_css":            data.AdditionalCSS,
		"additional_js":             data.AdditionalJS,
		"mathjax_support":           data.MathJaxSupport,
		"search_js":                 data.SearchJS,
		"search_enabled":            data.SearchEnabled,
		"path_to_root":              data.PathToRoot,
		"book_title":                data.BookTitle,
		"previous":                  data.Previous,
		"next":                      data.Next,
		"live_reload_endpoint":      data.LiveReloadEndpoint,
		"content":                   data.Content,
		"is_print":                  data.IsPrint,
		"fragment_map":              data.FragmentMap,
		"git_repository_url":        data.GitRepositoryUrl,
		"git_repository_edit_url":   data.GitRepositoryEditUrl,
		"git_repository_icon":       data.GitRepositoryIcon,
		"git_repository_icon_class": data.GitRepositoryIconClass,
	}

	result, err := tpl.Exec(dataMap)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}
	return result, nil
}
