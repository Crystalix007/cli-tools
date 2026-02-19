// Package main provides the CLI entry point for suggest-file, a tool that
// lists files matching glob patterns for use with interactive selectors like fzf.
// When no arguments are given, it recursively lists all files in the current directory.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Crystalix007/cli-tools/suggest-file/glob"
	"github.com/Crystalix007/cli-tools/suggest-file/walker"
)

var shellFlag string

var rootCmd = &cobra.Command{
	Use:   "suggest-file [PATTERN ...]",
	Short: "List files matching patterns. With no arguments, list all files recursively.",
	Long: `List files matching patterns. With no arguments, list all files recursively.

Argument resolution:
  DIRECTORY          Recursively list all files under it (e.g. ~, ~/Downloads, .)
  PATH/             Trailing slash: list files in that directory recursively
  PREFIX            Match entries starting with prefix, walk matching dirs
  GLOB              Standard glob with *, ?, [, ** for recursive matching

Supports ~ expansion to home directory.`,
	Example: `  suggest-file                     # list all files in current directory
  suggest-file ~                   # all files under home directory
  suggest-file ~/Downloads          # all files under ~/Downloads recursively
  suggest-file ~/.config/           # files directly in ~/.config (one level)
  suggest-file ~/D                  # files under ~/Downloads, ~/Documents, etc.
  suggest-file '~/.config/*.yaml'  # yaml files in ~/.config
  suggest-file '**/*.go'           # all Go files recursively`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if shellFlag != "" {
			script, ok := shellWrapper(shellFlag)
			if !ok {
				return fmt.Errorf("unknown shell %q (supported: bash, zsh)", shellFlag)
			}
			fmt.Print(script)
			return nil
		}

		// If no arguments provided, default to recursive listing of the current directory.
		if len(args) == 0 {
			if err := walker.Walk("."); err != nil {
				return err
			}
			return nil
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
		return nil
	},
}

func init() {
	rootCmd.Flags().StringVar(&shellFlag, "shell", "", "Enable suggest-file completion in your shell (bash|zsh). Prints a snippet to stdout; source it in your rc file.")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
