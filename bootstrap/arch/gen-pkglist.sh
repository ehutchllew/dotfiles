#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "$0")" && pwd)"

# shellcheck source=lib.sh
source "$DIR/lib.sh"

# Usage: sync_list <list file> <ignore file> <pacman query flags>
# Note: adding packages re-sorts the list, which moves any comment lines.
sync_list() {
  local list="$DIR/$1" ignore="$DIR/$2" query="$3"

  # Installed explicitly, but in neither the list nor the ignore file
  local new
  new=$(comm -13 <({ read_list "$list"; read_list "$ignore"; } | sort -u) <(pacman "$query" | sort -u))
  if [[ -n "$new" ]]; then
    echo "$new" >> "$list"
    sort -u -o "$list" "$list"
    echo "Added to $1:"
    echo "$new" | sed 's/^/  + /'
  fi

  # In the list but no longer installed (reported, not removed)
  local stale
  stale=$(comm -23 <(read_list "$list" | sort -u) <(pacman "$query" | sort -u))
  if [[ -n "$stale" ]]; then
    echo "In $1 but not installed:"
    echo "$stale" | sed 's/^/  - /'
  fi
}

sync_list pkglist.txt     ignore-pkglist.txt     -Qqen
sync_list aur-pkglist.txt ignore-aur-pkglist.txt -Qqem
