package source

import (
	"strings"
	"testing"
)

const sampleSrc = `package p

// Doc on Foo.
func Foo() {}

func bar() {
	x := 1 // trailing inline
	_ = x
	// free floating
	_ = 0
}

/* block comment */
var z int
`

func TestExtract(t *testing.T) {
	_, groups, err := Extract("sample.go", []byte(sampleSrc))
	if err != nil {
		t.Fatal(err)
	}

	byText := map[string]CommentGroup{}
	for _, g := range groups {
		byText[strings.TrimSpace(g.Text)] = g
	}

	if g, ok := byText["// Doc on Foo."]; !ok || !g.IsDoc || g.Inline {
		t.Errorf("doc comment: got %+v", g)
	}
	if g, ok := byText["// trailing inline"]; !ok || g.IsDoc || !g.Inline {
		t.Errorf("inline comment: got %+v", g)
	}
	if g, ok := byText["// free floating"]; !ok || g.IsDoc || g.Inline {
		t.Errorf("free-floating comment: got %+v", g)
	}
	if g, ok := byText["/* block comment */"]; !ok || g.LineComment {
		t.Errorf("block comment LineComment flag wrong: got %+v", g)
	}
}

func TestApplyBottomToTop(t *testing.T) {
	src := []byte("package p\n\n// one\nvar a int\n\n// two\nvar b int\n")
	_, groups, err := Extract("x.go", src)
	if err != nil {
		t.Fatal(err)
	}
	var edits []Edit
	for _, g := range groups {
		// Replace each comment with an uppercased version.
		edits = append(edits, Edit{Start: g.Start, End: g.End, NewText: strings.ToUpper(g.Text)})
	}
	out, err := Apply(src, edits)
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.Contains(got, "// ONE") || !strings.Contains(got, "// TWO") {
		t.Errorf("edits not applied correctly:\n%s", got)
	}
}
