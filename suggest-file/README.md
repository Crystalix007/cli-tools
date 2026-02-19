# suggest-file

A file finder that supports glob patterns with `~` expansion and `**` recursive matching, designed to feed into [`fzf`](https://github.com/junegunn/fzf) with [`bat`](https://github.com/sharkdp/bat) previews.

## Install

```sh
go install github.com/Crystalix007/cli-tools/suggest-file@latest
```

## Usage

```sh
suggest-file                     # list all files recursively (default Ctrl-T)
suggest-file '~/.config/*.yaml'  # yaml files in ~/.config
suggest-file '**/*.go'           # all Go files recursively
suggest-file '/usr/local/bin/*'  # files in an absolute path
```

## Shell integration

Source the appropriate script to bind **Ctrl-T** to fuzzy file selection:

```sh
# ZSH
source <(suggest-file shell zsh)

# Bash (>= 4.0)
source <(suggest-file shell bash)
```

### Customisation

Override defaults via environment variables:

| Variable | Purpose | Default |
|----------|---------|---------|
| `SUGGEST_FILE_OPTS` | Arguments passed to `suggest-file` | *(none)* |
| `SUGGEST_FILE_FZF_OPTS` | Arguments passed to `fzf` | `--multi --ansi --preview-window=right:60%:wrap` |
| `SUGGEST_FILE_BAT_OPTS` | Arguments passed to `bat` | `--color=always --line-range=:500 --style=numbers,changes` |

```sh
export SUGGEST_FILE_FZF_OPTS="--multi --height=40%"
export SUGGEST_FILE_BAT_OPTS="--color=always --theme=Dracula"
```
