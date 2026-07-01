# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`gh-octomerge` is a precompiled **GitHub CLI extension** (Go). The command `gh octomerge`
renders a Charm TUI that walks a user through installing the **octomerge GitHub App**
(<https://github.com/apps/octomerge>) on one of their organizations, then opens the App's
install page in the browser.

## Development

**The host has no Go toolchain — it comes from Flox.** Almost every "go: command not found"
is a missing `flox activate`. Run everything inside the env:

```sh
flox activate                      # Go 1.26 + gopls + goimports on PATH
go build ./... && go test ./...    # compile + unit tests
go vet ./...                       # vet check (no dedicated linter is configured)
go build -o gh-octomerge .         # build the gh-<name> binary
gh extension install --force .     # install this checkout as `gh octomerge`
gh octomerge                       # interactive run
gh octomerge --org my-org --yes    # non-interactive: opens the App page directly
```

Run a single test with the standard Go filter, e.g. `go test ./install/ -run TestResolveOrg`.
To run one command in the env without a subshell: `flox activate -- go test ./...`.

`GOTOOLCHAIN=local` and `GOPATH` inside `.flox/cache` are set by the manifest's `on-activate`
hook, so builds are self-contained and never download another toolchain.

## Architecture

Command-first, **domains over layers** — flat, no `internal/` or `pkg/`:

```
main.go        → cmd.Execute()
cmd/root.go    → cobra root wrapped by fang; parses --org/--yes; delegates to install.Run
install/       → the domain package — ALL behavior lives here
  install.go   → Options, InstallURL(), OpenApp(), Run() orchestration
  orgs.go      → ListUserOrgs() via go-gh REST (GET /user/orgs)
  form.go      → huh v2 TUI (org Select + manual-entry Input + Confirm)
```

The load-bearing rule: keep behavior in `install`, **decoupled from cobra and from huh**, so
it stays unit-testable. The command layer only parses flags and delegates. The TUI form
(`form.go`) is separable from orchestration (`install.go`), which is separable from GitHub
access (`orgs.go`). Preserve these seams when adding features.

## Non-obvious constraints

- **Extension naming.** A gh extension's command is its repo/binary name minus the `gh-`
  prefix, and `gh extension install .` names the command after the directory. Repo, folder,
  and binary MUST all be `gh-octomerge`. Never rename to `gh-extension` — it resolves to
  `gh extension` (gh's built-in manager) and becomes unreachable.
- **Install URL.** `install.InstallURL()` returns `https://github.com/apps/octomerge`, the
  public landing page with the Install button. Do NOT point at
  `https://github.com/organizations/<org>/settings/apps/octomerge` (the App *owner* settings
  page, reachable only by octomerge's own admins). A deep-link flow would use
  `.../apps/octomerge/installations/new`.
- **No token handling.** `orgs.go` uses go-gh's `api.DefaultRESTClient()`, which reuses gh's
  existing credentials and host. Don't add token plumbing.
- **`GET /user/orgs` only returns orgs with *visible* membership.** The form's manual-entry
  fallback exists precisely to cover private memberships and API failures — keep it.
- **huh v2 conditional visibility is per-group**, not per-field. The manual-entry Input is its
  own `huh.NewGroup(...).WithHideFunc(...)`; the Select and Confirm are separate groups.
- **`manualSentinel` is `"\x00manual"`** — the NUL byte guarantees it never collides with a
  real org login used as a Select value.
- A Ctrl-C / Esc abort is returned as an *unconfirmed result*, not an error, so the caller
  exits cleanly (see `runForm`'s `huh.ErrUserAborted` handling).
- `cmd/root.go` sets `SilenceUsage`/`SilenceErrors` because **fang** renders errors itself and
  also provides `--version`, styled help, and signal-aware cancellation.
- Charm dependencies use the **`charm.land/...` vanity import paths** (huh/v2, bubbletea/v2,
  lipgloss/v2), not `github.com/charmbracelet/...`.

## Releasing

Push a semver tag; `.github/workflows/release.yml` runs `cli/gh-extension-precompile`
(`go_version_file: go.mod`) to cross-compile binaries and attach them to the release. The
release workflow does **not** use Flox — precompile reads the Go version from `go.mod`.

```sh
git tag v0.1.0 && git push origin v0.1.0
```

The version string is injected at build time via
`-ldflags -X github.com/octomerge/gh-octomerge/cmd.version=<tag>` (defaults to `dev`).

## Skills

Repo guidance lives in `.agents/skills/` (symlinked as `.claude/`). Repo-specific:
`gh-octomerge-dev` (this extension), `flox-env` (the toolchain), `charm-tui` (huh v2 / Bubble
Tea patterns as used in `form.go`). Vendored from spf13/go-skills (general Go/CLI reference):
`go`, `cobra-viper`, `go-spec-reviewer`.
