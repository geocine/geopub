# Frontmatter Preprocessor

The Frontmatter Preprocessor strips YAML or TOML metadata from the beginning of chapter files before rendering.

## What It Does

Removes frontmatter blocks from your chapter markdown content. This is useful when you want to include metadata in your source files but don't want it displayed in the rendered HTML.

**Supported formats:**
- YAML: `---` ... `---`
- TOML: `+++` ... `+++`

## Example

**Source Markdown** (`chapter.md`):
```markdown
---
title: Introduction
author: Jane Doe
tags: [intro, getting-started]
draft: false
---

# Introduction

This is the actual chapter content that will be rendered.
```

**Rendered Output** (frontmatter stripped):
```html
<h1>Introduction</h1>
<p>This is the actual chapter content that will be rendered.</p>
```

## How to Enable

The Frontmatter Preprocessor is **disabled by default**. To enable it, add this to your `book.toml`:

```toml
[preprocessor.frontmatter]
# Built-in preprocessor - no command needed
# Just include this section to enable it
```

### With Other Preprocessors

You can combine it with other preprocessors using `before` and `after`:

```toml
[preprocessor.frontmatter]
before = ["index"]  # Run frontmatter stripper before index

[preprocessor.my-custom]
after = ["frontmatter"]  # Run after frontmatter is stripped
```

## Features

✅ **Supports both YAML and TOML** frontmatter formats
✅ **Handles nested chapters** recursively
✅ **Safe** - doesn't modify chapters without frontmatter
✅ **Simple** - no configuration needed beyond enabling it
✅ **Built-in** - no external commands or dependencies

## Use Cases

1. **Metadata Management**
   - Store chapter metadata (author, date, tags) in source
   - Keep rendered output clean

2. **Content Workflow**
   - Include SEO hints and metadata for processors
   - Use in combination with other preprocessors

3. **Multi-format Publishing**
   - Same source markdown for multiple outputs
   - Strip metadata only for HTML rendering

## Example Workflow

**Setup:**
```toml
# book.toml
[preprocessor.frontmatter]
# Enable frontmatter stripping

[preprocessor.token-replace]
after = ["frontmatter"]  # Run after metadata is stripped
AUTHOR = "Jane Doe"
```

**Chapter File** (`src/intro.md`):
```markdown
---
written_by: Jane Doe
difficulty: beginner
estimated_time: 5 minutes
---

# Getting Started

This chapter covers the basics...
```

**Result:**
- ✅ Frontmatter removed before rendering
- ✅ Token replacer sees clean content without metadata
- ✅ HTML output shows only chapter content

## Implementation Details

- **Location**: `internal/preprocessor/frontmatter/`
- **Type**: Built-in preprocessor (no external process)
- **Performance**: O(n) where n is content length
- **Memory**: Minimal - regex-based processing

## Testing

Comprehensive tests cover:
- YAML frontmatter stripping
- TOML frontmatter stripping
- Content without frontmatter (unchanged)
- Empty frontmatter
- Multiline YAML values
- Nested chapters
- Edge cases (dashes in content, etc.)

Run tests:
```bash
go test ./internal/preprocessor/frontmatter -v
```

## Limitations

- Frontmatter must be at the **very beginning** of the file
- Only the first frontmatter block is removed
- Frontmatter must be properly formed (matching delimiters)
- Dashes or plus signs in regular content won't be interpreted as frontmatter delimiters

## Security

This preprocessor only reads and transforms chapter content. It:
- ❌ Does NOT execute code
- ❌ Does NOT read files outside the book
- ❌ Does NOT modify other configuration
- ✅ IS safe to use with untrusted content

## Troubleshooting

### Frontmatter not being removed

**Check:**
1. Frontmatter is at the very start of the file (no leading whitespace)
2. Delimiters are on their own lines
3. You have enabled it in `book.toml`

**Example - Correct:**
```markdown
---
key: value
---

# Content
```

**Example - Wrong** (has leading space):
```markdown
 ---
key: value
---

# Content
```

### YAML vs TOML

Use consistent delimiters:
- `---` for YAML (3 dashes)
- `+++` for TOML (3 plus signs)

Don't mix them in the same file.

## References

- [YAML Frontmatter](https://jekyllrb.com/docs/front-matter/)
- [TOML Format](https://toml.io/)
- [GeoPub Preprocessor System](../../../README.md#preprocessors)
