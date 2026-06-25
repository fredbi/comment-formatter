package source

import (
	"fmt"
	"go/format"
	"sort"
)

// Edit replaces the byte range [Start, End) of the original source with
// NewText.
type Edit struct {
	Start   int
	End     int
	NewText string
}

// Apply splices edits into src and returns the result run through
// go/format.Source.
//
// Edits are applied bottom-to-top (descending Start) so that earlier
// replacements never invalidate the byte offsets of later ones.
//
// The final gofmt pass guarantees valid, canonical Go: in particular it
// normalizes doc-comment structure (lists, headings, code-block indentation)
// without rewrapping paragraph text, so reflowed line breaks are preserved.
func Apply(src []byte, edits []Edit) ([]byte, error) {
	if len(edits) == 0 {
		return format.Source(src)
	}

	sorted := make([]Edit, len(edits))
	copy(sorted, edits)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Start > sorted[j].Start })

	out := make([]byte, len(src))
	copy(out, src)
	for _, e := range sorted {
		if e.Start < 0 || e.End > len(out) || e.Start > e.End {
			return nil, fmt.Errorf("invalid edit range [%d:%d] for source of length %d", e.Start, e.End, len(out))
		}
		out = append(out[:e.Start], append([]byte(e.NewText), out[e.End:]...)...)
	}

	return format.Source(out)
}
