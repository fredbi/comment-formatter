// Package block splits a marker-stripped comment body into blank-line-delimited
// blocks and classifies each as prose, code, list, or heading.
//
// It also recognizes directive comments that must be passed through verbatim.
//
// Classification mirrors how go/doc/comment (and therefore gofmt) interprets
// doc-comment structure: indented lines are code, "- "/"1. " lines are lists,
// "# " lines are headings, everything else is prose.
//
// Only prose blocks are reflowed by the formatter; the rest are preserved
// (gofmt canonicalizes doc-comment structure on the final pass).
package block

import (
	"regexp"
	"strings"
)

// Kind classifies a block's content.
type Kind int

const (
	// Prose is free-flowing sentences, the only kind that gets reflowed.
	Prose Kind = iota
	// Code is an indented (verbatim) block.
	Code
	// List is a bullet or numbered list.
	List
	// Heading is a "# " heading line.
	Heading
)

// Block is a run of consecutive non-blank body lines plus its classification.
type Block struct {
	Kind  Kind
	Lines []string
}

var (
	reBullet   = regexp.MustCompile(`^\s*[-*+]\s`)
	reNumbered = regexp.MustCompile(`^\s*\d+[.)]\s`)
	reHeading  = regexp.MustCompile(`^#\s`)

	// reDirective matches tool directive comments such as //go:generate,
	// //go:build, //nolint:all, //revive:disable — a marker with no space
	// followed by word characters and a colon.
	reDirective = regexp.MustCompile(`^//[a-z_][a-z0-9_]*:`)
	// reLegacyBuild matches the legacy "// +build" constraint line.
	reLegacyBuild = regexp.MustCompile(`^//\s*\+build\b`)
	// reGenerated matches the "Code generated ... DO NOT EDIT." marker line.
	reGenerated = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)
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

// classify determines the kind of a non-empty block from its first line.
func classify(lines []string) Kind {
	first := lines[0]
	switch {
	case reHeading.MatchString(first):
		return Heading
	case reBullet.MatchString(first), reNumbered.MatchString(first):
		return List
	case len(first) > 0 && (first[0] == ' ' || first[0] == '\t'):
		return Code
	default:
		return Prose
	}
}

// IsDirective reports whether the raw comment-group text is a directive,
// build-constraint, or generated-code marker that must not be reformatted.
func IsDirective(raw string) bool {
	for line := range strings.SplitSeq(raw, "\n") {
		line = strings.TrimLeft(line, " \t")
		switch {
		case strings.HasPrefix(line, "//line "), line == "//line":
			return true
		case reDirective.MatchString(line):
			return true
		case reLegacyBuild.MatchString(line):
			return true
		}
	}
	return false
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
