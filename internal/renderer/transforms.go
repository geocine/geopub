package renderer

import (
	"fmt"
	htmlutil "html"
	"regexp"
	"strings"
)

// transformAdmonitions converts [!TAG] blockquotes to mdBook-like admonitions with icons.
func transformAdmonitions(html string) string {
	re := regexp.MustCompile(`(?is)<blockquote>\s*<p>\s*\[!([A-Z]+)\]\s*(.*?)</p>(.*?)</blockquote>`)
	icon := func(tag string) (string, string, string, bool) {
		switch strings.ToUpper(tag) {
		case "NOTE":
			return "note", "Note", `<svg viewbox="0 0 16 16" width="18" height="18"><path d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8Zm8-6.5a6.5 6.5 0 1 0 0 13 6.5 6.5 0 0 0 0-13ZM6.5 7.75A.75.75 0 0 1 7.25 7h1a.75.75 0 0 1 .75.75v2.75h.25a.75.75 0 0 1 0 1.5h-2a.75.75 0 0 1 0-1.5h.25v-2h-.25a.75.75 0 0 1-.75-.75ZM8 6a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path></svg>`, true
		case "TIP":
			return "tip", "Tip", `<svg viewbox="0 0 16 16" width="18" height="18"><path d="M8 1.5c-2.363 0-4 1.69-4 3.75 0 .984.424 1.625.984 2.304l.214.253c.223.264.47.556.673.848.284.411.537.896.621 1.49a.75.75 0 0 1-1.484.211c-.04-.282-.163-.547-.37-.847a8.456 8.456 0 0 0-.542-.68c-.084-.1-.173-.205-.268-.32C3.201 7.75 2.5 6.766 2.5 5.25 2.5 2.31 4.863 0 8 0s5.5 2.31 5.5 5.25c0 1.516-.701 2.5-1.328 3.259-.095.115-.184.22-.268.319-.207.245-.383.453-.541.681-.208.3-.33.565-.37.847a.751.751 0 0 1-1.485-.212c.084-.593.337-1.078.621-1.489.203-.292.45-.584.673-.848.075-.088.147-.173.213-.253.561-.679.985-1.32.985-2.304 0-2.06-1.637-3.75-4-3.75ZM5.75 12h4.5a.75.75 0 0 1 0 1.5h-4.5a.75.75 0 0 1 0-1.5ZM6 15.25a.75.75 0 0 1 .75-.75h2.5a.75.75 0 0 1 0 1.5h-2.5a.75.75 0 0 1-.75-.75Z"></path></svg>`, true
		case "IMPORTANT":
			return "important", "Important", `<svg viewbox="0 0 16 16" width="18" height="18"><path d="M0 1.75C0 .784.784 0 1.75 0h12.5C15.216 0 16 .784 16 1.75v9.5A1.75 1.75 0 0 1 14.25 13H8.06l-2.573 2.573A1.458 1.458 0 0 1 3 14.543V13H1.75A1.75 1.75 0 0 1 0 11.25Zm1.75-.25a.25.25 0 0 0-.25.25v9.5c0 .138.112.25.25.25h2a.75.75 0 0 1 .75.75v2.19l2.72-2.72a.749.749 0 0 1 .53-.22h6.5a.25.25 0 0 0 .25-.25v-9.5a.25.25 0 0 0-.25-.25Zm7 2.25v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 9a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"></path></svg>`, true
		case "WARNING":
			return "warning", "Warning", `<svg viewbox="0 0 16 16" width="18" height="18"><path d="M6.457 1.047c.659-1.234 2.427-1.234 3.086 0l6.082 11.378A1.75 1.75 0 0 1 14.082 15H1.918a1.75 1.75 0 0 1-1.543-2.575Zm1.763.707a.25.25 0 0 0-.44 0L1.698 13.132a.25.25 0 0 0 .22.368h12.164a.25.25 0 0 0 .22-.368Zm.53 3.996v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 11a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"></path></svg>`, true
		case "CAUTION":
			return "caution", "Caution", `<svg viewbox="0 0 16 16" width="18" height="18"><path d="M4.47.22A.749.749 0 0 1 5 0h6c.199 0 .389.079.53.22l4.25 4.25c.141.14.22.331.22.53v6a.749.749 0 0 1-.22.53l-4.25 4.25A.749.749 0 0 1 11 16H5a.749.749 0 0 1-.53-.22L.22 11.53A.749.749 0 0 1 0 11V5c0-.199.079-.389.22-.53Zm.84 1.28L1.5 5.31v5.38l3.81 3.81h5.38l3.81-3.81V5.31L10.69 1.5ZM8 4a.75.75 0 0 1 .75.75v3.5a.75.75 0 0 1-1.5 0v-3.5A.75.75 0 0 1 8 4Zm0 8a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path></svg>`, true
		}
		return "", "", "", false
	}
	return re.ReplaceAllStringFunc(html, func(m string) string {
		parts := re.FindStringSubmatch(m)
		if len(parts) < 4 {
			return m
		}
		tag := parts[1]
		first := parts[2]
		rest := parts[3]
		cls, title, svg, ok := icon(tag)
		if !ok {
			return m
		}
		var sb strings.Builder
		sb.WriteString(`<blockquote class="blockquote-tag blockquote-tag-` + cls + `">`)
		sb.WriteString(`<p class="blockquote-tag-title">` + svg + title + `</p>`)
		if strings.TrimSpace(first) != "" {
			sb.WriteString(`<p>` + first + `</p>`)
		}
		sb.WriteString(rest)
		sb.WriteString(`</blockquote>`)
		return sb.String()
	})
}

