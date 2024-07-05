# File Manipulation
## Recursively remove directories named '$1'
rrm(){
  local arg=$1
  if [[ -n $arg ]]
  then
    echo "\n:::::Directories to be removed:::::\n"
    find . -name $arg -type d -prune
    read -q "?Proceed? (y/n): " answer
    if [[ $answer == "Y" || $answer == "y" ]]
    then
      echo "\nProceeding...\n"
      find . -name $arg -type d -prune -exec rm -rf '{}' +
    else
      echo "\nAborting...\n"
    fi
  else
    echo "Please pass just one parameter: i.e. 'node_modules'"
  fi

  echo "All done!"
  return 0
}

# File Searching
## Find in File
fif() {
  if [ ! "$#" -gt 0 ]; then echo "Need a string to search for!"; return 1; fi
  rg --files-with-matches --hidden --ignore-case --no-messages "$1" | fzf --preview "highlight -O ansi -l {} 2> /dev/null | rg --colors 'match:bg:yellow' --ignore-case --pretty --context 10 '$1' || rg --ignore-case --pretty --context 10 '$1' {}"
}

# Git
## FZF Completion for `git checkout` -- triggered on `git co **<TAB>`
_fzf_complete_git() {
    ARGS="$@"
    local branches
    branches=$(git branch -vv --all)
    if [[ $ARGS == 'git co'* ]]; then
        _fzf_complete --reverse --multi -- "$@" < <(
            echo $branches
        )
    else
        eval "zle ${fzf_default_completion:-expand-or-complete}"
    fi
}

_fzf_complete_git_post() {
    awk '{print $1}'
}
