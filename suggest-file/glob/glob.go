// Package glob provides glob expansion logic for suggest-file.
// It supports tilde (~) expansion to the user's home directory and
// recursive globbing via ** using the doublestar library.
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
)

// Expand takes a glob pattern string and returns all matching file paths.
// It handles:
//   - ~ expansion to the user's home directory
//   - ** recursive directory matching via doublestar
//   - Standard glob characters (*, ?, [...])
func Expand(pattern string) ([]string, error) {
	expanded, err := expandTilde(pattern)
	if err != nil {
		return nil, err
	}
	pattern = expanded

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

		// Only include regular files and symlinks that resolve to files.
		info, err := os.Lstat(full)
		if err != nil {
			// Skip files we cannot stat (e.g. permission denied).
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			// Resolve symlink to check if target is a file.
			resolved, err := os.Stat(full)
			if err != nil {
				// Dangling symlink — skip.
				continue
			}
			if resolved.IsDir() {
				continue
			}
		} else if info.IsDir() {
			continue
		}

		results = append(results, full)
	}

	return results, nil
}

// expandTilde replaces a leading ~ or ~/ with the user's home directory.
// Returns an error if ~ is used but the home directory cannot be determined.
func expandTilde(pattern string) (string, error) {
	if pattern == "~" || strings.HasPrefix(pattern, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expanding ~: %w", err)
		}
		return filepath.Join(home, pattern[1:]), nil
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
		// No meta-characters at all — treat as a literal path.
		// Return the directory and the filename as the glob.
		dir := filepath.Dir(pattern)
		file := filepath.Base(pattern)
		return dir, file
	}

	glob = strings.Join(parts[idx:], string(filepath.Separator))
	return base, glob
}

// containsMeta reports whether s contains any glob metacharacters.
func containsMeta(s string) bool {
	return strings.ContainsAny(s, "*?[")
}