// transformImages normalizes <img> tags: cleans alt, ensures attribute order src, title (if any), alt.
func transformImages(html string) string {
	reImg := regexp.MustCompile(`(?is)<img[^>]*>`)
	reBR := regexp.MustCompile(`(?is)<br\s*/?>`)
	return reImg.ReplaceAllStringFunc(html, func(tag string) string {
		start := strings.Index(strings.ToLower(tag), `alt="`)
		if start == -1 {
			return tag
		}
		start += len(`alt="`)
		i := start
		end := -1
		for i < len(tag) {
			if tag[i] == '"' {
				k := i + 1
				for k < len(tag) && (tag[k] == ' ' || tag[k] == '\t' || tag[k] == '\n' || tag[k] == '\r') {
					k++
				}
				if k >= len(tag) || tag[k] == '>' || tag[k] == '/' || ((tag[k] >= 'A' && tag[k] <= 'Z') || (tag[k] >= 'a' && tag[k] <= 'z') || tag[k] == '_' || tag[k] == ':') {
					end = i
					break
				}
			}
			i++
		}
		if end == -1 {
			return tag
		}
		val := tag[start:end]
		val = reBR.ReplaceAllString(val, " ")
		val = strings.ReplaceAll(val, "&quot;alt&quot;", "“alt”")
		val = strings.ReplaceAll(val, "---", "—")
		val = regexp.MustCompile(`\s+`).ReplaceAllString(val, " ")
		reSrc := regexp.MustCompile(`(?is)src=\"([^\"]*)\"`)
		reTitle := regexp.MustCompile(`(?is)title=\"([^\"]*)\"`)
		src := ""
		if m := reSrc.FindStringSubmatch(tag); len(m) >= 2 {
			src = m[1]
		}
		title := ""
		if m := reTitle.FindStringSubmatch(tag); len(m) >= 2 {
			title = m[1]
		}
		var b strings.Builder
		b.WriteString("<img ")
		if src != "" {
			b.WriteString(`src="` + src + `"`)
		}
		if title != "" {
			b.WriteString(" ")
			b.WriteString(`title="` + title + `"`)
		}
		b.WriteString(" ")
		b.WriteString(`alt="` + val + `"`)
		b.WriteString(">")
		return b.String()
	})
}

