#!/bin/bash
#
# TODO
# Check if deps are not only present (-f), but executable (-x)
# if present and not executable, run a `chmod` to fix it

# Install Homebrew/Linuxbrew
# kernel_name=$(uname -s) # <-- not sure if we need custom config for Linux vs MacOS
# kernal_name == 'Darwin' || 'Linux'
if [[ ! -f "$(brew --prefix)/bin/brew" ]]; then
    echo "Installing Homebrew..."
    zsh $(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)
fi


# Install Brew Leaves
script_path="$(dirname $0)/brew-leaves.txt"
if [[ -e $script_path && -s $script_path ]]; then
  echo "Installing Brew Leaves..."
  xargs brew install < "brew-leaves.txt"
else
  echo "No brew leaves to install."
fi

# Install zap-zsh
if [[ ! -f "$HOME/.local/share/zap/zap.zsh" ]]; then
    echo "Installing Zap-Zsh..."
    zsh <(curl -s https://raw.githubusercontent.com/zap-zsh/zap/master/install.zsh) --branch release-v1
fi

# NeoVim
## Currently done through Brew above

# Stow
## Currently done through Brew above

# LunarVim
if [[ ! -f "$HOME/.local/bin/lvim" ]]; then
  echo "Installing LunarVim..."
  ## Might need to check and install lvim deps like cargo, node, make, etc.
  ## Also need to remove the default lvim user config file and use stowed version.
  LV_BRANCH='release-1.3/neovim-0.9' zsh <(curl -s https://raw.githubusercontent.com/LunarVim/LunarVim/release-1.3/neovim-0.9/utils/installer/install.sh)
fi

# Volta

if [[ ! -f "$HOME/.volta/bin/volta" ]]; then
  echo "Volta is not installed"
  read -q "?Would you like to install it? (y/n): " answer
  if [[ $answer == "Y" || $answer == "y" ]]; then
    echo "Installing Volta..."
    zsh <(curl https://get.volta.sh)
  else
    echo "Volta will not be installed."
  fi
fi
