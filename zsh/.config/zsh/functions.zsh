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
