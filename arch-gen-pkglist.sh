#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "$0")" && pwd)"

# Usage: sync_list <list file> <ignore file> <pacman query flags>
sync_list() {
  local list="$DIR/$1" ignore="$DIR/$2" query="$3"

  # Installed explicitly, but in neither the list nor the ignore file
  local new
  new=$(comm -13 <(sort -u "$list" "$ignore") <(pacman "$query" | sort))
  if [[ -n "$new" ]]; then
    echo "$new" >> "$list"
    sort -u -o "$list" "$list"
    echo "Added to $1:"
    echo "$new" | sed 's/^/  + /'
  fi

  # In the list but no longer installed (reported, not removed)
  local stale
  stale=$(comm -23 <(sort -u "$list") <(pacman "$query" | sort))
  if [[ -n "$stale" ]]; then
    echo "In $1 but not installed:"
    echo "$stale" | sed 's/^/  - /'
  fi
}

sync_list arch-pkglist.txt     arch-pkglist-ignore.txt     -Qqen
sync_list arch-aur-pkglist.txt arch-aur-pkglist-ignore.txt -Qqem
