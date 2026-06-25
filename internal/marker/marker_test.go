package marker

import "testing"

func TestStrip(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"// hello", "hello"},
		{"//hello", "hello"},
		{"//  hello", " hello"},
		{"// a\n// b", "a\nb"},
		{"// a\n\t// b", "a\nb"},
		{"//", ""},
		{"// a\n//\n// b", "a\n\nb"},
		{"//\tcode", "\tcode"},
	}
	for _, tc := range tests {
		if got := Strip(tc.in); got != tc.want {
			t.Errorf("Strip(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestRestore(t *testing.T) {
	tests := []struct {
		body, indent, want string
	}{
		{"hello", "", "// hello"},
		{"a\nb", "", "// a\n// b"},
		{"a\nb", "\t", "// a\n\t// b"},
		{"a\n\nb", "", "// a\n//\n// b"},
		{"", "", "//"},
	}
	for _, tc := range tests {
		if got := Restore(tc.body, tc.indent); got != tc.want {
			t.Errorf("Restore(%q, %q) = %q, want %q", tc.body, tc.indent, got, tc.want)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	// Restore(Strip(g)) must be stable under repetition for canonical input.
	groups := []struct{ group, indent string }{
		{"// a\n// b", ""},
		{"// a\n\t// b", "\t"},
		{"// a\n//\n// b", ""},
	}
	for _, g := range groups {
		once := Restore(Strip(g.group), g.indent)
		twice := Restore(Strip(once), g.indent)
		if once != twice {
			t.Errorf("round-trip not stable: %q -> %q -> %q", g.group, once, twice)
		}
	}
}
