// Command comment-formatter reformats the comments in Go source files.
//
// Like gofmt, with no path arguments it reads from standard input and writes
// the result to standard output; with path arguments it processes each .go file
// (directories are walked recursively) and prints the reformatted source unless
// -w (write in place) or -l (list changed files) is given.
//
// Usage:
//
//	comment-formatter [flags] [path ...]
//
// Flags:
//
//	-w           write result to (source) file instead of stdout
//	-l           list files whose formatting differs
//	-width int   target line width including the "// " marker (default 80)
//	-max-lines   dense-prose paragraph break threshold (default 4)
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/fredbi/comment-formatter/internal/discover"
	"github.com/fredbi/comment-formatter/internal/formatter"
)

func newFlagSet(stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet("comment-formatter", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: comment-formatter [flags] [path ...]")
		fs.PrintDefaults()
	}
	return fs
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

type config struct {
	write    bool
	list     bool
	width    int
	maxLines int
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet(stderr)
	var cfg config
	fs.BoolVar(&cfg.write, "w", false, "write result to (source) file instead of stdout")
	fs.BoolVar(&cfg.list, "l", false, "list files whose formatting differs")
	fs.IntVar(&cfg.width, "width", 80, "target line width including the \"// \" marker")
	fs.IntVar(&cfg.maxLines, "max-lines", 4, "dense-prose paragraph break threshold")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	opt := formatter.Options{Width: cfg.width, MaxLines: cfg.maxLines}
	paths := fs.Args()

	if len(paths) == 0 {
		if cfg.write || cfg.list {
			fmt.Fprintln(stderr, "comment-formatter: cannot use -w or -l with standard input")
			return 2
		}
		return processStdin(stdin, stdout, stderr, opt)
	}

	files, err := discover.Files(paths)
	if err != nil {
		fmt.Fprintf(stderr, "comment-formatter: %v\n", err)
		return 2
	}

	exit := 0
	for _, path := range files {
		if err := processFile(path, stdout, cfg, opt); err != nil {
			fmt.Fprintf(stderr, "comment-formatter: %s: %v\n", path, err)
			exit = 2
		}
	}
	return exit
}

func processStdin(stdin io.Reader, stdout, stderr io.Writer, opt formatter.Options) int {
	src, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(stderr, "comment-formatter: %v\n", err)
		return 2
	}
	out, err := formatter.Format("<stdin>", src, opt)
	if err != nil {
		fmt.Fprintf(stderr, "comment-formatter: %v\n", err)
		return 2
	}
	stdout.Write(out)
	return 0
}

func processFile(path string, stdout io.Writer, cfg config, opt formatter.Options) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out, err := formatter.Format(path, src, opt)
	if err != nil {
		return err
	}
	changed := !bytes.Equal(out, src)

	switch {
	case cfg.list:
		if changed {
			fmt.Fprintln(stdout, path)
		}
	case cfg.write:
		if changed {
			if err := os.WriteFile(path, out, 0o644); err != nil {
				return err
			}
		}
	default:
		stdout.Write(out)
	}
	return nil
}
