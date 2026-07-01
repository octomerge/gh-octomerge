---
name: flox-env
description: >
  Manage the Flox development environment for this repo, which provides the Go toolchain
  (the host has no Go). Use when activating the env, adding or pinning packages in
  .flox/env/manifest.toml, fixing "go: command not found", or reproducing the build.
---

# Flox dev environment

This repo gets its Go toolchain from Flox, not the host. The environment is defined in
`.flox/env/manifest.toml` (committed) and locked in `.flox/env/manifest.lock`.

## When to activate

- "go: command not found", or missing gopls/goimports
- Activating or reproducing the dev environment
- Adding, pinning, or upgrading a dev tool
- Editing `.flox/env/manifest.toml`

## Activate first

    flox activate                    # subshell with go/gopls/goimports on PATH
    flox activate -- go build ./...  # run one command in the env, non-interactively

Almost every "missing Go" problem is just a missing `flox activate`.

## manifest.toml anatomy

- `[install]` - packages (`go`, `gopls`, `gotools`, `go-task`). Add with `flox install <pkg>`,
  search with `flox search <pkg>`, pin with e.g. `go.version = "^1.26"`. `go-task` provides the
  `task` runner used by the repo's `Taskfile.yml` (e.g. `task install`).
- `[hook] on-activate` - runs on activate. Here it sets `GOPATH="$FLOX_ENV_CACHE/go"` and
  adds `$GOPATH/bin` to PATH, keeping the Go module cache inside the (git-ignored)
  `.flox/cache` for a self-contained, reproducible env.
- `[options] systems` - the platforms the env resolves for (aarch64/x86_64 ×
  darwin/linux). A contributor's system must be listed.

## Common tasks

    flox install <pkg>            # add a tool (updates [install] + lock)
    flox list                     # show installed packages
    flox activate -- go version   # verify the toolchain

## Reproducibility notes

- Commit `manifest.toml` and `manifest.lock`; `.flox/{run,cache,lib,log}` are git-ignored.
- Builds use `GOTOOLCHAIN=local` so Go uses the Flox-provided toolchain instead of
  downloading another.
- The release workflow does NOT use Flox - `cli/gh-extension-precompile` reads the Go
  version from `go.mod` directly.

## Troubleshooting

- "go: command not found" → you are not inside `flox activate`.
- "environment is not compatible with <system>" → add that system to `[options] systems`
  and re-lock (`flox activate`).
