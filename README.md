# geopub

<p align="center">
  <img src="./geopub.png" alt="geopub logo" width="250" />
</p>

***geopub*** is a utility for building structured documentation sites from Markdown files. It's inspired by [mdBook](https://github.com/rust-lang/mdBook), but is implemented with a Go backend to offer a familiar toolchain for developers who prefer Go for extension and contribution.

## Quick start

```bash
# Build the CLI
go build -o geopub .

# See available commands
geopub
```

## Build an existing book

Run from your book's root (where `book.toml` lives):

```bash
geopub build                 # uses [build.build-dir] from book.toml (default: "book")
geopub build -dest-dir out   # override output directory
geopub serve --open          # serve locally with live reload
```

## Initialize a new book

```bash
geopub init my-first-book              # interactive prompts
geopub init my-first-book --yes        # accept defaults (no prompts)
geopub init \
  --name my-first-book \
  --title "My First Book" \
  --src src \
  --build-dir book \
  --create-missing \
  --yes

cd my-first-book
geopub build
```

Notes:
- The positional directory name (e.g., `my-first-book`) takes precedence over `--name`.
- `--create-missing` auto-creates referenced chapters on build.
- `serve` is supported (host/port flags available); `--open` will launch your browser.

## License

**geopub** is dual-licensed:

* **Core Go code and original files** are licensed under the **MIT License** (see [LICENSE-MIT]).
* **Template and static asset files** (copied or derived from the mdBook project) are licensed under the **Mozilla Public License v2.0** (MPL-2.0), as required by the original license (see [LICENSE-MPL-2.0] for details).

[LICENSE]: ./LICENSE


