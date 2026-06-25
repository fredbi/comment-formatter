package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunStdin(t *testing.T) {
	in := strings.NewReader("package p\n\n// Foo does work. It also does more.\nfunc Foo() {}\n")
	var out, errBuf bytes.Buffer
	if code := run(nil, in, &out, &errBuf); code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errBuf.String())
	}
	got := out.String()
	if !strings.Contains(got, "// Foo does work.\n//\n// It also does more.") {
		t.Errorf("unexpected stdout:\n%s", got)
	}
}

func TestRunListAndWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.go")
	orig := "package p\n\n// Foo does work. It also does more.\nfunc Foo() {}\n"
	if err := os.WriteFile(path, []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}

	// -l lists the file as changed without modifying it.
	var out, errBuf bytes.Buffer
	if code := run([]string{"-l", path}, nil, &out, &errBuf); code != 0 {
		t.Fatalf("-l exit %d: %s", code, errBuf.String())
	}
	if strings.TrimSpace(out.String()) != path {
		t.Errorf("-l listed %q, want %q", out.String(), path)
	}
	if cur, _ := os.ReadFile(path); string(cur) != orig {
		t.Error("-l must not modify the file")
	}

	// -w rewrites the file in place.
	out.Reset()
	errBuf.Reset()
	if code := run([]string{"-w", path}, nil, &out, &errBuf); code != 0 {
		t.Fatalf("-w exit %d: %s", code, errBuf.String())
	}
	cur, _ := os.ReadFile(path)
	if !strings.Contains(string(cur), "// It also does more.") {
		t.Errorf("-w did not rewrite file:\n%s", cur)
	}

	// Running -l again now reports no changes (idempotent).
	out.Reset()
	if code := run([]string{"-l", path}, nil, &out, &errBuf); code != 0 {
		t.Fatalf("second -l exit %d", code)
	}
	if out.String() != "" {
		t.Errorf("expected no changes after -w, got %q", out.String())
	}
}

func TestRunStdinRejectsWrite(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := run([]string{"-w"}, strings.NewReader("package p\n"), &out, &errBuf); code != 2 {
		t.Errorf("expected exit 2 for -w with stdin, got %d", code)
	}
}
