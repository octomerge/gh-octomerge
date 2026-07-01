---
name: gh-octomerge-dev
description: >
  Develop, build, test, and release the gh-octomerge GitHub CLI extension. Use when
  working in this repository - adding commands or flags, wiring cobra/fang, building or
  installing the extension locally, cutting a tagged release, or changing the octomerge
  GitHub App install flow.
---

# Developing the gh-octomerge extension

`gh-octomerge` is a GitHub CLI extension (Go, precompiled) whose command `gh octomerge`
opens a Charm TUI form that helps a user install the **octomerge GitHub App**
(<https://github.com/apps/octomerge>) on one of their organizations.

## When to activate

- Adding or changing commands, flags, or the TUI form
- Building, installing, or debugging the extension locally
- Cutting a release
- Anything touching org discovery or the App install flow

## Architecture (command-first, domains over layers)

Flat and decoupled - no `internal/` or `pkg/` (see the `go` and `cobra-viper` skills):

    main.go            → cmd.Execute()
    cmd/root.go        → cobra root wrapped by fang; parses --org/--yes; calls install.Run
    install/           → the domain package (all behavior)
      install.go       → Options, InstallURL(), OpenApp(), Run() orchestration
      orgs.go          → ListUserOrgs() via go-gh REST (GET /user/orgs)
      form.go          → huh v2 TUI (org Select + manual Input + Confirm)
      install_test.go  → table-driven tests

Keep behavior in `install`, decoupled from cobra and from huh, so it stays testable. The
command layer only parses flags and delegates.

## Local dev loop

The host has no Go - the toolchain comes from Flox (see the `flox-env` skill):

    flox activate                      # Go 1.26.x on PATH
    go build ./... && go test ./...    # compile + unit tests
    go build -o gh-octomerge .         # produce the gh-<name> binary
    gh extension install --force .     # install this dir as `gh octomerge`
    gh octomerge                       # run the TUI
    gh octomerge --org my-org --yes    # non-interactive: opens the App page directly

`gh extension install .` names the command after the directory, so the folder MUST be
`gh-octomerge`.

## Naming rule (important)

A gh extension's command is its repo/binary name minus the `gh-` prefix. Never name this
`gh-extension` - that resolves to `gh extension`, which collides with gh's built-in
extension manager and is unreachable. Repo, folder, and binary are all `gh-octomerge`.

## Flags

- `--org, -o`  target organization; skips the picker
- `--yes, -y`  skip the confirmation prompt (with `--org`, fully non-interactive)

## Install URL

`gh octomerge` opens `https://github.com/apps/octomerge` - the public App landing page with
the Install button - defined in `install.InstallURL()`. Do NOT point at
`https://github.com/organizations/<org>/settings/apps/octomerge`; that is the app OWNER
settings page, reachable only by octomerge org admins. For a deep install flow,
`https://github.com/apps/octomerge/installations/new` is the one-line alternative.

## Releasing

Releases are precompiled by `.github/workflows/release.yml`
(`cli/gh-extension-precompile@v2`, `go_version_file: go.mod`). Push a tag:

    git tag v0.1.0 && git push origin v0.1.0

The action cross-compiles and attaches `gh-octomerge_<version>_<os>-<arch>` assets; users
then install with `gh extension install octomerge/gh-octomerge`. Inject the version via
`-ldflags -X github.com/octomerge/gh-octomerge/cmd.version=<tag>`.
