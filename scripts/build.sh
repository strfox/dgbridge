#!/usr/bin/env bash
OPTIND=1
all=0
while getopts "a" opt; do
  case "$opt" in
    a)
      all=1
      ;;
    *)
      exit 1
      ;;
  esac
done
shift $((OPTIND-1))
[ "${1:-}" = "--" ] && shift

script_dir="$(dirname "$0")"
declare -a src_dirs=("$script_dir/../src/dgbridge" "$script_dir/../src/ruletester")

for src_dir in "${src_dirs[@]}"
do
  if [ $all -eq 1 ]; then
    gox -os="!windows" -osarch="!darwin/386" "$src_dir"
  else
    go build "$src_dir"
  fi
done