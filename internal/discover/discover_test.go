package discover

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestFiles(t *testing.T) {
	root := t.TempDir()
	mk := func(rel string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("package p\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("a.go")
	mk("sub/b.go")
	mk("sub/notes.txt")
	mk("vendor/c.go")
	mk("testdata/d.go")
	mk(".hidden/e.go")

	got, err := Files([]string{root})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		filepath.Join(root, "a.go"),
		filepath.Join(root, "sub/b.go"),
	}
	if !slices.Equal(got, want) {
		t.Errorf("Files() = %v, want %v", got, want)
	}
}

func TestFilesExplicitFile(t *testing.T) {
	root := t.TempDir()
	// An explicitly named file is always included, even under testdata.
	p := filepath.Join(root, "testdata", "x.go")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("package p\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Files([]string{p})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []string{p}) {
		t.Errorf("Files(%q) = %v", p, got)
	}
}
