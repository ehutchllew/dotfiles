#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
ARCH="$(uname -m)"
REPO_LIST="$DIR/pkglist.txt"
AUR_LIST="$DIR/aur-pkglist.txt"
# Needed by the dotfiles bootstrap and to build yay, whatever the lists contain.
BOOTSTRAP_PKGS=(base-devel curl fontconfig git xz)

tmp=""
# Set by install_repo_packages, reported by report_unavailable.
UNAVAILABLE=""

# shellcheck source=lib.sh
source "$DIR/lib.sh"

cleanup() {
  if [[ -n "$tmp" && -d "$tmp" ]]; then
    rm -rf "$tmp"
  fi
}
trap cleanup EXIT

warn() { echo "WARNING: $*" >&2; }

require_normal_user() {
  if [[ $EUID -eq 0 ]]; then
    echo "Run this as your normal user, not root (makepkg refuses to run as root)." >&2
    exit 1
  fi
}

check_lists() {
  local list
  for list in "$REPO_LIST" "$AUR_LIST"; do
    if [[ ! -r "$list" ]]; then
      echo "Cannot read package list: $list" >&2
      exit 1
    fi
  done
}

# A full upgrade before installing anything avoids partial upgrades, and leaves
# fresh package databases for the availability check.
upgrade_system() {
  echo "== Upgrading system =="
  sudo pacman -Syu
}

# pacman rejects the whole transaction if any target is missing, so install
# only what the enabled repositories provide and record the rest.
install_repo_packages() {
  local requested available installable
  requested=$(
    {
      read_list "$REPO_LIST"
      printf '%s\n' "${BOOTSTRAP_PKGS[@]}"
    } | sort -u
  )
  available=$(pacman -Slq | sort -u)
  installable=$(comm -12 <(echo "$requested") <(echo "$available"))
  UNAVAILABLE=$(comm -23 <(echo "$requested") <(echo "$available"))

  if [[ -z "$installable" ]]; then
    warn "No packages from $REPO_LIST are available to install."
    return
  fi

  echo "== Installing repo packages =="
  echo "$installable" | sudo pacman -S --needed -
}

ensure_yay() {
  if command -v yay >/dev/null; then
    return
  fi

  echo "== Bootstrapping yay =="
  tmp=$(mktemp -d)
  git clone https://aur.archlinux.org/yay.git "$tmp/yay"
  (cd "$tmp/yay" && makepkg -si)
}

install_aur_packages() {
  local packages=() package
  while IFS= read -r package; do
    packages+=("$package")
  done < <(read_list "$AUR_LIST")

  if [[ ${#packages[@]} -eq 0 ]]; then
    return
  fi

  if [[ "$ARCH" != x86_64 ]]; then
    warn "Some AUR packages only support x86_64. If a build fails, remove it from $AUR_LIST and re-run."
  fi
  echo "== Installing AUR packages =="
  yay -S --needed "${packages[@]}"
}

report_unavailable() {
  if [[ -z "$UNAVAILABLE" ]]; then
    return
  fi

  echo >&2
  warn "These packages were skipped because no enabled repository provides them:"
  echo "$UNAVAILABLE" | sed 's/^/  - /' >&2
  echo >&2

  if [[ "$ARCH" == x86_64 ]] && ! pacman-conf --repo-list | grep -qx multilib; then
    echo "The multilib repository is disabled; packages such as steam require it." >&2
    echo "To enable it, uncomment these lines in /etc/pacman.conf:" >&2
    echo >&2
    echo "  [multilib]" >&2
    echo "  Include = /etc/pacman.d/mirrorlist" >&2
    echo >&2
    echo "Then re-run: bash $DIR/install-pkgs.sh" >&2
  elif [[ "$ARCH" == x86_64 ]]; then
    echo "They may have been renamed or removed from the repositories. Update $REPO_LIST accordingly." >&2
  else
    echo "They may not be built for $ARCH, or may have been renamed or removed." >&2
    echo "Remove them from $REPO_LIST if they are not needed here." >&2
  fi
}

main() {
  require_normal_user
  check_lists
  upgrade_system
  install_repo_packages
  ensure_yay
  install_aur_packages
  report_unavailable
}

main "$@"
