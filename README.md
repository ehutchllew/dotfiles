# My Dotfiles

## Bootstrapping Dependencies

Invoke the `install-dependencies.sh` script to ensure all workstation deps are present.

This will ensure programs like the following are installed:

- Homebrew/Linuxbrew
- NeoVim
- LunarVim
- Zap-Zsh
- Stow
- Volta

<!-- TODO: Will be removing the `brew-leaves` functionality in favor of using `mise` -->

Furthermore, the script will proceed to `brew install` a list of casks/formulae as well.

## Using Stow

Call `stow {directory}` from this project's root directory
Example:

```sh
stow zsh
```

This will create a symlink between the previously stowed directory and contents in the Linux/Mac home dir (~/).

For instance, running `stow zsh` will create a symlink at `~/.zshrc` and `~/.config/zsh/`.

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

To ensure all tmux plugins are installed run:

```sh
tmux
```

(leader key by default is `<CTRL-B>`)
When inside of a tmux session press `<leader> + I`.

To view the various tmux commands related to sessions, windows, panes, etc.
press `<leader> + ?`

To exit out of a pane, window, or session simply type `exit` at the terminal command line.

## Updating Dependencies (TODO)

Invoke the `update-dependencies.sh` script manually to check for updates to all the workstation deps.

A goal in the future is to provide some Git Hooks that will run this when new code is pulled down from the
remote branch.

## Custom Functions

### Recursively Remove

Invoke a recursive removal of a keyword using `rrm {keyword}`
Example:

```sh
rrm "node_modules"
```

This will remove all node_modules directories at the root of the function invocation and any sub-directories as well.

### Author: Evan Hutch (ev-the-dev)
