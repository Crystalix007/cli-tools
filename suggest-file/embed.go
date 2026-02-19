package main

import _ "embed"

//go:embed suggest-file.bash
var bashWrapper string

//go:embed suggest-file.zsh
var zshWrapper string

// shellWrapper returns the embedded shell wrapper script for the given shell
// name. The second return value is false if the shell is not recognised.
func shellWrapper(shell string) (string, bool) {
	switch shell {
	case "bash":
		return bashWrapper, true
	case "zsh":
		return zshWrapper, true
	default:
		return "", false
	}
}
