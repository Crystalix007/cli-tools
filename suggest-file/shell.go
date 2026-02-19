package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell [bash|zsh]",
	Short: "Enable suggest-file completion in your shell",
	Long:  "Enable suggest-file completion in your shell (bash|zsh). Prints a snippet to stdout; source it in your rc file.",
	Args:  cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh"},
	RunE: func(cmd *cobra.Command, args []string) error {
		shellName := args[0]
		script, ok := shellWrapper(shellName)
		if !ok {
			return fmt.Errorf("unknown shell %q (supported: bash, zsh)", shellName)
		}
		fmt.Print(script)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
}
