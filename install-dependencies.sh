#!/bin/bash

COLOR_OFF="\e[0m"
GREEN="\e[1;32m"
YELLOW="\e[1;33m"
CYAN="\e[1;36m"
#
# TODO
# Check if deps are not only present (-f), but executable (-x)
# if present and not executable, run a `chmod` to fix it

echo -e "${GREEN}Let's get crackin'!${COLOR_OFF}"

# Install Homebrew/Linuxbrew
# kernel_name=$(uname -s) # <-- not sure if we need custom config for Linux vs MacOS
# kernal_name == 'Darwin' || 'Linux'
if [[ ! -f "$(brew --prefix)/bin/brew" ]]; then
    echo -e "${CYAN}Installing Homebrew...${COLOR_OFF}"
    zsh $(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)
fi


# Install Brew Leaves
script_path="$(dirname $0)/brew-leaves.txt"
if [[ -e $script_path && -s $script_path ]]; then
  echo -e "${CYAN}Installing Brew Leaves...${COLOR_OFF}"
  xargs brew install < "brew-leaves.txt"
else
  echo "No brew leaves to install."
fi

# Install zap-zsh
if [[ ! -f "$HOME/.local/share/zap/zap.zsh" ]]; then
    echo -e "${CYAN}Installing Zap-Zsh...${COLOR_OFF}"
    zsh <(curl -s https://raw.githubusercontent.com/zap-zsh/zap/master/install.zsh) --branch release-v1
fi

# NeoVim
## Currently done through Brew above

# Stow
## Currently done through Brew above

# LunarVim
if [[ ! -f "$HOME/.local/bin/lvim" ]]; then
  echo -e "${CYAN}Installing LunarVim...${COLOR_OFF}"
  ## Might need to check and install lvim deps like cargo, node, make, etc.
  ## Also need to remove the default lvim user config file and use stowed version.
  LV_BRANCH='release-1.3/neovim-0.9' zsh <(curl -s https://raw.githubusercontent.com/LunarVim/LunarVim/release-1.3/neovim-0.9/utils/installer/install.sh)
fi

# Volta

if [[ ! -f "$HOME/.volta/bin/volta" ]]; then
  echo -e "${YELLOW}Volta is not installed${COLOR_OFF}"
  echo -e -n "${YELLOW}Would you like to install it? ${GREEN}(y/n):${COLOR_OFF} "   
  read -q answer 
  echo
  if [[ $answer == "Y" || $answer == "y" ]]; then
    echo -e "${CYAN}Installing Volta...${COLOR_OFF}"
    zsh <(curl https://get.volta.sh)
  else
    echo -e "${CYAN}Volta will not be installed.${COLOR_OFF}"
  fi
fi

echo -e "${GREEN}All Done!${COLOR_OFF}"
