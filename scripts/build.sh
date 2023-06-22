#!/usr/bin/env bash
set -e

script_dir="$(dirname "$0")"
version=$(cat "$script_dir/../VERSION")
declare -a src_dirs=("$script_dir/../src/dgbridge" "$script_dir/../src/ruletester")
linker_flags="-X dgbridge/src/lib.Version=$version"

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
shift $((OPTIND - 1))
[ "${1:-}" = "--" ] && shift

for src_dir in "${src_dirs[@]}"; do
  project_name=$(basename "$src_dir")
  if [ $all -eq 1 ]; then
    gox \
      -ldflags "$linker_flags" \
      -os="!windows !plan9" \
      -osarch="!darwin/386" \
      -output="{{.Dir}}_{{.OS}}_{{.Arch}}-$version" \
      "$src_dir"
  else
    go build \
      -ldflags "$linker_flags" \
      -o "$project_name-$version" \
      "$src_dir"
  fi
done
