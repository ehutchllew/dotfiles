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

BREW_BIN="/home/linuxbrew/.linuxbrew/bin"
if [[ $(uname) == "Darwin" && $(uname -m) == "arm64" ]]; then
	BREW_BIN="/opt/homebrew/bin"
elif [[ $(uname) == "Darwin" && $(uname -m) == "x86_64" ]]; then
	BREW_BIN="/usr/local/bin"
else
	BREW_BIN="/home/linuxbrew/.linuxbrew/bin"
fi

# Define Paths
CARGO_BIN="$HOME/.cargo/bin"
LOCAL_BIN="$HOME/.local/bin"
PULUMI_BIN="$HOME/.pulumi/bin"

# Build PATH
NEW_PATH=""
add_to_path "${LOCAL_BIN}"
add_to_path "${BREW_BIN}"
add_to_path "${CARGO_BIN}"
add_to_path "${PULUMI_BIN}"

# Export PATH
if [[ -n "${NEW_PATH}" ]]; then
	export PATH="$NEW_PATH:$PATH"
fi
	


# EDITOR
export EDITOR="nvim"

# FZF
export FZF_DEFAULT_OPTS="--tmux 80% --layout=reverse --border"
