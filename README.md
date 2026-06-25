# comment-formatter

## Reformat my go comments

A small, dependency-free CLI that reformats the **comments** in Go source files —
the way `gofmt` reformats code. It rewraps prose, splits sentences onto their own
lines, enforces end-of-block punctuation, and breaks up dense paragraphs, while
leaving code blocks, lists, headings, directives and trailing comments alone.

It is **syntax-only** (parses with `go/parser`, no type checking, no build), and
every file it changes is run back through `go/format.Source`, so output is always
valid, canonical Go. The transform is **idempotent**: running it twice changes
nothing.

## Install

```sh
go install github.com/fredbi/comment-formatter/cmd/comment-formatter@latest
```

## Usage

Like `gofmt`: with no paths it reads stdin and writes stdout; with paths it
processes each `.go` file (directories are walked recursively).

```sh
comment-formatter file.go            # print reformatted source to stdout
comment-formatter -w ./...           # rewrite files in place
comment-formatter -l ./pkg           # list files whose comments would change
comment-formatter -width 100 -w .    # use a 100-column target width
cat file.go | comment-formatter      # filter mode
```

### Flags

| Flag         | Default | Meaning                                                       |
|--------------|---------|---------------------------------------------------------------|
| `-w`         | off     | write the result back to the source file instead of stdout    |
| `-l`         | off     | list files whose formatting differs (no output written)       |
| `-width`     | `80`    | target max line width, counting the leading indent + `// `     |
| `-max-lines` | `4`     | dense-prose threshold: paragraphs longer than this are broken |

## What it does

Each reformatted comment's body is split into blocks (separated by blank lines).
**Prose** blocks are reformatted; **code** (indented), **list** (`- ` / `1. `) and
**heading** (`# `) blocks are preserved.

A block counts as prose only when *every* line is plain prose. A single
indented line, list item, heading, or markdown link definition makes the whole
block verbatim, so its layout is preserved untouched.

For prose:

1. **Full stop** — every prose block ends in `.`/`!`/`?`/`:`/`)`/`]`; a `.` is
   appended if missing. *Exception:* a short single-line comment (e.g.
   `// increment counter`) is left as-is.
2. **One sentence per line** — sentences are segmented and each starts a new line.
   Abbreviations (`e.g.`, `etc.`), decimals (`3.14`), versions (`v1.2.3`), URLs,
   qualified identifiers and `[links]` are not mistaken for sentence ends.
3. **Width reflow** — each sentence is wrapped to the target width at word
   boundaries (no hyphenation).
4. **Synopsis blank line** *(godoc doc comments only)* — the first sentence becomes
   its own paragraph, followed by a blank line.
5. **Dense-prose breaks** — a run of body sentences longer than `-max-lines`
   physical lines is split into paragraphs.

## What it leaves alone

- `/* … */` block comments
- Trailing/inline comments (`x := 1 // sets x`)
- Directive comments (`//go:generate`, `//go:build`, `//nolint…`, `//line …`,
  legacy `// +build`)
- Generated files (`// Code generated … DO NOT EDIT.`)
- Indented blocks (code, list continuations), `- `/`1. ` lists, `# ` headings
- Markdown link definitions (`[ref]: https://…`) — kept on one line, never
  wrapped

## Not (yet) supported

Identifier `[bracket]` auto-linking, `/* */` reflow, glob patterns, spell-checking,
and rune/terminal-width measurement (widths are counted in bytes).

## Development

```sh
go test ./...                       # unit + golden + idempotency tests
go test ./internal/reflow -fuzz=FuzzReflowIdempotent      # fuzz the engine
go test ./internal/formatter -run TestGolden -update      # refresh golden files
```
>>>>>>> 3076c11 (Initial commit: comment-formatter, a gofmt-style Go comment reflow CLI)
