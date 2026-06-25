// Package marker strips and restores the "//" line-comment markers that wrap a
// comment group's text.
//
// A comment group extracted from source looks like a run of "//"-prefixed
// lines, where continuation lines also carry the surrounding code indentation.
// [Strip] turns such a group into its plain body text (one logical line per
// source line, markers and one optional leading space removed).
//
// [Restore] does the inverse, re-emitting gofmt-stable "//" lines.
package marker

import "strings"

// Strip removes the "//" marker and one optional following space from every
// line of a //-comment group, returning the joined body text.
//
// Leading indentation (tabs or spaces preceding the marker on continuation
// lines) is discarded.
// A bare "//" line becomes an empty body line.
func Strip(group string) string {
	lines := strings.Split(group, "\n")
	out := make([]string, len(lines))
	for i, line := range lines {
		// Drop any indentation that precedes the marker on continuation lines.
		trimmed := strings.TrimLeft(line, " \t")
		rest := strings.TrimPrefix(trimmed, "//")
		// Drop a single conventional space after the marker; keep the rest verbatim
		// so code-block and list indentation survive.
		rest = strings.TrimPrefix(rest, " ")
		out[i] = rest
	}
	return strings.Join(out, "\n")
}

// Restore renders body back into //-comment source lines.
//
// The first line is emitted without indentation (the caller splices it in place
// of the original marker, after the existing indent); every continuation line
// is prefixed with indent.
//
// Non-empty lines become "// text"; empty lines become a bare "//" with no
// trailing space.
func Restore(body, indent string) string {
	lines := strings.Split(body, "\n")
	var b strings.Builder
	for i, line := range lines {
		if i > 0 {
			b.WriteByte('\n')
			b.WriteString(indent)
		}
		if line == "" {
			b.WriteString("//")
			continue
		}
		b.WriteString("// ")
		b.WriteString(line)
	}
	return b.String()
}