// transformFootnotes converts Goldmark footnote HTML to mdBook-like structure.
func transformFootnotes(html string, numToLabel map[string]string) string {
	reSup := regexp.MustCompile(`(?is)<sup id=\"[^\"]*\">\s*<a href=\"#fn:([^\"]+)\"[^>]*>([^<]+)</a>\s*</sup>`)
	counts := map[string]int{}
	html = reSup.ReplaceAllStringFunc(html, func(m string) string {
		sub := reSup.FindStringSubmatch(m)
		if len(sub) < 3 {
			return m
		}
		num := sub[1]
		text := sub[2]
		label := num
		if l, ok := numToLabel[num]; ok {
			label = l
		}
		counts[label]++
		escLabel := htmlEscape(label)
		id := fmt.Sprintf("fr-%s-%d", escLabel, counts[label])
		return fmt.Sprintf(`<sup class="footnote-reference" id="%s"><a href="#footnote-%s">%s</a></sup>`, id, escLabel, text)
	})
	reSection := regexp.MustCompile(`(?is)<div class=\"footnotes\"[^>]*>.*?<ol>(.*?)</ol>\s*</div>`)
	html = reSection.ReplaceAllStringFunc(html, func(m string) string {
		inner := reSection.FindStringSubmatch(m)
		if len(inner) < 2 {
			return m
		}
		list := inner[1]
		list = strings.ReplaceAll(list, "&#160;", " ")
		reId := regexp.MustCompile(`(?is)<li id=\"fn:([^\"]+)\"`)
		list = reId.ReplaceAllStringFunc(list, func(s string) string {
			x := reId.FindStringSubmatch(s)
			if len(x) < 2 {
				return s
			}
			num := x[1]
			label := num
			if l, ok := numToLabel[num]; ok {
				label = l
			}
			return fmt.Sprintf(`<li id="footnote-%s"`, htmlEscape(label))
		})
		backCounts := map[string]int{}
		reBack := regexp.MustCompile(`(?is)<a href=\"#fnref[^:\"]*:?([^\"]*)\"[^>]*>.*?</a>`)
		list = reBack.ReplaceAllStringFunc(list, func(a string) string {
			x := reBack.FindStringSubmatch(a)
			if len(x) < 2 {
				return a
			}
			num := x[1]
			label := num
			if l, ok := numToLabel[num]; ok {
				label = l
			}
			backCounts[label]++
			text := "↩"
			if backCounts[label] > 1 {
				text = fmt.Sprintf("↩%d", backCounts[label])
			}
			return fmt.Sprintf(`<a href="#fr-%s-%d">%s</a>`, htmlEscape(label), backCounts[label], text)
		})
		return fmt.Sprintf(`<hr><ol class="footnote-definition">%s</ol>`, list)
	})
	return html
}

