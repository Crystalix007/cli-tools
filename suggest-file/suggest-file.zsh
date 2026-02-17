# suggest-file.zsh -- Source this file to bind Ctrl-T to suggest-file + fzf in ZSH.
#
# Requires: suggest-file on $PATH, fzf on $PATH.
# Optional: bat for file preview.
#
# Environment variables for customisation:
#   SUGGEST_FILE_OPTS      -- extra arguments passed to suggest-file
#   SUGGEST_FILE_FZF_OPTS  -- extra arguments passed to fzf (overrides defaults)
#   SUGGEST_FILE_BAT_OPTS  -- extra arguments passed to bat in the preview

# ---------------------------------------------------------------------------
# Defaults
# ---------------------------------------------------------------------------

# Default fzf options: multi-select, inline preview via bat.
__suggest_file_fzf_default_opts="--multi --ansi --preview-window=right:60%:wrap"

# Default bat options: colour always, line numbers, plain style.
__suggest_file_bat_default_opts="--color=always --line-range=:500 --style=numbers,changes"

# ---------------------------------------------------------------------------
# Widget
# ---------------------------------------------------------------------------

# __suggest_file_widget runs suggest-file, pipes results through fzf, and
# inserts the selected path(s) into the current command line.
__suggest_file_widget() {
  local sf_opts="${SUGGEST_FILE_OPTS:-}"
  local fzf_opts="${SUGGEST_FILE_FZF_OPTS:-$__suggest_file_fzf_default_opts}"
  local bat_opts="${SUGGEST_FILE_BAT_OPTS:-$__suggest_file_bat_default_opts}"

  # Build the fzf preview command using bat.
  local preview_cmd="bat ${bat_opts} -- {}"

  # Run suggest-file and pipe into fzf.
  # ${=var} performs word-splitting in zsh (safe alternative to eval).
  # suggest-file errors are shown on stderr; fzf stderr is suppressed.
  local result
  result="$(
    suggest-file ${=sf_opts} \
      | fzf ${=fzf_opts} --preview="${preview_cmd}" \
      2>/dev/null
  )"

  if [[ -n "$result" ]]; then
    # Quote each selected path and join with spaces.
    local quoted=""
    while IFS= read -r line; do
      quoted+="${(q)line} "
    done <<< "$result"

    # Insert the selected file(s) at the cursor position.
    LBUFFER+="${quoted% }"
  fi

  # Redraw the prompt.
  zle reset-prompt
}

# Register the widget and bind it to Ctrl-T.
zle -N __suggest_file_widget
bindkey '^T' __suggest_file_widget
