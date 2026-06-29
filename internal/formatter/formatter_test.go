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

// exampleSrc holds a testable example whose Output block must survive untouched,
// alongside a same-named comment shape that must NOT be preserved.
const exampleSrc = `package golden

import "fmt"

func ExampleFoo() {
	fmt.Println("hello")
	// Output:
	// hello this expected line is intentionally far longer than eighty columns so that a reflow would corrupt the example were it ever applied
}

func notAnExample() {
	// Output: this is not inside an example so this long line should be reflowed by the formatter as ordinary prose without any preservation here today
	fmt.Println("x")
}
`

// TestExampleOutputPreservation checks that the "// Output:" block of a testable
// example is preserved only in *_test.go files, and only inside Example funcs.
func TestExampleOutputPreservation(t *testing.T) {
	const outputLine = "// hello this expected line is intentionally far longer than eighty columns so that a reflow would corrupt the example were it ever applied"
	// The look-alike comment, as a single unbroken source line. Reflow wraps it
	// (and adds a trailing period), so its survival verbatim means preservation.
	const notExampleLine = "// Output: this is not inside an example so this long line should be reflowed by the formatter as ordinary prose without any preservation here today"

	// In a test file the example Output block is left verbatim...
	out, err := Format("example_test.go", []byte(exampleSrc), DefaultOptions())
	if err != nil {
		t.Fatalf("Format(test): %v", err)
	}
	if !strings.Contains(string(out), outputLine) {
		t.Errorf("example Output block was modified in a _test.go file:\n%s", out)
	}
	// ...while the look-alike comment in a non-example function is still reflowed.
	if strings.Contains(string(out), notExampleLine) {
		t.Errorf("non-example comment was not reflowed:\n%s", out)
	}

	// The same source in a non-test file gets no special treatment: the long
	// Output line is reflowed like any other prose.
	out, err = Format("example.go", []byte(exampleSrc), DefaultOptions())
	if err != nil {
		t.Fatalf("Format(non-test): %v", err)
	}
	if strings.Contains(string(out), outputLine) {
		t.Errorf("example Output block should be reflowed in a non-test file:\n%s", out)
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
