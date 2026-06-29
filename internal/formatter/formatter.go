// Package formatter ties the pipeline together: it extracts comment groups from
// a Go source file, reflows the eligible ones, and splices the results back,
// returning canonical gofmt-valid bytes.
//
// Eligibility: only //-style comments that occupy their own line(s) are
// reformatted.
// Trailing inline comments, /* */ block comments, directive and build comments,
// the "// Output:" block of a testable example, and generated files are passed
// through untouched.
package formatter

import (
	"bytes"
	"strings"

	"github.com/fredbi/comment-formatter/internal/block"
	"github.com/fredbi/comment-formatter/internal/marker"
	"github.com/fredbi/comment-formatter/internal/reflow"
	"github.com/fredbi/comment-formatter/internal/source"
)

// Options configures the formatter.
type Options struct {
	Width    int // target full-line width, default 80
	MaxLines int // dense-prose paragraph threshold, default 4
}

// DefaultOptions returns the default formatting options.
func DefaultOptions() Options {
	return Options{Width: 80, MaxLines: 4}
}

func (o Options) normalized() Options {
	if o.Width <= 0 {
		o.Width = 80
	}
	if o.MaxLines <= 0 {
		o.MaxLines = 4
	}
	return o
}

// Format reflows the comments in src and returns the reformatted file.
//
// It is a pure function (no I/O).
// Generated files and files that need no comment change are returned unchanged;
// files with comment edits are additionally run through go/format.Source.
// A parse or format error leaves src untouched and is returned to the caller.
func Format(filename string, src []byte, opt Options) ([]byte, error) {
	opt = opt.normalized()

	if block.IsGenerated(src) {
		return src, nil
	}

	_, groups, err := source.Extract(filename, src)
	if err != nil {
		return src, err
	}

	isTest := strings.HasSuffix(filename, "_test.go")

	var edits []source.Edit
	for _, g := range groups {
		if !eligible(g, isTest) {
			continue
		}
		body := marker.Strip(g.Text)
		newBody := reflow.Reflow(body, reflow.Config{
			Width:    opt.Width,
			MaxLines: opt.MaxLines,
			IsDoc:    g.IsDoc,
			Indent:   g.Indent,
		})
		newText := marker.Restore(newBody, g.Indent)
		if newText != g.Text {
			edits = append(edits, source.Edit{Start: g.Start, End: g.End, NewText: newText})
		}
	}

	if len(edits) == 0 {
		return src, nil
	}
	return source.Apply(src, edits)
}

// Changed reports whether Format would modify src.
func Changed(filename string, src []byte, opt Options) (bool, error) {
	out, err := Format(filename, src, opt)
	if err != nil {
		return false, err
	}
	return !bytes.Equal(out, src), nil
}

// eligible reports whether a comment group should be reformatted.
//
// isTest indicates the group's file carries the "_test.go" suffix — the same
// signal go/build uses to recognise test sources. It gates the example-output
// exception, which is only meaningful in test files.
func eligible(g source.CommentGroup, isTest bool) bool {
	switch {
	case !g.LineComment: // /* */ block comment
		return false
	case g.Inline: // trailing comment after code
		return false
	case block.IsDirective(g.Text):
		return false
	case isTest && g.InExampleFunc && block.IsExampleOutput(g.Text):
		// The expected-output block of a testable example is compared verbatim;
		// reflowing it would break the example.
		return false
	default:
		return true
	}
}
