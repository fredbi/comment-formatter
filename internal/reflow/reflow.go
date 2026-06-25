// Package reflow reformats the body text of a single comment group.
//
// It applies, in order: full-stop completion, sentence segmentation,
// one-sentence-per-line, width wrapping, a forced synopsis blank line (godoc
// doc comments only), and dense-prose paragraph breaks.
// Non-prose blocks (code, lists, headings) pass through untouched.
//
// The transform is a pure, idempotent function: Reflow(Reflow(x)) == Reflow(x).
package reflow

import (
	"strings"

	"github.com/fredbi/comment-formatter/internal/block"
)

const markerWidth = len("// ")

// Config controls how a comment body is reflowed.
type Config struct {
	// Width is the target maximum width of the full source line, including the
	// leading indentation and "// " marker.
	Width int
	// MaxLines is the dense-prose threshold: a run of body sentences longer than
	// this many physical lines is broken into paragraphs.
	MaxLines int
	// IsDoc is true for godoc doc comments, enabling the synopsis blank line.
	IsDoc bool
	// Indent is the whitespace preceding the comment's "//" marker; it counts
	// against the wrap budget.
	Indent string
	// Abbrev overrides the abbreviation set used during sentence segmentation.
	//
	// A nil value uses DefaultAbbrev.
	Abbrev map[string]struct{}
}

// budget returns the number of bytes available for comment text on one line.
func (c Config) budget() int {
	return max(c.Width-len(c.Indent)-markerWidth, 1)
}

// Reflow reformats a marker-stripped comment body and returns the new body
// (still marker-stripped).
//
// A comment whose entire content is a single sentence on a single line is
// returned unchanged (the single-line exemption from the full-stop rule).
func Reflow(body string, cfg Config) string {
	blocks := block.Split(body)
	if len(blocks) == 0 {
		return body
	}
	if exempt(blocks, cfg) {
		return body
	}

	maxLines := max(cfg.MaxLines, 1)
	budget := cfg.budget()

	var paras [][]string
	for bi, blk := range blocks {
		if blk.Kind != block.Prose {
			paras = append(paras, blk.Lines)
			continue
		}

		sentences := proseSentences(blk.Lines, cfg.Abbrev)
		wrapped := make([][]string, len(sentences))
		for i, s := range sentences {
			wrapped[i] = Wrap(s, budget)
		}

		start := 0
		// Synopsis rule: only the first sentence of the first (prose) block.
		if cfg.IsDoc && bi == 0 && len(wrapped) > 0 {
			paras = append(paras, wrapped[0])
			start = 1
		}
		paras = append(paras, groupByMaxLines(wrapped[start:], maxLines)...)
	}

	return joinParagraphs(paras)
}

// exempt reports whether the comment must be left verbatim: a single prose
// block that is a single sentence on a single line which already fits the width
// budget (e.g. "// increment counter").
//
// A single sentence that overflows the budget is not exempt — it is still
// wrapped (the exemption covers the full-stop rule, not width reflow).
func exempt(blocks []block.Block, cfg Config) bool {
	if len(blocks) != 1 {
		return false
	}
	b := blocks[0]
	if b.Kind != block.Prose || len(b.Lines) != 1 {
		return false
	}
	line := strings.TrimSpace(b.Lines[0])
	if len(line) > cfg.budget() {
		return false
	}
	return len(SplitSentences(line, cfg.Abbrev)) <= 1
}

// proseSentences unwraps a prose block to single spaces, ensures it ends in
// terminal punctuation, and segments it into sentences.
func proseSentences(lines []string, abbrev map[string]struct{}) []string {
	text := strings.Join(lines, " ")
	text = strings.Join(strings.Fields(text), " ")
	text = ensureTerminal(text)
	return SplitSentences(text, abbrev)
}

// ensureTerminal appends a full stop unless the text already ends in terminal
// punctuation, looking past any trailing closing quotes (so a sentence ending
// in `!"` or `."` counts as terminated).
//
// Idempotent: a second call is a no-op.
func ensureTerminal(text string) string {
	end := len(text)
	for end > 0 && (text[end-1] == '"' || text[end-1] == '\'' || text[end-1] == '`') {
		end--
	}
	if end == 0 {
		if text == "" {
			return text
		}
		return text + "."
	}
	switch text[end-1] {
	case '.', '!', '?', ':', ')', ']':
		return text
	default:
		return text + "."
	}
}

// groupByMaxLines packs sentences (each already wrapped to physical lines) into
// paragraphs of at most maxLines lines, inserting a break before a sentence
// that would overflow the current paragraph.
//
// A break is never inserted before the first sentence of a paragraph, so a
// single over-long sentence forms its own paragraph and the result is
// idempotent.
func groupByMaxLines(sentences [][]string, maxLines int) [][]string {
	var paras [][]string
	var cur []string
	for _, s := range sentences {
		if len(cur) > 0 && len(cur)+len(s) > maxLines {
			paras = append(paras, cur)
			cur = nil
		}
		cur = append(cur, s...)
	}
	if len(cur) > 0 {
		paras = append(paras, cur)
	}
	return paras
}

// joinParagraphs joins paragraphs (each a slice of physical lines) with a
// single blank body line between them.
func joinParagraphs(paras [][]string) string {
	var out []string
	for i, p := range paras {
		if i > 0 {
			out = append(out, "")
		}
		out = append(out, p...)
	}
	return strings.Join(out, "\n")
}
