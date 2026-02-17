// Package walker provides a recursive directory walker for suggest-file.
// It is used when no glob arguments are given, listing all files under a
// root directory (the default Ctrl-T behaviour).
//
// Hidden directories (names starting with '.') are skipped, but hidden
// files within visible directories are included.
package walker

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// IsIncludableFile reports whether path should be included in results.
// It resolves symlinks and returns true only for regular files (or
// symlinks that resolve to regular files). Dangling symlinks, directories,
// and special files (pipes, sockets, devices) return false.
func IsIncludableFile(path string, mode fs.FileMode) bool {
	if mode&fs.ModeSymlink != 0 {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			return false
		}
		return info.Mode().IsRegular()
	}
	return mode.IsRegular()
}

// walkFiltered is the shared implementation for Walk and WalkCollect.
// It recursively walks root, skipping hidden directories and non-regular
// files, and calls emit for each included path.
func walkFiltered(root string, emit func(path string)) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "suggest-file: %v\n", err)
			return nil
		}

		// Skip hidden directories (starting with '.') other than the root itself.
		if d.IsDir() && path != root && d.Name()[0] == '.' {
			return fs.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		if IsIncludableFile(path, d.Type()) {
			emit(path)
		}

		return nil
	})
}

// Walk recursively walks the directory tree rooted at root and prints
// every regular file path to stdout, one per line. Directories that
// cannot be read (e.g. due to permission errors) are skipped with a
// warning on stderr rather than aborting.
func Walk(root string) error {
	return walkFiltered(root, func(path string) {
		fmt.Println(path)
	})
}

// WalkCollect recursively walks the directory tree rooted at root and
// returns all regular file paths as a slice. Like Walk, hidden directories
// are skipped, but hidden files within visible directories are included.
// Errors reading individual entries are reported on stderr and skipped.
func WalkCollect(root string) ([]string, error) {
	var results []string
	err := walkFiltered(root, func(path string) {
		results = append(results, path)
	})
	return results, err
}

// ListDir lists files directly within dir (one level deep, non-recursive).
// Only regular files and symlinks that resolve to files are included.
// Hidden entries are included (consistent with explicit directory listing).
func ListDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("listing %q: %w", dir, err)
	}

	var results []string

	for _, e := range entries {
		full := filepath.Join(dir, e.Name())

		if IsIncludableFile(full, e.Type()) {
			results = append(results, full)
		}
	}

	return results, nil
}
