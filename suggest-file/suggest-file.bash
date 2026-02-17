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
# inserts the selected path(s) into the current command line at the cursor.
__suggest_file_widget() {
  local sf_opts="${SUGGEST_FILE_OPTS:-}"
  local fzf_opts="${SUGGEST_FILE_FZF_OPTS:-$__suggest_file_fzf_default_opts}"
  local bat_opts="${SUGGEST_FILE_BAT_OPTS:-$__suggest_file_bat_default_opts}"

  # Build the fzf preview command using bat.
  local preview_cmd="bat ${bat_opts} -- {}"

  # Run suggest-file and pipe into fzf.
  # Word splitting on $sf_opts and $fzf_opts is intentional here to allow
  # multiple flags to be passed via a single string variable.
  local result
  result="$(
    suggest-file $sf_opts \
      | fzf $fzf_opts --preview="${preview_cmd}" \
      2>/dev/null
  )"

  if [[ -n "$result" ]]; then
    # Quote each selected path and join with spaces.
    local quoted=""
    while IFS= read -r line; do
      printf -v escaped '%q' "$line"
      quoted+="${escaped} "
    done <<< "$result"

    # Remove the trailing space.
    quoted="${quoted% }"

    # Insert the selected file(s) at the cursor position.
    local left="${READLINE_LINE:0:$READLINE_POINT}"
    local right="${READLINE_LINE:$READLINE_POINT}"
    READLINE_LINE="${left}${quoted}${right}"
    READLINE_POINT=$(( ${#left} + ${#quoted} ))
  fi
}

# Bind the widget to Ctrl-T.
bind -x '"\C-t": __suggest_file_widget'
