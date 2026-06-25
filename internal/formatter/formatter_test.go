package formatter

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

// TestGolden runs Format over every *.input.go fixture and compares with the
// matching *.golden.go file.
//
// Run with -update to regenerate the goldens.
func TestGolden(t *testing.T) {
	inputs, err := filepath.Glob("testdata/golden/*.input.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) == 0 {
		t.Fatal("no golden inputs found")
	}

	for _, in := range inputs {
		name := strings.TrimSuffix(filepath.Base(in), ".input.go")
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(in)
			if err != nil {
				t.Fatal(err)
			}
			got, err := Format(in, src, DefaultOptions())
			if err != nil {
				t.Fatalf("Format: %v", err)
			}

			goldenPath := strings.TrimSuffix(in, ".input.go") + ".golden.go"
			if *update {
				if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden (run -update?): %v", err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("Format(%s) mismatch:\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
			}

			// The golden output must be a fixed point.
			again, err := Format(goldenPath, want, DefaultOptions())
			if err != nil {
				t.Fatalf("Format(golden): %v", err)
			}
			if !bytes.Equal(again, want) {
				t.Errorf("golden %s is not a fixed point:\n--- got ---\n%s", name, again)
			}
		})
	}
}

func FuzzFormatIdempotent(f *testing.F) {
	seeds := []string{
		"package p\n\n// Foo does X. It also does Y.\nfunc Foo() {}\n",
		"package p\n\n// short\nvar x int\n",
		"package p\n\n//go:generate x\ntype T int\n",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, body string) {
		src := []byte(body)
		once, err := Format("fuzz.go", src, DefaultOptions())
		if err != nil {
			t.Skip() // unparseable input is out of scope
		}
		twice, err := Format("fuzz.go", once, DefaultOptions())
		if err != nil {
			t.Fatalf("second Format failed on valid output: %v", err)
		}
		if !bytes.Equal(once, twice) {
			t.Errorf("Format not idempotent:\n--- once ---\n%s\n--- twice ---\n%s", once, twice)
		}
	})
}
