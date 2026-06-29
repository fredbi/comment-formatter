// Package source parses Go files into comment groups with byte ranges, and
// splices reformatted comments back into the original bytes.
//
// It is deliberately syntax-only: it uses go/parser with parser.ParseComments
// and never loads type information.
//
// Every comment in the file is surfaced — declaration doc comments,
// free-floating comments, and trailing inline comments — each annotated with
// the metadata the formatter needs to decide how (or whether) to reformat it.
package source

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// CommentGroup describes one comment group located in a source file.
type CommentGroup struct {
	// Start and End are byte offsets into the original source.
	//
	// The range [Start, End) covers the comment group exactly, beginning at the
	// first marker and ending after the last comment line (leading indentation
	// before the first marker is not included).
	Start int
	End   int

	// Text is the raw source slice src[Start:End].
	Text string

	// Indent is the whitespace that precedes the first marker on its line, reused
	// to re-indent continuation lines on restore.
	Indent string

	// IsDoc is true when the group is attached to a declaration (or the file) as
	// its .Doc comment — i.e. a godoc doc comment.
	IsDoc bool

	// Inline is true when the group trails other code on the same line (e.g. "x :=
	// 1 // sets x").
	Inline bool

	// LineComment is true for //-style groups, false for /* */ groups.
	LineComment bool

	// InExampleFunc is true when the group sits inside the body of a testable
	// example function (a top-level "func Example...()" with no receiver). It
	// gates the "// Output:" preservation rule, which must not fire for ordinary
	// comments that merely happen to start with "Output:".
	InExampleFunc bool
}

// Extract parses src and returns the file set together with every comment
// group, sorted by ascending Start offset.
func Extract(filename string, src []byte) (*token.FileSet, []CommentGroup, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	docGroups := collectDocGroups(file)
	exampleBodies := collectExampleBodies(fset, file)

	groups := make([]CommentGroup, 0, len(file.Comments))
	for _, cg := range file.Comments {
		if len(cg.List) == 0 {
			continue
		}
		start := fset.Position(cg.Pos()).Offset
		end := fset.Position(cg.End()).Offset

		groups = append(groups, CommentGroup{
			Start:         start,
			End:           end,
			Text:          string(src[start:end]),
			Indent:        lineIndent(src, start),
			IsDoc:         docGroups[cg],
			Inline:        hasCodeBefore(src, start),
			LineComment:   cg.List[0].Text[:2] == "//",
			InExampleFunc: withinAny(exampleBodies, start),
		})
	}
	return fset, groups, nil
}

// span is a half-open byte range [start, end) in the source.
type span struct{ start, end int }

// collectExampleBodies returns the byte ranges of the bodies of top-level
// testable example functions — "func Example...()" declarations with no
// receiver. Comment groups falling inside one of these ranges carry
// InExampleFunc=true.
func collectExampleBodies(fset *token.FileSet, file *ast.File) []span {
	var spans []span
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Body == nil {
			continue
		}
		if !strings.HasPrefix(fn.Name.Name, "Example") {
			continue
		}
		spans = append(spans, span{
			start: fset.Position(fn.Body.Pos()).Offset,
			end:   fset.Position(fn.Body.End()).Offset,
		})
	}
	return spans
}

// withinAny reports whether offset lies inside any of the given spans.
func withinAny(spans []span, offset int) bool {
	for _, s := range spans {
		if offset >= s.start && offset < s.end {
			return true
		}
	}
	return false
}

// collectDocGroups returns the set of comment groups referenced as the .Doc
// comment of the file or any declaration.
//
// Trailing .Comment groups are intentionally excluded — they are not godoc
// doc comments.
func collectDocGroups(file *ast.File) map[*ast.CommentGroup]bool {
	docs := make(map[*ast.CommentGroup]bool)
	add := func(cg *ast.CommentGroup) {
		if cg != nil {
			docs[cg] = true
		}
	}

	add(file.Doc)
	ast.Inspect(file, func(n ast.Node) bool {
		switch d := n.(type) {
		case *ast.GenDecl:
			add(d.Doc)
		case *ast.FuncDecl:
			add(d.Doc)
		case *ast.TypeSpec:
			add(d.Doc)
		case *ast.ValueSpec:
			add(d.Doc)
		case *ast.Field:
			add(d.Doc)
		}
		return true
	})
	return docs
}

// lineIndent returns the run of spaces/tabs immediately preceding offset on its
// line.
//
// If any non-whitespace precedes offset on the line, the comment is inline and
// the indent is irrelevant (the empty string is returned).
func lineIndent(src []byte, offset int) string {
	start := lineStart(src, offset)
	for i := start; i < offset; i++ {
		if src[i] != ' ' && src[i] != '\t' {
			return ""
		}
	}
	return string(src[start:offset])
}

// hasCodeBefore reports whether non-whitespace appears before offset on its
// line, i.e. the comment trails code.
func hasCodeBefore(src []byte, offset int) bool {
	start := lineStart(src, offset)
	for i := start; i < offset; i++ {
		if src[i] != ' ' && src[i] != '\t' {
			return true
		}
	}
	return false
}

// lineStart returns the byte offset of the first character on the line
// containing offset.
func lineStart(src []byte, offset int) int {
	i := offset
	for i > 0 && src[i-1] != '\n' {
		i--
	}
	return i
}
