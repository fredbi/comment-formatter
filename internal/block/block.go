// Package block splits a marker-stripped comment body into blank-line-delimited
// blocks and classifies each as foldable prose or verbatim.
//
// A block is foldable only when every one of its lines is plain prose: not
// indented, and not a list item, heading, or markdown link definition.
//
// Any other shape (an indented line, a "- " or "1. " list, a "# " heading, a
// "[ref]: url" link definition) makes the whole block verbatim, so its line
// structure is preserved untouched.
// Only foldable blocks are reflowed by the formatter.
package block

import (
	"regexp"
	"strings"
)

// Kind classifies a block.
type Kind int

const (
	// Prose is foldable free-flowing text.
	Prose Kind = iota
	// Verbatim is preserved exactly: indented blocks, lists, headings and link
	// definitions.
	Verbatim
)

// Block is a run of consecutive non-blank body lines plus its classification.
type Block struct {
	Kind  Kind
	Lines []string
}

var (
	reBullet   = regexp.MustCompile(`^[-*+]\s`)
	reNumbered = regexp.MustCompile(`^\d+[.)]\s`)
	// reHeading matches markdown-style headings, single or multi-hash ("#
	// Overview", "## Details"). godoc only renders single-hash headings, but all
	// of them are preserved verbatim — never terminated or wrapped.
	reHeading = regexp.MustCompile(`^#+\s`)
	// reLinkDef matches a markdown reference-link definition, e.g. "[spec]:
	// https://example.com".
	//
	// Such lines must never be wrapped.
	reLinkDef = regexp.MustCompile(`^\[[^\]]+\]:(\s|$)`)

	// reDirective matches tool directive comments such as //go:generate,
	// //go:build, //nolint:all, //revive:disable — a marker with no space
	// followed by word characters and a colon.
	reDirective = regexp.MustCompile(`^//[a-z_][a-z0-9_]*:`)
	// reLegacyBuild matches the legacy "// +build" constraint line.
	reLegacyBuild = regexp.MustCompile(`^//\s*\+build\b`)
	// reGenerated matches the "Code generated ... DO NOT EDIT." marker line.
	reGenerated = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)
	// reExampleOutput matches the "// Output:" (or "// Unordered output:") line
	// that opens the expected-output block of a testable example. It mirrors the
	// pattern go/doc uses to recognise example output, applied here to the
	// marker-stripped first line of a comment group.
	reExampleOutput = regexp.MustCompile(`(?i)^[[:space:]]*(unordered )?output:`)
)

// Split breaks a stripped comment body into classified blocks.
//
// Consecutive blank body lines collapse into a single separator; leading and
// trailing blank lines are dropped.
func Split(body string) []Block {
	lines := strings.Split(body, "\n")

	var blocks []Block
	var cur []string
	flush := func() {
		if len(cur) > 0 {
			blocks = append(blocks, Block{Kind: classify(cur), Lines: cur})
			cur = nil
		}
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		cur = append(cur, line)
	}
	flush()
	return blocks
}

// classify returns Prose only when every line is foldable plain prose.
func classify(lines []string) Kind {
	for _, line := range lines {
		if !foldable(line) {
			return Verbatim
		}
	}
	return Prose
}

// foldable reports whether a stripped body line is plain prose that may be
// rewrapped.
//
// Indented lines (the marker for code, list continuations and other structure)
// and list/heading/link-definition lines are not foldable.
func foldable(line string) bool {
	if line == "" {
		return true // blank lines never occur inside a block, but be safe
	}
	if line[0] == ' ' || line[0] == '\t' {
		return false
	}
	switch {
	case reBullet.MatchString(line),
		reNumbered.MatchString(line),
		reHeading.MatchString(line),
		reLinkDef.MatchString(line):
		return false
	default:
		return true
	}
}

// IsDirective reports whether the raw comment-group text is a directive,
// build-constraint, SPDX license header, or generated-code marker that must not
// be reformatted.
func IsDirective(raw string) bool {
	for line := range strings.SplitSeq(raw, "\n") {
		line = strings.TrimLeft(line, " \t")
		switch {
		case strings.HasPrefix(line, "//line "), line == "//line":
			return true
		case strings.HasPrefix(line, "//nolint"):
			return true
		case strings.HasPrefix(line, "// SPDX"), strings.HasPrefix(line, "//SPDX"):
			return true
		case reDirective.MatchString(line):
			return true
		case reLegacyBuild.MatchString(line):
			return true
		}
	}
	return false
}

// IsExampleOutput reports whether the raw comment-group text opens with an
// example "// Output:" (or "// Unordered output:") marker.
//
// Such a group holds the expected output of a testable example: its lines are
// compared verbatim by the test runner, so reflowing them would break the
// example. Only the first line is inspected — the caller is expected to gate
// this on the group living inside an Example function (see
// [github.com/fredbi/comment-formatter/internal/source.CommentGroup.InExampleFunc]).
func IsExampleOutput(raw string) bool {
	first, _, _ := strings.Cut(raw, "\n")
	first = strings.TrimLeft(first, " \t")
	first = strings.TrimPrefix(first, "//")
	return reExampleOutput.MatchString(first)
}

// IsGenerated reports whether src is a generated file (carries a "Code
// generated ... DO NOT EDIT." line before its package clause).
func IsGenerated(src []byte) bool {
	for line := range strings.SplitSeq(string(src), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "package ") {
			return false
		}
		if reGenerated.MatchString(strings.TrimSpace(line)) {
			return true
		}
	}
	return false
}
