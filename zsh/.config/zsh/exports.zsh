function add_to_path() {
	local bin_path="$1"
	if [[ -d "${bin_path}" ]]; then
      if [[ -z "${NEW_PATH}" ]]; then
          NEW_PATH="${bin_path}"
      else
          NEW_PATH="${NEW_PATH}:${bin_path}"
      fi
  else
      echo "Warning: Bin Path Directory Not Found: ${bin_path}" >&2
  fi
}

# Define Paths
LOCAL_BIN="$HOME/.local/bin"

# Build PATH
NEW_PATH=""
if [[ -d "${LOCAL_BIN}" ]]; then
    add_to_path "${LOCAL_BIN}"
fi

# Export PATH
if [[ -n "${NEW_PATH}" ]]; then
	export PATH="$NEW_PATH:$PATH"
fi
	


# EDITOR
export EDITOR="nvim"

# FZF
export FZF_DEFAULT_OPTS="--tmux 80% --layout=reverse --border"
