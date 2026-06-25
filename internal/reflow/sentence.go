package reflow

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// DefaultAbbrev is the set of lower-cased abbreviations (without their trailing
// dot) that must not be treated as sentence ends.
//
// Lone single-letter initials (a., U.S., e.g.) are handled separately by a
// length check, so they need not be listed here.
// The map is exported so callers and tests can extend it.
var DefaultAbbrev = map[string]struct{}{
	"etc": {}, "vs": {}, "cf": {}, "al": {}, "viz": {}, "resp": {},
	"approx": {}, "fig": {}, "no": {}, "vol": {}, "ch": {}, "pp": {},
	"eq": {}, "ca": {}, "est": {}, "mr": {}, "mrs": {}, "ms": {}, "dr": {},
}

// SplitSentences splits one unwrapped prose block into sentences.
//
// A boundary is a '.', '!' or '?' (plus any trailing closing
// punctuation/quotes) that is followed by whitespace and a sentence-start rune
// (upper-case letter, digit, '[' or a quote), or by end of input.
//
// It deliberately under-segments rather than over-segments: dots inside
// versions, decimals, URLs and qualified identifiers are safe because they are
// not followed by whitespace; abbreviations and digits before a dot suppress a
// boundary.
//
// These accepted failure modes keep the splitter pure and idempotent without
// NLP.
// Each returned sentence is trimmed of surrounding whitespace and retains its
// terminal punctuation.
func SplitSentences(block string, abbrev map[string]struct{}) []string {
	if abbrev == nil {
		abbrev = DefaultAbbrev
	}
	block = strings.TrimSpace(block)
	if block == "" {
		return nil
	}

	var sentences []string
	start := 0
	for i := 0; i < len(block); i++ {
		c := block[i]
		if c != '.' && c != '!' && c != '?' {
			continue
		}
		if !isBoundary(block, i, abbrev) {
			continue
		}

		// Consume any trailing closing punctuation that belongs to this sentence.
		end := i + 1
		for end < len(block) && isCloser(block[end]) {
			end++
		}

		// Require whitespace + sentence-start (or end of input) to confirm.
		rest := block[end:]
		trimmed := strings.TrimLeft(rest, " \t")
		if len(rest) == len(trimmed) && trimmed != "" {
			// No whitespace separates this from the next token: not a boundary.
			continue
		}
		if trimmed != "" && !startsSentence(trimmed) {
			continue
		}

		sentences = append(sentences, strings.TrimSpace(block[start:end]))
		// Advance past the whitespace to the next sentence start.
		start = end + (len(rest) - len(trimmed))
		i = start - 1
	}

	if start < len(block) {
		if s := strings.TrimSpace(block[start:]); s != "" {
			sentences = append(sentences, s)
		}
	}
	return sentences
}

// isBoundary applies the abbreviation, lone-initial, digit and ellipsis guards
// to the punctuation char at index i.
func isBoundary(s string, i int, abbrev map[string]struct{}) bool {
	if s[i] == '.' {
		// Ellipsis: part of a "..." run is never a boundary.
		if (i > 0 && s[i-1] == '.') || (i+1 < len(s) && s[i+1] == '.') {
			return false
		}
		// Digit before the dot: decimals, versions, "rule 3." — suppress.
		if i > 0 && s[i-1] >= '0' && s[i-1] <= '9' {
			return false
		}
		// Abbreviations and lone initials.
		token := s[wordStart(s, i):i] // letters immediately before the dot
		if len(token) == 1 && isLetter(token[0]) {
			return false // lone initial: "U.S.", "e.g.", "a."
		}
		if _, ok := abbrev[strings.ToLower(token)]; ok {
			return false
		}
	}
	return true
}

// wordStart returns the index of the first letter of the contiguous letter run
// ending just before index i.
func wordStart(s string, i int) int {
	j := i
	for j > 0 && isLetter(s[j-1]) {
		j--
	}
	return j
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isCloser(b byte) bool {
	switch b {
	case ')', ']', '"', '\'', '`':
		return true
	default:
		return false
	}
}

// startsSentence reports whether the next token begins a new sentence.
func startsSentence(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	switch {
	case unicode.IsUpper(r), unicode.IsDigit(r):
		return true
	case r == '[' || r == '`' || r == '"' || r == '\'':
		return true
	default:
		return false
	}
}
