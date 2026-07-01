# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`gh-octomerge` is a precompiled **GitHub CLI extension** (Go). The command `gh octomerge`
renders a Charm TUI that walks a user through installing the **octomerge GitHub App**
(<https://github.com/apps/octomerge>) on one of their organizations and opens the App's
install page in the browser - **unless the App is already installed on that org, in which
case it skips the install page**. The flow then continues in the terminal: it generates the
org's **`.octomerge`** configuration repository from the `octomerge/octomerge` template
(private by default; the user can choose public), or reports that it already exists.

## Development

**The host has no Go toolchain - it comes from Flox.** Almost every "go: command not found"
is a missing `flox activate`. Run everything inside the env:

```sh
flox activate                      # Go 1.26 + gopls + goimports on PATH
go build ./... && go test ./...    # compile + unit tests
go vet ./...                       # vet check (no dedicated linter is configured)
go build -o gh-octomerge .         # build the gh-<name> binary
gh extension install --force .     # install this checkout as `gh octomerge`
gh octomerge                       # interactive run
gh octomerge --org my-org --yes    # non-interactive: opens the App page + creates .octomerge (private)
gh octomerge --org my-org --yes --public   # ...and make the .octomerge repo public
```

Run a single test with the standard Go filter, e.g. `go test ./install/ -run TestResolveOrg`.
To run one command in the env without a subshell: `flox activate -- go test ./...`.

`GOTOOLCHAIN=local` and `GOPATH` inside `.flox/cache` are set by the manifest's `on-activate`
hook, so builds are self-contained and never download another toolchain.

## Architecture

Command-first, **domains over layers** - flat, no `internal/` or `pkg/`:

```
main.go        → cmd.Execute()
cmd/root.go    → cobra root wrapped by fang; parses --org/--yes/--public; delegates to install.Run
install/       → the domain package - ALL behavior lives here
  install.go   → Options, InstallURL(), OpenApp(), Run() + setupConfigRepo() orchestration
  orgs.go      → ListUserOrgs()/LookupOrgID()/AppInstalled() via go-gh REST
                 (GET /user/orgs, /orgs/{org}, /orgs/{org}/installations)
  repo.go      → RepoExists()/CreateFromTemplate() via go-gh REST (POST .../generate)
  form.go      → huh v2 TUI: selectOrg (org Select + manual Input), confirmInstall
                 (open-browser Confirm, shown only when not installed), and
                 runConfigForm (config confirm + visibility Select)
```

The load-bearing rule: keep behavior in `install`, **decoupled from cobra and from huh**, so
it stays unit-testable. The command layer only parses flags and delegates. The TUI forms
(`form.go`) are separable from orchestration (`install.go`), which is separable from GitHub
access (`orgs.go`, `repo.go`). Preserve these seams when adding features.

## Non-obvious constraints

- **Extension naming.** A gh extension's command is its repo/binary name minus the `gh-`
  prefix, and `gh extension install .` names the command after the directory. Repo, folder,
  and binary MUST all be `gh-octomerge`. Never rename to `gh-extension` - it resolves to
  `gh extension` (gh's built-in manager) and becomes unreachable.
- **Install URL.** `install.InstallURL()` returns `https://github.com/apps/octomerge`, the
  public landing page with the Install button. Do NOT point at
  `https://github.com/organizations/<org>/settings/apps/octomerge` (the App *owner* settings
  page, reachable only by octomerge's own admins). A deep-link flow would use
  `.../apps/octomerge/installations/new`.
- **No token handling.** `orgs.go` and `repo.go` use go-gh's `api.DefaultRESTClient()`, which
  reuses gh's existing credentials and host. Don't add token plumbing.
- **`GET /user/orgs` only returns orgs with *visible* membership.** The form's manual-entry
  fallback exists precisely to cover private memberships and API failures - keep it.
- **Install detection is best-effort.** `AppInstalled` reads `GET /orgs/{org}/installations`
  (matching `app_slug == "octomerge"`), which needs an org owner with the `admin:org` scope.
  On success it lets `Run` skip the browser step (and its `confirmInstall` prompt); on *any*
  error `Run` warns and falls back to showing the install step, so users without the scope
  see the pre-existing behavior. This is why the open-browser Confirm is its own step
  (`confirmInstall`) rather than bundled into `selectOrg` - the org must be known before the
  check, and the check decides whether that prompt is shown at all.
- **Config repo.** `repo.go` generates `<org>/.octomerge` from the `octomerge/octomerge`
  **template** via `POST /repos/octomerge/octomerge/generate` (repo name is literally
  `.octomerge`, like `.github`). Existence is pre-checked with a 404-tolerant
  `GET /repos/{owner}/.octomerge` (`RepoExists` treats `*api.HTTPError` `StatusCode == 404`
  as "doesn't exist"); if it exists the flow prints `The repository .octomerge already exists`
  and stops. `--yes` creates it unattended (private unless `--public`).
- **huh v2 conditional visibility is per-group**, not per-field. The manual-entry Input is its
  own `huh.NewGroup(...).WithHideFunc(...)`; the Select and Confirm are separate groups.
- **`manualSentinel` is `"\x00manual"`** - the NUL byte guarantees it never collides with a
  real org login used as a Select value.
- A Ctrl-C / Esc abort is returned as an *unconfirmed result*, not an error, so the caller
  exits cleanly (see the `huh.ErrUserAborted` handling in `selectOrg`, `confirmInstall`, and
  `runConfigForm`).
- `cmd/root.go` sets `SilenceUsage`/`SilenceErrors` because **fang** renders errors itself and
  also provides `--version`, styled help, and signal-aware cancellation.
- Charm dependencies use the **`charm.land/...` vanity import paths** (huh/v2, bubbletea/v2,
  lipgloss/v2), not `github.com/charmbracelet/...`.

## Releasing

Releases are automated by `.github/workflows/release.yml` (**`workflow_dispatch`**) using
**semantic-release v25** (config in `.releaserc.yaml`, YAML). It reads Conventional Commits on
`main`, computes the next `v<semver>` tag, writes/commits `CHANGELOG.md` back to `main`, and
publishes a GitHub Release. Binaries are cross-compiled by `script/build.sh` (invoked via
`@semantic-release/exec`) and uploaded by `@semantic-release/github`;
`actions/attest-build-provenance` attests them. The workflow uses Node (`actions/setup-node`)
+ Go (`actions/setup-go`, `go.mod`), not Flox.

The version string is injected for real here: `script/build.sh` builds with
`-ldflags -X github.com/octomerge/gh-octomerge/cmd.version=<tag>` (defaults to `dev` for local
builds). Asset names are `<os>-<arch>[.exe]`, which `gh` matches by suffix on install/upgrade.

## Skills

Repo guidance lives in `.agents/skills/` (symlinked as `.claude/`). Repo-specific:
`gh-octomerge-dev` (this extension), `flox-env` (the toolchain), `charm-tui` (huh v2 / Bubble
Tea patterns as used in `form.go`). Vendored from spf13/go-skills (general Go/CLI reference):
`go`, `cobra-viper`, `go-spec-reviewer`.
