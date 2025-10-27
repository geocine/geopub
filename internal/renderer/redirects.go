package renderer

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/aymerick/raymond"
)

// generateRedirects reads [output.html.redirect] mappings from config and emits redirect pages.
// Only emits a redirect file for a base path if there's a mapping without a fragment for that base.
func (r *HtmlRenderer) generateRedirects(ctx *RenderContext) error {
	// Expect map under output.html.redirect
	htmlOut, ok := ctx.Config.Output["html"].(map[string]interface{})
	if !ok {
		return nil
	}
	redir, ok := htmlOut["redirect"].(map[string]interface{})
	if !ok {
		return nil
	}

	type group struct {
		baseTarget string
		fragments  map[string]string // fragment (with leading #) -> target
	}
	groups := map[string]*group{}

	for sk, v := range redir {
		src := sk
		dst, ok2 := v.(string)
		if !ok2 {
			continue
		}
		// Normalize src: trim leading slash
		src = strings.TrimPrefix(src, "/")
		// Split off fragment
		base := src
		frag := ""
		if i := strings.Index(src, "#"); i >= 0 {
			base = src[:i]
			frag = src[i:]
		}
		g := groups[base]
		if g == nil {
			g = &group{fragments: map[string]string{}}
			groups[base] = g
		}
		if frag == "" {
			g.baseTarget = dst
		} else {
			g.fragments[frag] = dst
		}
	}

	// Load redirect.hbs template
	var tmplFS fs.FS
	var base string
	if ctx.AssetsFS != nil {
		tmplFS = ctx.AssetsFS
		base = "frontend/templates/"
	} else {
		tmplFS = os.DirFS(filepath.Join("frontend", "templates"))
		base = ""
	}
	tplBytes, err := fs.ReadFile(tmplFS, base+"redirect.hbs")
	if err != nil {
		return err
	}
	tpl, err := raymond.Parse(string(tplBytes))
	if err != nil {
		return err
	}

	// Emit files
	for srcBase, g := range groups {
		if g.baseTarget == "" { // only fragments present, nothing to emit
			continue
		}
		outPath := filepath.Join(ctx.DestDir, filepath.FromSlash(srcBase))
		// Do not overwrite existing content page
		if _, err := os.Stat(outPath); err == nil {
			continue
		}
		fragJSON, _ := json.Marshal(g.fragments)
		data := map[string]interface{}{
			"url":          g.baseTarget,
			"fragment_map": string(fragJSON),
		}
		out, err := tpl.Exec(data)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(outPath, []byte(out), 0644); err != nil {
			return err
		}
	}
	return nil
}
