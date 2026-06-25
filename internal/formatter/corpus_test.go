package formatter

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// corpusRoots are real-world Go trees used to stress idempotency.
//
// They are skipped when absent so the test stays portable.
var corpusRoots = []string{
	"/home/fred/src/github.com/fredbi/go-fred-mcp",
	runtimeGOROOTsrc(),
}

func runtimeGOROOTsrc() string {
	if r := os.Getenv("GOROOT"); r != "" {
		return filepath.Join(r, "src")
	}
	return ""
}

// TestCorpusIdempotent runs Format twice over every .go file it can find in the
// corpus roots and asserts the second pass equals the first.
//
// It does not assert any particular formatting — only that the transform is a
// fixed point on real code.
func TestCorpusIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping corpus test in -short mode")
	}

	var checked int
	for _, root := range corpusRoots {
		if root == "" {
			continue
		}
		if _, err := os.Stat(root); err != nil {
			t.Logf("skipping absent corpus root %s", root)
			continue
		}

		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil //nolint:nilerr // skip unreadable entries
			}
			if d.IsDir() {
				if name := d.Name(); name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") {
					if path != root {
						return fs.SkipDir
					}
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			src, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			once, err := Format(path, src, DefaultOptions())
			if err != nil {
				return nil // unparseable / build-tagged file: out of scope
			}
			twice, err := Format(path, once, DefaultOptions())
			if err != nil {
				t.Errorf("%s: second Format failed: %v", path, err)
				return nil
			}
			if !bytes.Equal(once, twice) {
				t.Errorf("%s: Format is not idempotent", path)
			}
			checked++
			return nil
		})
	}
	t.Logf("checked %d files for idempotency", checked)
	if checked == 0 {
		t.Skip("no corpus files available")
	}
}
