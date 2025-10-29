# GeoPub Preprocessor Examples

This directory contains example preprocessors that demonstrate how to extend GeoPub with custom content transformations.

## Available Preprocessors

### 1. Token Replace (Node.js)
**Directory**: `preprocessor-token-replace/`

A simple preprocessor that replaces tokens in chapter content.

**Features**:
- Replaces `{{TOKEN}}` with configured values
- Works with any token name
- No dependencies required

**Quick Start**:
```toml
[preprocessor.token-replace]
command = "node examples/preprocessors/preprocessor-token-replace/preprocessor.js"
renderers = ["html"]
AUTHOR = "Jane Doe"
VERSION = "1.0.0"
```

**Usage in Markdown**:
```markdown
Written by: {{AUTHOR}}
Version: {{VERSION}}
```

**See**: [`preprocessor-token-replace/README.md`](./preprocessor-token-replace/README.md)

---

### 2. Go Example (Go)
**Directory**: `preprocessor-go/`

A Go preprocessor that calculates reading time for each chapter.

**Features**:
- Demonstrates Go SDK usage
- Adds reading time estimates
- Uses `github.com/geocine/geopub/internal/preprocessor/sdk`

**Quick Start**:
```bash
# Build the binary
go build -o examples/preprocessors/preprocessor-go/geopub-go-example ./examples/preprocessors/preprocessor-go/main.go

# Or use go run directly
```

**Configuration**:
```toml
[preprocessor.go-example]
command = "examples/preprocessors/preprocessor-go/geopub-go-example"
renderers = ["html"]
```

**See**: [`preprocessor-go/README.md`](./preprocessor-go/README.md)

---

## Quick Reference

| Preprocessor | Language | Purpose | Setup |
|--------------|----------|---------|-------|
| token-replace | Node.js | Replace tokens | `npm` (none required) |
| go-example | Go | Reading time | Compile binary |

## How to Use These Examples

### 1. Simple Test (Node.js Token Replace)

```bash
# No setup needed, just add to book.toml:
[preprocessor.token-replace]
command = "node examples/preprocessors/preprocessor-token-replace/preprocessor.js"
AUTHOR = "Your Name"

# Then use in markdown:
# By: {{AUTHOR}}

# Build
geopub build --verbose
```

### 2. Compile Go Example

```bash
cd geopub
go build -o examples/preprocessors/preprocessor-go/geopub-go-example ./examples/preprocessors/preprocessor-go/main.go

# Add to book.toml:
[preprocessor.go-example]
command = "examples/preprocessors/preprocessor-go/geopub-go-example"

# Build
geopub build --verbose
```

---

## Test Book Included

This directory includes a test book (`book.toml`, `src/`) for testing preprocessors:

```bash
# From this directory
geopub build --verbose
```

---

## Creating Your Own Preprocessor

### Using Node.js/JavaScript

See [`preprocessor-token-replace/README.md`](./preprocessor-token-replace/README.md) for:
- JSON protocol details
- How to read stdin/write stdout
- Error handling

### Using Go

See [`preprocessor-go/README.md`](./preprocessor-go/README.md) for:
- SDK usage
- SDK functions
- Building and deployment

### Using Other Languages

Any language that can:
1. Read JSON from stdin
2. Write JSON to stdout
3. Be executed as a command

Examples: Python, Rust, Java, C#, Ruby, etc.

---

## Protocol Overview

All preprocessors must follow the mdBook-compatible JSON protocol:

**Input** (stdin):
```json
{
  "book": {
    "sections": [
      {
        "chapter": {
          "name": "Chapter Name",
          "content": "# Markdown content",
          "path": "ch1.md",
          "sub_items": []
        }
      }
    ]
  },
  "config": {
    "book": {...},
    "build": {...},
    "preprocessor": {...}
  },
  "renderer": "html",
  "version": "0.1"
}
```

**Output** (stdout):
```json
{
  "book": {
    "sections": [
      {
        "chapter": {
          "name": "Chapter Name",
          "content": "# Modified markdown content",
          "path": "ch1.md",
          "sub_items": []
        }
      }
    ]
  },
  "config": {...},
  "renderer": "html",
  "version": "0.1"
}
```

---

## Configuration in book.toml

Each preprocessor can have its own configuration:

```toml
[preprocessor.my-preprocessor]
command = "path/to/preprocessor"           # Required: how to run it
renderers = ["html"]                       # Optional: which renderers to use (default: all)
before = ["other-name"]                    # Optional: run before these preprocessors
after = ["index"]                          # Optional: run after these preprocessors
custom_option = "value"                    # Optional: passed to preprocessor as config
```

---

## Debugging

### Enable Verbose Output
```bash
geopub build --verbose
```

Shows:
- Preprocessor execution order
- Which preprocessors run
- Commands being executed

### Disable External Preprocessors
```bash
geopub build --no-externals
```

Only runs built-in preprocessors (index).

### Manual Protocol Testing

Test a preprocessor manually:

```bash
# Create input
cat > input.json << 'EOF'
{
  "book": {"sections": [{"chapter": {"name": "Test", "content": "{{TOKEN}}", "path": "test.md", "sub_items": []}}]},
  "config": {},
  "renderer": "html",
  "version": "0.1"
}
EOF

# Test Node.js preprocessor
node preprocessor-token-replace/preprocessor.js < input.json

# Test Go preprocessor
./preprocessor-go/geopub-go-example < input.json

# Check if output is valid JSON
cat output.json | jq .
```

---

## Security Notes

**Warning**: Preprocessors run with the same privileges as the `geopub` CLI.

- Only use preprocessors from trusted sources
- Review code before using
- Use `--no-externals` flag for untrusted builds

---

## Troubleshooting

### "Preprocessor not found"
Make sure the command path is correct and the preprocessor is executable.

### "Invalid JSON" error
Preprocessor must output valid JSON. Check:
- No logging to stdout
- Correct JSON format
- No extra characters

### Go build fails
Build from the geopub directory where `go.mod` is located:
```bash
cd geopub
go build -o examples/preprocessors/preprocessor-go/geopub-go-example ./examples/preprocessors/preprocessor-go/main.go
```

### Node.js not found
Install Node.js or use the full path:
```toml
command = "C:\\Program Files\\nodejs\\node.exe examples/preprocessors/preprocessor-token-replace/preprocessor.js"
```

---

## Next Steps

1. **Try token-replace**: Quick test with no setup needed
2. **Build go-example**: See how Go SDK works
3. **Create your own**: Use one as a template
4. **Share**: Contribute examples back!

---

## Resources

- [GeoPub README](../../README.md) - Main documentation
- [Preprocessor Protocol](../../README.md#preprocessor-protocol) - Full protocol spec
- Individual README files in each preprocessor directory

