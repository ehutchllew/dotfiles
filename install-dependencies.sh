#!/bin/bash
set -euo pipefail

COLOR_OFF="\e[0m"
GREEN="\e[1;32m"
YELLOW="\e[1;33m"
CYAN="\e[1;36m"
RED="\e[1;31m"

DOTFILES_DIR="$(cd "$(dirname "$0")" && pwd)"
SHARED_STOW_PACKAGES=(mise nvim tmux zsh)
NERD_FONTS=(DepartureMono)
TEMP_DIR=""

log_info() { echo -e "${CYAN}$1${COLOR_OFF}"; }
log_success() { echo -e "${GREEN}$1${COLOR_OFF}"; }
log_warn() { echo -e "${YELLOW}$1${COLOR_OFF}"; }
log_error() { echo -e "${RED}$1${COLOR_OFF}" >&2; }

cleanup() {
    if [[ -n "$TEMP_DIR" && -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

require_command() {
    if ! command -v "$1" >/dev/null 2>&1; then
        log_error "Required command not found: $1"
        exit 1
    fi
}

# The kernel and architecture come from uname; Linux distribution details do not.
case "$(uname -s)" in
    Darwin) PLATFORM=macos ;;
    Linux)
        if [[ ! -r /etc/os-release ]]; then
            log_error "Cannot identify Linux distribution: /etc/os-release is missing."
            exit 1
        fi
        # os-release is specified as shell-compatible variable assignments.
        . /etc/os-release
        case " ${ID:-} ${ID_LIKE:-} " in
            *" arch "*) PLATFORM=arch ;;
            *" debian "*|*" ubuntu "*) PLATFORM=debian ;;
            *) log_error "Unsupported Linux distribution: ${ID:-unknown} (${ID_LIKE:-no ID_LIKE})."; exit 1 ;;
        esac
        ;;
    *) log_error "Unsupported operating system: $(uname -s)."; exit 1 ;;
esac

case "$(uname -m)" in
    x86_64|amd64) ARCH=x86_64 ;;
    arm64|aarch64) ARCH=arm64 ;;
    *) log_error "Unsupported architecture: $(uname -m)."; exit 1 ;;
esac

log_info "Platform: ${PLATFORM}, architecture: ${ARCH}"
if [[ "$PLATFORM" == macos && "$ARCH" != arm64 ]]; then
    if [[ "$(sysctl -n sysctl.proc_translated 2>/dev/null || true)" == 1 ]]; then
        log_error "This terminal is running under Rosetta (Intel mode). Re-run from a native Apple Silicon terminal."
    else
        log_error "Only Apple Silicon Macs are supported."
    fi
    exit 1
fi
if [[ "$ARCH" == arm64 && "$PLATFORM" != macos ]]; then
    log_warn "ARM64 package and tool availability varies by distribution; some packages may be skipped or need removing."
fi

# Platform-specific stow packages, added to the shared ones. These can also
# depend on ARCH if a config only applies to one architecture.
case "$PLATFORM" in
    macos) STOW_PACKAGES=("${SHARED_STOW_PACKAGES[@]}" alacritty zsh-macos) ;;
    debian) STOW_PACKAGES=("${SHARED_STOW_PACKAGES[@]}" alacritty zsh-debian) ;;
    arch) STOW_PACKAGES=("${SHARED_STOW_PACKAGES[@]}" hypr waybar rofi zsh-arch) ;;
esac

if [[ "$EUID" -eq 0 ]]; then
    log_error "Run this script as your normal user, not root."
    exit 1
fi

# ============================================================
# Phase 0: Platform packages and prerequisites
# ============================================================
log_info "=== Phase 0: Platform Packages ==="

case "$PLATFORM" in
    arch)
        require_command pacman
        require_command sudo
        log_warn "The Arch package list includes workstation, desktop, gaming, and AUR packages."
        log_warn "The Arch installer will perform a pacman system upgrade (pacman -Syu)."
        bash "${DOTFILES_DIR}/bootstrap/arch/install-pkgs.sh"
        ;;
    debian)
        require_command apt-get
        require_command dpkg-query
        require_command sudo
        MISSING_PKGS=()
        for package in zsh build-essential pkg-config curl git libssl-dev fontconfig xz-utils ca-certificates procps file alacritty; do
            if [[ "$(dpkg-query -W -f='${Status}' "$package" 2>/dev/null || true)" != "install ok installed" ]]; then
                MISSING_PKGS+=("$package")
            fi
        done
        if [[ ${#MISSING_PKGS[@]} -gt 0 ]]; then
            log_info "Installing missing packages: ${MISSING_PKGS[*]}"
            sudo apt-get update
            sudo apt-get install -y "${MISSING_PKGS[@]}"
        else
            log_success "System prerequisites already installed."
        fi
        ;;
    macos)
        require_command curl
        require_command git
        ;;
esac

for command in curl git tar zsh; do
    require_command "$command"
done
if [[ "$PLATFORM" != macos ]]; then
    require_command fc-cache
fi
TEMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/dotfiles-bootstrap.XXXXXXXX")"

# ============================================================
# Phase 1: Foundation (Homebrew where applicable + Stow)
# ============================================================
log_info "=== Phase 1: Foundation ==="

if [[ "$PLATFORM" != arch ]]; then
    if command -v brew >/dev/null 2>&1; then
        BREW_BIN="$(command -v brew)"
    else
        case "$PLATFORM" in
            macos) BREW_BIN=/opt/homebrew/bin/brew ;;
            debian) BREW_BIN=/home/linuxbrew/.linuxbrew/bin/brew ;;
        esac
    fi

    if [[ ! -x "$BREW_BIN" ]]; then
        log_info "Installing Homebrew..."
        curl -fsSL -o "${TEMP_DIR}/install-homebrew.sh" https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh
        /bin/bash "${TEMP_DIR}/install-homebrew.sh"
    fi
    if [[ ! -x "$BREW_BIN" ]]; then
        log_error "Homebrew not found at ${BREW_BIN}; check its installation and platform support."
        exit 1
    fi
    BREW_SHELLENV="$("$BREW_BIN" shellenv)"
    eval "$BREW_SHELLENV"

    if ! command -v stow >/dev/null 2>&1; then
        brew install stow
    fi
