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
      install.go       → Options, InstallURL(org), OpenApp(), Run() orchestration
      orgs.go          → ListUserOrgs() + LookupOrgID() via go-gh REST (GET /user/orgs, /orgs/{org})
      form.go          → huh v2 TUI (org Select + manual Input + Confirm)
      install_test.go  → table-driven tests

Keep behavior in `install`, decoupled from cobra and from huh, so it stays testable. The
command layer only parses flags and delegates.

## Local dev loop

The host has no Go - the toolchain and the Task runner both come from Flox (see the
`flox-env` skill). Drive the loop through the Taskfile:

    flox activate                      # Go 1.26.x + task on PATH
    task install                       # rebuild the binary and (re)install as `gh octomerge`
    task test                          # unit tests
    task build                         # just compile ./gh-octomerge
    gh octomerge                       # run the TUI
    gh octomerge --org my-org --yes    # non-interactive: deep-links to the org's install page

gh manages a local extension as a symlink to this repo and does NOT recompile it, so a
precompiled extension only reflects source changes after `go build -o gh-octomerge .`.
`gh extension install --force .` only re-links and even *refuses* when the command is already
installed, which is why `task install` rebuilds then removes + re-adds the symlink.
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

`gh octomerge` deep-links straight to the chosen org's install page so the user never picks
the org twice. `InstallURL(org Org)` (in `install.go`) returns
`https://github.com/apps/octomerge/installations/new/permissions?suggested_target_id=<id>`
when the org's numeric ID is known - `suggested_target_id` pre-selects that account. The ID
comes from `GET /user/orgs` for picked orgs (free) or from `LookupOrgID` (`GET /orgs/{org}`)
for manual entry and the `--org` flag; if it can't be resolved, it falls back to
`https://github.com/apps/octomerge/installations/new`, where GitHub shows the account picker.
Do NOT point at `https://github.com/organizations/<org>/settings/apps/octomerge`; that is the
app OWNER settings page, reachable only by octomerge admins.

## Releasing

Releases are precompiled by `.github/workflows/release.yml`
(`cli/gh-extension-precompile@v2`, `go_version_file: go.mod`). Push a tag:

    git tag v0.1.0 && git push origin v0.1.0

The action cross-compiles and attaches `gh-octomerge_<version>_<os>-<arch>` assets; users
then install with `gh extension install octomerge/gh-octomerge`. Inject the version via
`-ldflags -X github.com/octomerge/gh-octomerge/cmd.version=<tag>`.
