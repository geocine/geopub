# geopub

<p align="center">
  <img src="./geopub.png" alt="geopub logo" width="250" />
</p>

***geopub*** is a tool for building and publishing websites from Markdown files. It’s inspired by [mdBook](https://github.com/rust-lang/mdBook) but written in Go, with a small JavaScript frontend.

## Disclaimer

This project is for personal use. I created it to customize mdBook and because I mostly work with Go and JavaScript.

It’s not a direct replacement for mdBook. Some Rust-based features aren’t included, and I’ve modified or added features to suit my needs.

**Use at your own risk.** Some behavior may differ from mdBook.


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

All the code in this repository is released under the ***Mozilla Public License v2.0***, for more information take a look at the [LICENSE] file.

[LICENSE]: ./LICENSE


