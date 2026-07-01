#!/usr/bin/env bash
# Cross-compile gh-octomerge for every platform gh can install, baking in the
# release version. Invoked by semantic-release (@semantic-release/exec):
#   bash script/build.sh <tag>     e.g. bash script/build.sh v0.1.0
# File names match cli/gh-extension-precompile (gh matches the "<os>-<arch>"
# suffix), so `gh extension install/upgrade` keeps working.
set -euo pipefail

tag="${1:?usage: build.sh <version-tag>}"
ldflags="-s -w -X github.com/octomerge/gh-octomerge/cmd.version=${tag}"

platforms=(
  darwin-amd64 darwin-arm64
  freebsd-386 freebsd-amd64 freebsd-arm64
  linux-386 linux-amd64 linux-arm linux-arm64
  windows-386 windows-amd64 windows-arm64
)

rm -rf dist
mkdir -p dist
for p in "${platforms[@]}"; do
  goos="${p%%-*}"
  goarch="${p#*-}"
  ext=""
  [ "$goos" = "windows" ] && ext=".exe"
  echo "building dist/${p}${ext}"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
    go build -trimpath -ldflags "$ldflags" -o "dist/${p}${ext}" .
done
