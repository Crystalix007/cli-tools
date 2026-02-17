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

// Walk recursively walks the directory tree rooted at root and prints
// every regular file path to stdout, one per line. Directories that
// cannot be read (e.g. due to permission errors) are skipped with a
// warning on stderr rather than aborting.
func Walk(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Report the error but keep walking.
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

		// For symlinks, resolve to check target type.
		if d.Type()&fs.ModeSymlink != 0 {
			resolved, err := os.Stat(path)
			if err != nil {
				// Dangling symlink — skip silently.
				return nil
			}
			if resolved.IsDir() {
				return nil
			}
			fmt.Println(path)
			return nil
		}

		// Only print regular files.
		if !d.Type().IsRegular() {
			return nil
		}

		fmt.Println(path)
		return nil
	})
}
