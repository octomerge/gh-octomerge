# gh-octomerge

A [GitHub CLI](https://cli.github.com) extension that walks you through installing the
**[octomerge GitHub App](https://github.com/apps/octomerge)** ("Your GitHub merging
assistant") on one of your organizations - from a small terminal form.

Built with [Charm](https://charm.sh) [huh](https://github.com/charmbracelet/huh) (Bubble
Tea) for the TUI, [Cobra](https://github.com/spf13/cobra) +
[fang](https://github.com/charmbracelet/fang) for the command, and
[go-gh](https://github.com/cli/go-gh) for GitHub access.

## Install

```sh
gh extension install octomerge/gh-octomerge
```

Upgrade later with `gh extension upgrade octomerge`. Requires `gh` ≥ 2.0.

## Usage

```sh
gh octomerge                    # interactive: pick an org, confirm, open the App page
gh octomerge --org my-org --yes # non-interactive: open the install page directly
```

| Flag | Description |
| --- | --- |
| `-o, --org` | Target GitHub organization (skips the picker) |
| `-y, --yes` | Skip the confirmation prompt |
| `-v, --version` | Print the version |

## What it does

1. Lists the organizations you belong to (`GET /user/orgs`, using your existing `gh` auth).
2. You pick one - or choose **manual entry** to type any org.
3. After you confirm, it opens the install page **for the org you picked** - pre-selected via
   `suggested_target_id`, so you don't choose it again - where you review the permissions and
   click **Install**.

> Only an organization **owner** can finish the install; for anyone else GitHub turns it into
> a request to the org's owners. (If the org's numeric ID can't be resolved, it falls back to
> the standard install page with the account picker.)

## Development

The Go toolchain and the [Task](https://taskfile.dev) runner are provided by
[Flox](https://flox.dev) - you don't need Go installed on your host.

```sh
flox activate   # Go, gopls, goimports, and task on PATH
task install    # rebuild the binary and (re)install as `gh octomerge`
task test       # run the unit tests
task --list     # show all tasks
```

Note: gh manages a local extension as a **symlink** to this repo and does not recompile it, so
source changes only take effect once the binary is rebuilt - which is what `task install` does.
(`gh extension install --force .` only re-links, and refuses when already installed.)

See the `flox-env` and `gh-octomerge-dev` skills under `.agents/skills/` for details.

## Architecture

Command-first and flat (domains over layers - no `internal/`):

```
main.go        → cmd.Execute()
cmd/root.go    → Cobra root wrapped by fang; parses flags; calls install.Run
install/       → domain logic: org discovery, the huh form, opening the App page
```

## Releasing

Push a semver tag; [`.github/workflows/release.yml`](./.github/workflows/release.yml) uses
[`cli/gh-extension-precompile`](https://github.com/cli/gh-extension-precompile) to build
cross-platform binaries and attach them to the release:

```sh
git tag v0.1.0 && git push origin v0.1.0
```

## License

[Apache-2.0](./LICENSE).
