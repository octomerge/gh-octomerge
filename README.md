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
3. After you confirm, it opens <https://github.com/apps/octomerge> in your browser, where
   you click **Install** and choose the organization.

> Only an organization **owner** can complete a GitHub App installation.

## Development

The Go toolchain is provided by [Flox](https://flox.dev) - you don't need Go installed on
your host.

```sh
flox activate                   # Go + gopls + goimports on PATH
go build ./... && go test ./... # build + unit tests
go build -o gh-octomerge .      # build the binary
gh extension install --force .  # install this checkout as `gh octomerge`
gh octomerge                    # try it
```

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

[Apache-2.0](./LICENSE). The vendored skills (`go`, `cobra-viper`, `go-spec-reviewer`)
are MIT, © [spf13/go-skills](https://github.com/spf13/go-skills).
