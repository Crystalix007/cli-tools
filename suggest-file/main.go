// Package main provides the CLI entry point for suggest-file, a tool that
// lists files matching glob patterns for use with interactive selectors like fzf.
// When no arguments are given, it recursively lists all files in the current directory.
package main

import (
	"fmt"
	"os"

	"github.com/Crystalix007/cli-tools/suggest-file/glob"
	"github.com/Crystalix007/cli-tools/suggest-file/walker"
)

func main() {
	args := os.Args[1:]

	// Handle help flag.
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Println("Usage: suggest-file [PATTERN ...]")
		fmt.Println("       suggest-file --shell bash|zsh")
		fmt.Println("List files matching patterns. With no arguments, list all files recursively.")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --shell bash|zsh  Enable suggest-file completion in your shell.")
		fmt.Println("                    Prints a snippet to stdout; source it in your rc file.")
		fmt.Println("")
		fmt.Println("Argument resolution:")
		fmt.Println("  DIRECTORY          Recursively list all files under it (e.g. ~, ~/Downloads, .)")
		fmt.Println("  PATH/             Trailing slash: list files in that directory recursively")
		fmt.Println("  PREFIX            Match entries starting with prefix, walk matching dirs")
		fmt.Println("  GLOB              Standard glob with *, ?, [, ** for recursive matching")
		fmt.Println("")
		fmt.Println("Supports ~ expansion to home directory.")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  suggest-file                     # list all files in current directory")
		fmt.Println("  suggest-file ~                   # all files under home directory")
		fmt.Println("  suggest-file ~/Downloads          # all files under ~/Downloads recursively")
		fmt.Println("  suggest-file ~/.config/           # files directly in ~/.config (one level)")
		fmt.Println("  suggest-file ~/D                  # files under ~/Downloads, ~/Documents, etc.")
		fmt.Println("  suggest-file '~/.config/*.yaml'  # yaml files in ~/.config")
		fmt.Println("  suggest-file '**/*.go'           # all Go files recursively")
		return
	}

	// Handle --shell flag.
	if len(args) >= 1 && args[0] == "--shell" {
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "suggest-file: --shell requires an argument (bash or zsh)")
			os.Exit(1)
		}
		script, ok := shellWrapper(args[1])
		if !ok {
			fmt.Fprintf(os.Stderr, "suggest-file: unknown shell %q (supported: bash, zsh)\n", args[1])
			os.Exit(1)
		}
		fmt.Print(script)
		return
	}

	// If no arguments provided, default to recursive listing of the current directory.
	if len(args) == 0 {
		if err := walker.Walk("."); err != nil {
			fmt.Fprintf(os.Stderr, "suggest-file: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Expand each glob pattern and print matching file paths.
	exitCode := 0
	for _, pattern := range args {
		matches, err := glob.Expand(pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "suggest-file: %s: %v\n", pattern, err)
			exitCode = 1
			continue
		}
		if len(matches) == 0 {
			fmt.Fprintf(os.Stderr, "suggest-file: %s: no matches\n", pattern)
		}
		for _, match := range matches {
			fmt.Println(match)
		}
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
