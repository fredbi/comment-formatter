package reflow

import "testing"

func cfg(width, maxLines int, isDoc bool) Config {
	return Config{Width: width, MaxLines: maxLines, IsDoc: isDoc}
}

func TestReflow(t *testing.T) {
	tests := []struct {
		name string
		body string
		cfg  Config
		want string
	}{
		{
			name: "single-line single-sentence exempt",
			body: "increment counter",
			cfg:  cfg(80, 4, false),
			want: "increment counter",
		},
		{
			name: "single-line exempt keeps missing period",
			body: "bar is a helper",
			cfg:  cfg(80, 4, true),
			want: "bar is a helper",
		},
		{
			name: "synopsis split for doc",
			body: "Foo does the work. It also does more.",
			cfg:  cfg(80, 4, true),
			want: "Foo does the work.\n\nIt also does more.",
		},
		{
			name: "no synopsis for non-doc",
			body: "Foo does the work. It also does more.",
			cfg:  cfg(80, 4, false),
			want: "Foo does the work.\nIt also does more.",
		},
		{
			name: "multi-line single sentence gets period",
			body: "Baz processes input\nacross two lines without a period",
			cfg:  cfg(80, 4, true),
			want: "Baz processes input across two lines without a period.",
		},
		{
			name: "code block passthrough",
			body: "Example:\n\n\tx := 1\n\ty := 2",
			cfg:  cfg(80, 4, true),
			want: "Example:\n\n\tx := 1\n\ty := 2",
		},
		{
			name: "list passthrough",
			body: "Items:\n\n  - one\n  - two",
			cfg:  cfg(80, 4, true),
			want: "Items:\n\n  - one\n  - two",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Reflow(tc.body, tc.cfg)
			if got != tc.want {
				t.Errorf("Reflow:\n got:\n%s\n want:\n%s", got, tc.want)
			}
		})
	}
}

func TestReflowWidthWrap(t *testing.T) {
	// Width 30 -> budget 27 for text (no indent).
	// Each wrapped line's text must be <= 27 bytes.
	body := "This is one fairly long sentence that must wrap across several short lines."
	got := Reflow(body, cfg(30, 4, false))
	for _, line := range splitLines(got) {
		if line == "" {
			continue
		}
		if len(line) > 27 {
			t.Errorf("line exceeds budget 27: %q (len %d)", line, len(line))
		}
	}
}

func TestReflowDenseProseBreak(t *testing.T) {
	// Five one-line sentences, maxLines 4 -> a blank line before the 5th (counting
	// from the synopsis-excluded body).
	// Non-doc: no synopsis.
	body := "S1 alpha. S2 beta. S3 gamma. S4 delta. S5 epsilon."
	got := Reflow(body, cfg(80, 4, false))
	want := "S1 alpha.\nS2 beta.\nS3 gamma.\nS4 delta.\n\nS5 epsilon."
	if got != want {
		t.Errorf("dense break:\n got:\n%s\n want:\n%s", got, want)
	}
}

func TestReflowIdempotent(t *testing.T) {
	bodies := []string{
		"increment counter",
		"Foo does X. It also does Y. And Z is here too.",
		"S1 alpha. S2 beta. S3 gamma. S4 delta. S5 epsilon. S6 zeta. S7 eta.",
		"Baz processes input\nacross two lines without a period",
		"Example:\n\n\tx := 1\n\ty := 2",
		"A very very very very very very very very very very very very long sentence.",
	}
	for _, isDoc := range []bool{true, false} {
		for _, b := range bodies {
			c := cfg(40, 3, isDoc)
			once := Reflow(b, c)
			twice := Reflow(once, c)
			if once != twice {
				t.Errorf("not idempotent (isDoc=%v) for %q:\n once:\n%s\n twice:\n%s", isDoc, b, once, twice)
			}
		}
	}
}

func FuzzReflowIdempotent(f *testing.F) {
	f.Add("Foo does X. It also does Y.", true, 80, 4)
	f.Add("increment counter", false, 80, 4)
	f.Add("S1. S2. S3. S4. S5. S6.", true, 40, 2)
	f.Fuzz(func(t *testing.T, body string, isDoc bool, width, maxLines int) {
		if width < 4 || width > 200 || maxLines < 1 || maxLines > 20 {
			t.Skip()
		}
		c := cfg(width, maxLines, isDoc)
		once := Reflow(body, c)
		twice := Reflow(once, c)
		if once != twice {
			t.Errorf("not idempotent:\n in:   %q\n once: %q\n twice:%q", body, once, twice)
		}
	})
}

func splitLines(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == '\n' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	out = append(out, cur)
	return out
}
