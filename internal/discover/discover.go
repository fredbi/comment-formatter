// Package discover expands CLI path arguments into the set of .go files to
// format.
//
// Directories are walked recursively; vendor, testdata, .git and other
// dot-directories are skipped.
package discover

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// skipDir names directory entries that are never descended into.
var skipDir = map[string]bool{
	"vendor":   true,
	"testdata": true,
}

// Files expands the given path arguments (files or directories) into a sorted,
// de-duplicated list of .go file paths.
//
// A file argument is always included regardless of name; directory arguments
// are walked recursively.
func Files(args []string) ([]string, error) {
	seen := make(map[string]bool)
	var out []string
	add := func(p string) {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}

	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			add(arg)
			continue
		}
		err = filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path == arg {
					return nil
				}
				name := d.Name()
				if skipDir[name] || (strings.HasPrefix(name, ".") && name != ".") {
					return fs.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(d.Name(), ".go") {
				add(path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Strings(out)
	return out, nil
}
