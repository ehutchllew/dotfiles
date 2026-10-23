#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "$0")" && pwd)"

if [[ $EUID -eq 0 ]]; then
  echo "Run this as your normal user, not root (makepkg refuses to run as root)." >&2
  exit 1
fi

# Strip blank lines and comments so the lists can be annotated
read_list() {
  grep -vE '^\s*(#|$)' "$1" || true
}

echo "== Installing repo packages =="
sudo pacman -Syu --needed - < <(read_list "$DIR/arch-pkglist.txt")

if ! command -v yay >/dev/null; then
  echo "== Bootstrapping yay =="
  sudo pacman -S --needed git base-devel
  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT
  git clone https://aur.archlinux.org/yay.git "$tmp/yay"
  (cd "$tmp/yay" && makepkg -si --noconfirm)
fi

if [[ -s "$DIR/arch-aur-pkglist.txt" ]]; then
  echo "== Installing AUR packages =="
  yay -S --needed - < <(read_list "$DIR/arch-aur-pkglist.txt")
fi
