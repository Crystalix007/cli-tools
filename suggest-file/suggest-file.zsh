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
# replaces the current word on the command line with the selected path(s).
__suggest_file_widget() {
  local sf_opts="${SUGGEST_FILE_OPTS:-}"
  local fzf_opts="${SUGGEST_FILE_FZF_OPTS:-$__suggest_file_fzf_default_opts}"
  local bat_opts="${SUGGEST_FILE_BAT_OPTS:-$__suggest_file_bat_default_opts}"

  # Extract the current word (token left of the cursor, delimited by spaces).
  # This is passed to suggest-file as a glob pattern.
  local current_word="${LBUFFER##* }"
  local prefix="${LBUFFER%$current_word}"

  # Build the fzf preview command using bat.
  local preview_cmd="bat ${bat_opts} -- {}"

  # If the current word is non-empty, use it as the suggest-file pattern.
  # Otherwise, suggest-file runs with no args (recursive listing).
  local result
  if [[ -n "$current_word" ]]; then
    result="$(
      suggest-file ${=sf_opts} "$current_word" \
        | fzf ${=fzf_opts} --preview="${preview_cmd}" \
          --query="${current_word:t}" \
        2>/dev/null
    )"
  else
    result="$(
      suggest-file ${=sf_opts} \
        | fzf ${=fzf_opts} --preview="${preview_cmd}" \
        2>/dev/null
    )"
  fi

  if [[ -n "$result" ]]; then
    # Quote each selected path and join with spaces.
    local quoted=""
    while IFS= read -r line; do
      quoted+="${(q)line} "
    done <<< "$result"

    # Replace the current word with the selected file(s).
    LBUFFER="${prefix}${quoted% }"
  fi

  # Redraw the prompt.
  zle reset-prompt
}

# Register the widget and bind it to Ctrl-T.
zle -N __suggest_file_widget
bindkey '^T' __suggest_file_widget
