# My Dotfiles


## Bootstrapping Dependencies
Invoke the `install-dependencies.sh` script to ensure all workstation deps are present.

This will ensure programs like the following are installed:
 * Homebrew/Linuxbrew
 * NeoVim
 * LunarVim
 * Zap-Zsh
 * Stow
 * Volta

Furthermore, the script will proceed to `brew install` a list of casks/formulae as well.


## Using Stow
Call `stow {directory}` from this project's root directory
Example:
```sh
stow zsh
```
This will create a symlink between the previously stowed directory and contents in the Linux/Mac home dir (~/).

For instance, running `stow zsh` will create a symlink at `~/.zshrc` and `~/.config/zsh/`.


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
