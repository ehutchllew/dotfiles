#!/bin/bash

COLOR_OFF="\e[0m"
GREEN="\e[1;32m"
YELLOW="\e[1;33m"
CYAN="\e[1;36m"
RED="\e[1;31m"

DOTFILES_DIR="$(cd "$(dirname "$0")" && pwd)"
STOW_PACKAGES=(alacritty lvim mise nvim tmux zsh)
NERD_FONTS=(DepartureMono)

log_info() { echo -e "${CYAN}$1${COLOR_OFF}"; }
log_success() { echo -e "${GREEN}$1${COLOR_OFF}"; }
log_warn() { echo -e "${YELLOW}$1${COLOR_OFF}"; }
log_error() { echo -e "${RED}$1${COLOR_OFF}"; }

# Detect Homebrew prefix based on platform/architecture
detect_brew_prefix() {
    if [[ "$(uname)" == "Darwin" && "$(uname -m)" == "arm64" ]]; then
        echo "/opt/homebrew"
    elif [[ "$(uname)" == "Darwin" && "$(uname -m)" == "x86_64" ]]; then
        echo "/usr/local"
    else
        echo "/home/linuxbrew/.linuxbrew"
    fi
}

BREW_PREFIX="$(detect_brew_prefix)"
BREW_BIN="${BREW_PREFIX}/bin/brew"

log_success "Let's get crackin'!"

# ============================================================
# Phase 1: Foundation (Homebrew + Stow)
# ============================================================
log_info "=== Phase 1: Foundation ==="

# Install Homebrew
if [[ -x "${BREW_BIN}" ]]; then
    log_success "Homebrew already installed."
else
    log_info "Installing Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    log_success "Homebrew installed!"
fi

# Ensure brew is in PATH for the remainder of this script
eval "$(${BREW_BIN} shellenv)"

# Install stow first — needed to symlink configs before other tools install
if command -v stow &>/dev/null; then
    log_success "Stow already installed."
else
    log_info "Installing Stow..."
    brew install stow
    log_success "Stow installed!"
fi

# ============================================================
# Phase 2: Stow Configs
# ============================================================
log_info "=== Phase 2: Stow Configs ==="

for package in "${STOW_PACKAGES[@]}"; do
    log_info "Stowing ${package}..."
    stow -d "${DOTFILES_DIR}" -t "${HOME}" -R "${package}"
done
log_success "All configs stowed!"

# ============================================================
# Phase 3: Install Tools
# ============================================================
log_info "=== Phase 3: Install Tools ==="

# Install Brew formulae from Brewfile
log_info "Installing Brew packages from Brewfile..."
brew bundle --file="${DOTFILES_DIR}/Brewfile"
log_success "Brew packages installed!"

# Install Mise
if [[ -x "$HOME/.local/bin/mise" ]]; then
    log_success "Mise already installed."
else
    log_info "Installing Mise..."
    curl https://mise.run | sh
    log_success "Mise installed!"
fi

# Install all tools defined in the (now stowed) mise config
log_info "Installing Mise tools..."
"$HOME/.local/bin/mise" install
log_success "Mise tools installed!"

# Install Nerd Fonts
if [[ "$(uname)" == "Darwin" ]]; then
    FONT_DIR="$HOME/Library/Fonts"
else
    FONT_DIR="$HOME/.local/share/fonts"
fi

for font in "${NERD_FONTS[@]}"; do
    if find "${FONT_DIR}" -name "*${font}*" -print -quit 2>/dev/null | grep -q .; then
        log_success "${font} Nerd Font already installed."
    else
        mkdir -p "${FONT_DIR}"
        log_info "Installing ${font} Nerd Font..."
        curl -fsSL -o /tmp/${font}.tar.xz \
            "https://github.com/ryanoasis/nerd-fonts/releases/latest/download/${font}.tar.xz"
        tar -xf /tmp/${font}.tar.xz -C "${FONT_DIR}"
        rm -f /tmp/${font}.tar.xz
        log_success "${font} Nerd Font installed!"
    fi
done

if [[ "$(uname)" != "Darwin" ]]; then
    fc-cache -fv
fi

# ============================================================
# Phase 4: Plugin Managers
# ============================================================
log_info "=== Phase 4: Plugin Managers ==="

# Install Zap-Zsh (--keep prevents overwriting the stowed .zshrc)
if [[ -f "$HOME/.local/share/zap/zap.zsh" ]]; then
    log_success "Zap-Zsh already installed."
else
    log_info "Installing Zap-Zsh..."
    zsh <(curl -s https://raw.githubusercontent.com/zap-zsh/zap/master/install.zsh) --branch release-v1 --keep
    log_success "Zap-Zsh installed!"
fi

# Install Tmux Plugin Manager
if [[ -d "$HOME/.tmux/plugins/tpm" ]]; then
    log_success "TPM already installed."
else
    log_info "Installing Tmux Plugin Manager..."
    git clone https://github.com/tmux-plugins/tpm "$HOME/.tmux/plugins/tpm"
    log_success "TPM installed!"
fi

log_success "=== All Done! ==="