// transformDefinitionLists wraps dt terms with anchors and assigns unique ids; merges consecutive terms.
func transformDefinitionLists(html string) string {
	reDL := regexp.MustCompile(`(?is)<dl>(.*?)</dl>`)
	reNode := regexp.MustCompile(`(?is)(<dt>.*?</dt>|<dd>.*?</dd>)`)
	reDTInner := regexp.MustCompile(`(?is)^<dt>(.*?)</dt>$`)
	reBR := regexp.MustCompile(`(?is)<br\s*/?>`)
	reStrip := regexp.MustCompile(`(?is)<[^>]+>`)
	used := map[string]int{}
	buildDT := func(content string) string {
		// For slugging, remove <br> entirely to join terms
		slugBase := reBR.ReplaceAllString(content, "")
		plain := htmlutil.UnescapeString(reStrip.ReplaceAllString(slugBase, ""))
		id := slugify(plain)
		if id == "" {
			id = "term"
		}
		if cnt, ok := used[id]; ok {
			cnt++
			used[id] = cnt
			id = fmt.Sprintf("%s-%d", id, cnt)
		} else {
			used[id] = 0
		}
		return fmt.Sprintf(`<dt id="%s"><a class="header" href="#%s">%s</a></dt>`, id, id, content)
	}
	return reDL.ReplaceAllStringFunc(html, func(block string) string {
		inner := reDL.FindStringSubmatch(block)
		if len(inner) < 2 {
			return block
		}
		nodes := reNode.FindAllString(inner[1], -1)
		if len(nodes) == 0 {
			return block
		}
		anyDDHasP := false
		for _, n := range nodes {
			if strings.HasPrefix(strings.ToLower(n), "<dd>") {
				ddInner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(n, "<dd>"), "</dd>"))
				if strings.HasPrefix(strings.ToLower(ddInner), "<p>") {
					anyDDHasP = true
					break
				}
			}
		}
		var sb strings.Builder
		sb.WriteString("<dl>")
		for i := 0; i < len(nodes); {
			node := nodes[i]
			if strings.HasPrefix(strings.ToLower(node), "<dt>") {
				j := i
				var parts []string
				for j < len(nodes) && strings.HasPrefix(strings.ToLower(nodes[j]), "<dt>") {
					m := reDTInner.FindStringSubmatch(nodes[j])
					if len(m) >= 2 {
						parts = append(parts, m[1])
					}
					j++
				}
				if len(parts) <= 1 {
					sb.WriteString(buildDT(parts[0]))
				} else {
					var disp strings.Builder
					for k := 0; k < len(parts); k++ {
						t := htmlutil.UnescapeString(reStrip.ReplaceAllString(parts[k], ""))
						t = strings.TrimSpace(t)
						endsBS := strings.HasSuffix(t, "\\")
						if endsBS {
							t = strings.TrimSuffix(t, "\\")
						}
						disp.WriteString(htmlEscape(t))
						if k < len(parts)-1 {
							if endsBS {
								disp.WriteString("<br>")
							} else {
								disp.WriteString(" ")
							}
						}
					}
					sb.WriteString(buildDT(disp.String()))
				}
				i = j
				continue
			}
			ddInner := node
			if strings.HasPrefix(strings.ToLower(node), "<dd>") {
				inner := strings.TrimSuffix(strings.TrimPrefix(node, "<dd>"), "</dd>")
				if !strings.Contains(inner, "<") {
					txt := htmlutil.UnescapeString(inner)
					txt = strings.ReplaceAll(txt, "&#160;", " ")
					parts := regexp.MustCompile(`\s+:\s+`).Split(strings.TrimSpace(txt), -1)
					if len(parts) >= 2 {
						var buildNested func([]string) string
						buildNested = func(items []string) string {
							if len(items) == 0 {
								return ""
							}
							if len(items) == 1 {
								return htmlEscape(strings.TrimSpace(items[0]))
							}
							var nb strings.Builder
							nb.WriteString("<dl>")
							nb.WriteString(buildDT(htmlEscape(strings.TrimSpace(items[0]))))
							nb.WriteString("<dd>")
							if len(items) == 2 {
								nb.WriteString(htmlEscape(strings.TrimSpace(items[1])))
							} else {
								nb.WriteString(buildNested(items[1:]))
							}
							nb.WriteString("</dd></dl>")
							return nb.String()
						}
						ddInner = "<dd>" + buildNested(parts) + "</dd>"
					} else if strings.TrimSpace(inner) != "" && anyDDHasP {
						ddInner = "<dd><p>" + strings.TrimSpace(inner) + "</p></dd>"
					}
				}
			}
			sb.WriteString(ddInner)
			i++
		}
		sb.WriteString("</dl>")
		return sb.String()
	})
}

// transformLinkTermParagraphsToDL turns paragraphs like <p><a>term</a> : def</p> into definition lists.
func transformLinkTermParagraphsToDL(html string) string {
	re := regexp.MustCompile(`(?is)<p>\s*(<a [^>]+>.*?</a>)\s*:\s*(.*?)</p>`)
	return re.ReplaceAllString(html, `<dl><dt>$1</dt><dd>$2</dd></dl>`)
}

// transformLinksMdToHtml converts href attribute values ending in .md to .html (preserving fragments and query).
func transformLinksMdToHtml(html string) string {
	re := regexp.MustCompile(`href=\"([^\"]+)\.md([^\"]*)\"`)
	return re.ReplaceAllString(html, `href="$1.html$2"`)
}

// transformCollapseBrWhitespace collapses whitespace immediately following <br> tags.
func transformCollapseBrWhitespace(html string) string {
	re := regexp.MustCompile(`(?is)<br>\s+`)
	return re.ReplaceAllString(html, "<br>")
}
