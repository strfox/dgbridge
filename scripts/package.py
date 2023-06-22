#!/usr/bin/env python3
import argparse
import os
import subprocess
from pathlib import Path


def main():
    parser = argparse.ArgumentParser(prog="Dgbridge Packager")
    parser.add_argument("--build_dir", required=True)
    parser.add_argument("--project_root", required=True)
    args = parser.parse_args()

    build_dir = Path(args.build_dir)
    project_root = Path(args.project_root)

    require_dir(build_dir)
    require_dir(project_root)

    for file in build_dir.glob("*"):
        if not file.is_file():
            continue
        subprocess.run(
            ["zip",
             "-r",
             f"{os.path.basename(file)}.zip",
             file,
             os.path.join(project_root, "tests"),
             os.path.join(project_root, "rules")])


def require_dir(path):
    if not os.path.isdir(path):
        print("no such directory: " + str(path))
        exit(1)


if __name__ == "__main__":
    main()