fi
require_command stow

# ============================================================
# Phase 2: Stow Configs
# ============================================================
log_info "=== Phase 2: Stow Configs ==="
for package in "${STOW_PACKAGES[@]}"; do
    log_info "Checking ${package} for conflicts..."
    stow -n -d "$DOTFILES_DIR" -t "$HOME" -R "$package"
done
for package in "${STOW_PACKAGES[@]}"; do
    log_info "Stowing ${package}..."
    stow -d "$DOTFILES_DIR" -t "$HOME" -R "$package"
done
log_success "All configs stowed!"

# ============================================================
# Phase 3: Tools and fonts
# ============================================================
log_info "=== Phase 3: Install Tools ==="
if [[ "$PLATFORM" != arch ]]; then
    brew bundle --file="${DOTFILES_DIR}/bootstrap/brew/Brewfile"
fi

if command -v mise >/dev/null 2>&1; then
    MISE_BIN="$(command -v mise)"
elif [[ -x "$HOME/.local/bin/mise" ]]; then
    MISE_BIN="$HOME/.local/bin/mise"
elif [[ "$PLATFORM" == arch ]]; then
    log_error "Mise was not found after installing the Arch package list."
    exit 1
else
    log_info "Installing Mise..."
    curl -fsSL -o "${TEMP_DIR}/install-mise.sh" https://mise.run
    sh "${TEMP_DIR}/install-mise.sh"
    MISE_BIN="$HOME/.local/bin/mise"
fi
if [[ ! -x "$MISE_BIN" ]]; then
    log_error "Mise not found at ${MISE_BIN}."
    exit 1
fi
(cd "$HOME" && "$MISE_BIN" install)

if [[ "$PLATFORM" == macos ]]; then
    FONT_DIR="$HOME/Library/Fonts"
else
    FONT_DIR="$HOME/.local/share/fonts"
fi
for font in "${NERD_FONTS[@]}"; do
    if find "$FONT_DIR" -name "*${font}*" -print -quit 2>/dev/null | grep -q .; then
        log_success "${font} Nerd Font already installed."
    else
        mkdir -p "$FONT_DIR"
        log_info "Installing ${font} Nerd Font..."
        curl -fsSL -o "${TEMP_DIR}/${font}.tar.xz" \
            "https://github.com/ryanoasis/nerd-fonts/releases/latest/download/${font}.tar.xz"
        tar -xf "${TEMP_DIR}/${font}.tar.xz" -C "$FONT_DIR"
    fi
done
if [[ "$PLATFORM" != macos ]]; then
    fc-cache -f "$FONT_DIR"
fi

# ============================================================
# Phase 4: Plugin managers
# ============================================================
log_info "=== Phase 4: Plugin Managers ==="
if [[ ! -f "$HOME/.local/share/zap/zap.zsh" ]]; then
    log_info "Installing Zap-Zsh..."
    curl -fsSL -o "${TEMP_DIR}/install-zap.zsh" https://raw.githubusercontent.com/zap-zsh/zap/master/install.zsh
    zsh "${TEMP_DIR}/install-zap.zsh" --branch release-v1 --keep
fi
if [[ ! -d "$HOME/.tmux/plugins/tpm" ]]; then
    log_info "Installing Tmux Plugin Manager..."
    git clone https://github.com/tmux-plugins/tpm "$HOME/.tmux/plugins/tpm"
fi

# Changing the login shell is optional and only offered in an interactive session.
if [[ "$PLATFORM" != macos ]] && command -v getent >/dev/null 2>&1; then
    LOGIN_SHELL="$(getent passwd "$(id -un)" | cut -d: -f7 || true)"
    if [[ -n "$LOGIN_SHELL" && "${LOGIN_SHELL##*/}" != zsh && -t 0 ]]; then
        if read -r -p "Change your login shell to zsh? [y/N] " answer && [[ "$answer" == y || "$answer" == Y ]]; then
            chsh -s "$(command -v zsh)"
            log_success "Login shell changed; it takes effect on your next login."
        fi
    fi
fi

log_success "=== All Done! ==="
