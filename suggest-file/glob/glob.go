// Package glob provides argument resolution logic for suggest-file.
// It supports tilde (~) expansion to the user's home directory,
// recursive globbing via ** using the doublestar library, directory
// listing, and prefix-based matching.
//
// Resolution precedence for a given argument:
//  1. Expand tilde (~) to the user's home directory.
//  2. If the argument contains glob metacharacters (*, ?, [, {) → doublestar
//     glob expansion.
//  3. If the argument ends with '/' → list files in that directory only
//     (one level deep, non-recursive).
//  4. If the argument resolves to an existing regular file → return it.
//  5. If the argument resolves to an existing directory → recursively walk
//     it using the walker package.
//  6. Otherwise → prefix match: the last path component is treated as a
//     prefix. Entries matching "<pattern>*" are found; matching directories
//     are walked recursively, matching files are included directly.
//
// Note: the doublestar library does not match hidden files/directories
// (names starting with '.') unless the pattern explicitly starts with '.'.
// This differs from the walker package which skips hidden directories but
// includes hidden files within visible directories. When using explicit
// glob patterns, hidden entries must be targeted via patterns like '.*' or
// '.config/**'.
package glob

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/Crystalix007/cli-tools/suggest-file/walker"
)

// Expand takes an argument string and returns all matching file paths.
// See package documentation for the full resolution logic.
func Expand(pattern string) ([]string, error) {
	// Step 1: Expand tilde.
	expanded, err := expandTilde(pattern)
	if err != nil {
		return nil, err
	}

	// Remember whether the original argument had a trailing slash before
	// any cleaning occurs. expandTilde preserves trailing slashes.
	trailingSlash := strings.HasSuffix(expanded, "/")

	// Step 2: If the pattern contains glob metacharacters, use doublestar.
	if containsMeta(expanded) {
		return expandGlob(expanded)
	}

	// Clean the path for filesystem operations, but only after checking
	// the trailing slash (filepath.Clean strips it).
	cleaned := filepath.Clean(expanded)

	// Step 3: Trailing slash → list files in that directory (one level).
	if trailingSlash {
		return walker.ListDir(cleaned)
	}

	// Step 4/5: Stat the path to determine if it's a file or directory.
	info, err := os.Stat(cleaned)
	if err == nil {
		if info.IsDir() {
			return walker.WalkCollect(cleaned)
		}
		if info.Mode().IsRegular() {
			// Exact file match — return it directly.
			return []string{cleaned}, nil
		}
	} else if !os.IsNotExist(err) {
		// Real error (e.g. permission denied) — surface it instead of
		// silently falling through to prefix matching.
		return nil, fmt.Errorf("stat %q: %w", cleaned, err)
	}

	// Step 6: Prefix match — treat the last component as a prefix.
	return expandPrefix(cleaned)
}

// expandGlob performs doublestar glob expansion on a pattern that contains
// metacharacters. Only regular files (and symlinks resolving to regular files)
// are included in the results.
func expandGlob(pattern string) ([]string, error) {
	// Split the pattern into a base directory and the glob portion.
	// This allows doublestar to work correctly with absolute and relative paths.
	base, globPart := splitPattern(pattern)

	fsys := os.DirFS(base)
	matches, err := doublestar.Glob(fsys, globPart)
	if err != nil {
		return nil, fmt.Errorf("glob %q: %w", pattern, err)
	}

	// Reconstruct full paths by joining the base back.
	results := make([]string, 0, len(matches))
	for _, m := range matches {
		full := filepath.Join(base, m)

		info, err := os.Lstat(full)
		if err != nil {
			// Skip files we cannot stat (e.g. permission denied).
			continue
		}

		if walker.IsIncludableFile(full, info.Mode()) {
			results = append(results, full)
		}
	}

	return results, nil
}

// expandPrefix treats the last component of path as a prefix and finds all
// entries in the parent directory that start with it. Matching directories
// are walked recursively; matching regular files are included directly.
func expandPrefix(path string) ([]string, error) {
	dir := filepath.Dir(path)
	prefix := filepath.Base(path)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %q: %w", dir, err)
	}

	var results []string

	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), prefix) {
			continue
		}

		full := filepath.Join(dir, e.Name())

		// Check if this is a directory (resolving symlinks).
		isDir := false
		if e.Type()&os.ModeSymlink != 0 {
			resolved, err := os.Stat(full)
			if err != nil {
				// Dangling symlink — skip.
				continue
			}
			isDir = resolved.IsDir()
		} else {
			isDir = e.IsDir()
		}

		if isDir {
			// Directory matching prefix — walk recursively.
			collected, err := walker.WalkCollect(full)
			if err != nil {
				return nil, err
			}
			results = append(results, collected...)
			continue
		}

		if walker.IsIncludableFile(full, e.Type()) {
			results = append(results, full)
		}
	}

	return results, nil
}

// expandTilde replaces a leading ~ or ~/ with the user's home directory.
// A bare "~" is expanded to the home directory path.
// Returns an error if ~ is used but the home directory cannot be determined.
func expandTilde(pattern string) (string, error) {
	if pattern == "~" || strings.HasPrefix(pattern, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expanding ~: %w", err)
		}
		if len(pattern) > 1 {
			return home + pattern[1:], nil
		}
		return home, nil
	}
	return pattern, nil
}

// splitPattern splits a glob pattern into a static base directory and the
// remaining glob expression. The base is the longest prefix of path components
// that contain no glob meta-characters.
func splitPattern(pattern string) (base, glob string) {
	// Clean the pattern to normalise separators.
	pattern = filepath.Clean(pattern)

	isAbs := filepath.IsAbs(pattern)
	parts := strings.Split(pattern, string(filepath.Separator))

	// On Unix, absolute paths produce an empty first element from the
	// leading separator — skip it when scanning for metacharacters.
	start := 0
	if isAbs && len(parts) > 0 && parts[0] == "" {
		start = 1
	}

	// Find the first component that contains a glob meta-character.
	idx := start
	for idx < len(parts) {
		if containsMeta(parts[idx]) {
			break
		}
		idx++
	}

	if idx == start {
		// The very first meaningful component is a glob.
		if isAbs {
			return string(filepath.Separator), strings.Join(parts[start:], string(filepath.Separator))
		}
		return ".", pattern
	}

	base = strings.Join(parts[:idx], string(filepath.Separator))
	if base == "" {
		// Absolute path on Unix: the join of an empty first element gives "".
		base = string(filepath.Separator)
	}

	if idx == len(parts) {
		// All metacharacters were eliminated by filepath.Clean (e.g. "foo/*/..").
		// Fall back to matching the basename literally in its parent directory.
		dir := filepath.Dir(pattern)
		file := filepath.Base(pattern)
		return dir, file
	}

	glob = strings.Join(parts[idx:], string(filepath.Separator))
	return base, glob
}

// containsMeta reports whether s contains any glob metacharacters.
// Includes '{' for doublestar's brace/alternation syntax.
func containsMeta(s string) bool {
	return strings.ContainsAny(s, "*?[{")
}
