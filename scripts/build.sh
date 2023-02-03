#!/usr/bin/env sh
OPTIND=1

script_dir="$(dirname "$0")"
src_dir="$script_dir/../src/dgbridge"

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

if [ $all -eq 1 ]; then
  gox -os="!windows" -osarch="!darwin/386" "$src_dir"
else
  go build "$src_dir"
fi
