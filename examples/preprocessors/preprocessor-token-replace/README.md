# GeoPub Token Replace Preprocessor Example

This is a simple example preprocessor written in Node.js that demonstrates how to develop external preprocessors for GeoPub.

## What it does

The token replace preprocessor reads chapter content and replaces tokens (e.g., `{{AUTHOR_NAME}}`) with values specified in `book.toml`.

## Installation

1. Copy this directory to your book:
```bash
cp -r /path/to/geopub/examples/preprocessor-token-replace /path/to/your/book/preprocessors/
```

2. Install dependencies (none for this simple example):
```bash
cd preprocessors/token-replace
npm install  # No dependencies, but good practice
```

3. Make the preprocessor executable:
```bash
chmod +x preprocessor.js
```

## Configuration

Add the preprocessor to your `book.toml`:

```toml
[preprocessor.token-replace]
command = "node preprocessors/token-replace/preprocessor.js"
renderers = ["html"]  # Only run for HTML rendering

# Define tokens to replace
AUTHOR = "Jane Doe"
VERSION = "1.0.0"
BUILD_DATE = "2025-10-29"
```

## Usage

In your markdown files, use tokens like this:

```markdown
# My Book

Written by: {{AUTHOR}}
Version: {{VERSION}}
Built: {{BUILD_DATE}}
```

When you run `geopub build`, the preprocessor will:
1. Receive the book structure and configuration as JSON via stdin
2. Replace all tokens in chapter content
3. Write the modified book back to stdout
4. GeoPub applies those changes before rendering

## How it works

The preprocessor implements the mdBook-compatible preprocessor protocol:

1. **Input**: JSON from stdin containing:
   - Book structure (chapters, content, etc.)
   - Configuration (including preprocessor-specific config)
   - Renderer name ("html")

2. **Processing**: Replace tokens in chapter content

3. **Output**: Modified JSON to stdout

## Architecture

```
┌─────────────────┐
│ GeoPub CLI      │
└────────┬────────┘
         │ JSON (book, config, renderer)
         │
         ├────────────────────────────────────┐
         │                                    │
         v                                    v
    ┌──────────┐                      ┌─────────────────────┐
    │ stdin    │                      │ preprocessor.js     │
    └──────────┘                      │                     │
                                      │ 1. Parse JSON       │
                                      │ 2. Replace tokens   │
                                      │ 3. Output JSON      │
                                      │                     │
                                      └──────────┬──────────┘
                                                 │ JSON (modified book)
                                                 v
                                            ┌──────────┐
                                            │ stdout   │
                                            └──────────┘
```

## Extending the example

To add more functionality:

1. **Use a template engine** instead of simple string replacement:
```javascript
const mustache = require('mustache');
chapter.content = mustache.render(chapter.content, config);
```

2. **Add date/time functions**:
```javascript
const config = {
  YEAR: new Date().getFullYear(),
  MONTH: new Date().toLocaleString('default', { month: 'long' }),
};
```

3. **Use external data sources**:
```javascript
const config = JSON.parse(fs.readFileSync('metadata.json', 'utf8'));
```

4. **Transform content** in other ways:
- Add reading time estimates
- Insert ads or notices
- Generate table of contents
- Add watermarks or metadata

## Error handling

If the preprocessor encounters an error:
1. Write error message to stderr
2. Exit with non-zero code
3. GeoPub will report the error and halt the build

Example:
```javascript
if (!context.book || !context.book.sections) {
  console.error('Invalid preprocessor context');
  process.exit(1);
}
```

## Security

⚠️ **Warning**: External preprocessors run with the same privileges as the `geopub` CLI. Be cautious about:
- Installing preprocessors from untrusted sources
- What configuration values you pass to them
- What files they have access to

Use `geopub build --no-externals` to run without external preprocessors.

## See also

- [Go preprocessor SDK](../preprocessor-go/)
- [GeoPub Documentation](../../geopub/README.md)
