# Shared helpers for the Arch package scripts. Source, don't execute.

# Print package names from a list file, ignoring comments (including inline
# ones), blank lines, and surrounding whitespace.
read_list() {
  awk '{ sub(/#.*/, ""); sub(/^[[:space:]]+/, ""); sub(/[[:space:]]+$/, "") } $0 != "" { print }' "$1"
}
