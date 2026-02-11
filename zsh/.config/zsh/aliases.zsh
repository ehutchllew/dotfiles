# NVIM
alias n="nvim ."

# FZF
alias fzfp="fzf --preview 'bat --style=numbers --color=always {}'"

# Git
alias gc="git commit -S -m"
alias ga="git add ."
alias gfp="git fetch && git pull"
alias gph="git push -u origin HEAD"
alias gg="lazygit"

# List
alias l="eza -lahF --icons --git-ignore"
alias lt="eza -lahF --icons --git-ignore --tree --level=2"
alias ll="ls -alFG --color"
alias la="ls -a"

# File or Dir Size
# Calc bytes recursively for argument
alias dub="du -b"
# Calc total bytes for argument only
alias dus="du -sb"
# Calc human readable bytes
alias duh="du -sbh"

# File Manipulation
alias cb="xclip -selection 'clipboard' -i"
