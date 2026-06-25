package reflow

import "strings"

// Wrap greedily wraps a single sentence into lines no longer than budget bytes,
// breaking only at word boundaries.
//
// A word longer than budget is emitted alone on its own (over-long) line rather
// than broken — no hyphenation.
// Runs of whitespace collapse to a single space.
//
// Wrap is the inverse-stable half of the idempotency contract: callers must
// unwrap text to single spaces before wrapping, so that re-wrapping
// already-wrapped text reproduces the same lines.
func Wrap(sentence string, budget int) []string {
	if budget < 1 {
		budget = 1
	}
	words := strings.Fields(sentence)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	var cur strings.Builder
	for _, w := range words {
		switch {
		case cur.Len() == 0:
			cur.WriteString(w)
		case cur.Len()+1+len(w) <= budget:
			cur.WriteByte(' ')
			cur.WriteString(w)
		default:
			lines = append(lines, cur.String())
			cur.Reset()
			cur.WriteString(w)
		}
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}
