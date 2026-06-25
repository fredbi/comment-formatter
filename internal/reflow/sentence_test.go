package reflow

import (
	"reflect"
	"testing"
)

func TestSplitSentences(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"single", "Just one sentence.", []string{"Just one sentence."}},
		{"two", "First one. Second one.", []string{"First one.", "Second one."}},
		{"bang and question", "Really? Yes! Done.", []string{"Really?", "Yes!", "Done."}},
		{"no trailing punct", "Trailing free", []string{"Trailing free"}},

		// Guards: must NOT split.
		{"abbrev etc", "Supports Linux, macOS, etc. and others here.", []string{"Supports Linux, macOS, etc. and others here."}},
		{"abbrev eg lone-letter", "Use a flag, e.g. the verbose one.", []string{"Use a flag, e.g. the verbose one."}},
		{"abbrev ie", "That is, i.e. exactly this.", []string{"That is, i.e. exactly this."}},
		{"decimal", "Pi is 3.14 today.", []string{"Pi is 3.14 today."}},
		{"version", "Needs v1.2.3 or later.", []string{"Needs v1.2.3 or later."}},
		{"qualified ident", "Call fmt.Sprintf here.", []string{"Call fmt.Sprintf here."}},
		{"url", "See http://example.com/a.b for details.", []string{"See http://example.com/a.b for details."}},
		{"ellipsis", "Wait... and then go on.", []string{"Wait... and then go on."}},
		{"digit before dot", "See rule 3. More text follows.", []string{"See rule 3. More text follows."}},

		// Real boundaries even near tricky tokens.
		{"period after url then capital", "See http://x.io. Then continue.", []string{"See http://x.io.", "Then continue."}},
		{"bracket start", "Done. [Reader] is next.", []string{"Done.", "[Reader] is next."}},
		{"digit start", "Done. 42 is the answer.", []string{"Done.", "42 is the answer."}},
		{"paren closer", "Done (really). Next now.", []string{"Done (really).", "Next now."}},

		// Lowercase next word: do not split (accepted under-segmentation).
		{"lowercase next", "Done. see also below.", []string{"Done. see also below."}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SplitSentences(tc.in, nil)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("SplitSentences(%q):\n got  %#v\n want %#v", tc.in, got, tc.want)
			}
		})
	}
}
