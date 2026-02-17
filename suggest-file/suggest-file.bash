# suggest-file.bash -- Source this file to bind Ctrl-T to suggest-file + fzf in Bash.
#
# Requires: bash >= 4.0, suggest-file on $PATH, fzf on $PATH.
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
  local left="${READLINE_LINE:0:$READLINE_POINT}"
  local right="${READLINE_LINE:$READLINE_POINT}"
  local current_word="${left##* }"
  local prefix="${left%$current_word}"

  # Build the fzf preview command using bat.
  local preview_cmd="bat ${bat_opts} -- {}"

  # If the current word is non-empty, use it as the suggest-file pattern.
  # Otherwise, suggest-file runs with no args (recursive listing).
  # Word splitting on $sf_opts and $fzf_opts is intentional to allow
  # multiple flags via a single string variable.
  local result
  if [[ -n "$current_word" ]]; then
    result="$(
      suggest-file $sf_opts "$current_word" \
        | fzf $fzf_opts --preview="${preview_cmd}" \
        2>/dev/null
    )"
  else
    result="$(
      suggest-file $sf_opts \
        | fzf $fzf_opts --preview="${preview_cmd}" \
        2>/dev/null
    )"
  fi

  if [[ -n "$result" ]]; then
    # Quote each selected path and join with spaces.
    local quoted=""
    while IFS= read -r line; do
      printf -v escaped '%q' "$line"
      quoted+="${escaped} "
    done <<< "$result"

    # Remove the trailing space.
    quoted="${quoted% }"

    # Replace the current word with the selected file(s).
    READLINE_LINE="${prefix}${quoted}${right}"
    READLINE_POINT=$(( ${#prefix} + ${#quoted} ))
  fi
}

# Bind the widget to Ctrl-T.
bind -x '"\C-t": __suggest_file_widget'
