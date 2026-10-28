# My Dotfiles

## Overview

Dotfile configurations to replicate development environment across different OS architectures and distro flavors.

Current Supported Platforms:
- Apple Silicon macOS
- Debian-based Linux
- Arch-based Linux

### Package Managers
- Mise: used for runtime management
    - Tools/software that can have varying versions within a system, i.e. per project.
    - ex: `go`, `node`, `zig`
- Pacman: system-wide management from Arch official repo
    - Tools/software that need just one stable version.
    - ex: `git`, `hyprland`, `neovim`, `stow`, `zsh`
- Yay: system-wide management from Arch user repo
    - Like Pacman but for packages not in the official repo.
    - ex: `zen-browser`
- Apt: system-wide management for Debian-based platforms
    - See: Pacman/Yay for description and examples.
- Homebrew/Linuxbrew: system-wide management for Mac & Debian-based Linux
   - See: Pacman/Yay for description and examples.

## Bootstrapping Dependencies

Run the bootstrap as your normal user from a checkout of this repository:

```sh
bash install-dependencies.sh
```

It detects macOS or the Linux distribution and the machine architecture, then
installs packages, stows the configs for that platform, installs the tools in
the mise config, and sets up the DepartureMono Nerd Font, Zap-Zsh, and TPM.

Package installation varies by platform:
- **macOS:** Homebrew installs the packages in `bootstrap/brew/Brewfile`.
- **Debian-based Linux:** apt installs system prerequisites and Alacritty, then
  Linuxbrew installs the packages in `bootstrap/brew/Brewfile`.
- **Arch-based Linux:** `bootstrap/arch/install-pkgs.sh` uses pacman and yay
  with `bootstrap/arch/pkglist.txt` and `bootstrap/arch/aur-pkglist.txt`.

## Updating Dependencies

To record newly installed Arch packages:
```sh
bash bootstrap/arch/gen-pkglist.sh
```

For Arch: Packages that come from the OS installation, such as the kernel and bootloader,
belong in `bootstrap/arch/ignore-pkglist.txt` or `bootstrap/arch/ignore-aur-pkglist.txt`. The lists accept
`#` comments, but `bootstrap/arch/gen-pkglist.sh` re-sorts a list when it adds packages, which
moves comment lines.

To update dependency versions:
- **All**:
    - Mise: `mise upgrade` or `mise upgrade [flags] [tool@version]`
- **Arch-based**:
    - Pacman: `sudo pacman -Syu`.
    - Yay: `yay -Syu` or `yay -S <aur-pkg>`.
- **Debian-based**:
    - Apt: `sudo apt update && sudo apt upgrade` or `sudo apt update && sudo apt install --only-upgrade <pkg>`.
    - Homebrew: `brew update && brew upgrade` or `brew update && brew upgrade [formula]`.
- **macOS**:
    - Homebrew: `brew update && brew upgrade` or `brew update && brew upgrade [formula]`.

## Using Stow

Call `stow {directory}` from this project's root directory
Example:
```sh
stow zsh zsh-macos
```

This will create a symlink between the previously stowed directory and contents in the Linux/Mac home dir (~/).

For instance, running `stow zsh` will create a symlink at `~/.zshrc` and `~/.config/zsh/`.

The bootstrap stows these packages on each platform:

| Platform | Stow packages |
|---|---|
| All | `mise`, `nvim`, `tmux`, `zsh` |
| macOS | All + `alacritty`, `zsh-macos` |
| Debian | All + `alacritty`, `zsh-debian` |
| Arch | All + `hypr`, `waybar`, `rofi`, `zsh-arch` |

## Using Mise

Mise is a way to install and manage different versions of software.

### Installing Tools from Config Files
To install all the tools listed within the stowed `config.toml` file you simply have to run:

```sh
mise install
```

This works as well when using idiomatic versioning files like `.nvmrc` or `.go-version`.

### Listing Installed, Active, and Remote Tools
To list all tools installed by mise run:

```sh
mise ls
```

To list tools available to install and their package versions run:

```sh
mise ls-remote {tool}[@{version}]
```

Example:

```
mise ls-remote node
mise ls-remote node@20
```

To view the latest version of a tool (and version) run:

```sh
mise latest {tool}[@{version}]
```

Example:

```sh
mise latest node
mise latest node@20
```

### Using and Pinning Tools
To use a specific version of a tool run:

```sh
mise use {tool}[@{version}]
```

Example:

```sh
mise use node@20
```

### Removing Tools
To remove a tool from mise run:

```sh
mise uninstall [--all] [--dry-run] {installed_tool}[@{version}]
```

Example:

```
mise uninstall node
mise uninstall node@20
```

## Using TMUX

Run a Tmux session:
```sh
tmux
```

(leader key by default is `<CTRL-B>`)
When inside of a tmux session press `<leader> + I`.

To view the various tmux commands related to sessions, windows, panes, etc.
press `<leader> + ?`

To exit out of a pane, window, or session simply type `exit` at the terminal command line.

## Custom Functions

### Recursively Remove

Invoke a recursive removal of a keyword using `rrm {keyword}`
Example:

```sh
rrm "node_modules"
```

After showing matching directories and a confirmation, this will remove all node_modules directories at the root of the function invocation and any sub-directories as well.

## TODO

- Verify `install-dependencies.sh` on macOS and Linux container/VM images.
- Update NeoVim `<leader>sf` keymap to include option to open file in new vertical window.
    - Do the same for all the `<leader>s*` commands.

### Author: Evan Hutch (ev-the-dev)
