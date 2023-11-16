# My Dotfiles

## Using Stow
Call `stow {directory}` from this project's root directory
Example:
```sh
stow zsh
```
This will create a symlink between the previously stowed directory and contents in the Linux/Mac home dir (~/).
For instance, running `stow zsh` will create a symlink at `~/.zshrc` and `~/.config/zsh/`.
