# danger-go

A small Go-native experiment inspired by Danger.

`danger-go` reads `.danger.yaml` or `.danger.yml`, inspects the repository, and reports policy failures. The first version is intentionally config-driven so it can run as a single binary without Ruby.

This repository dogfoods `danger-go` in GitHub Actions.

PR comment posting is verified with a same-repository test pull request.

## Documentation

Read the full documentation at:

https://harryvince.github.io/danger-go/

Key pages:

- [Quickstart](https://harryvince.github.io/danger-go/quickstart/)
- [GitHub Action](https://harryvince.github.io/danger-go/github-action/)
- [Configuration](https://harryvince.github.io/danger-go/configuration/)
- [Examples](https://harryvince.github.io/danger-go/examples/)
- [Changelog and Releases](https://harryvince.github.io/danger-go/changelog/)
- [Roadmap](https://harryvince.github.io/danger-go/roadmap/)

Documentation source lives in [`docs/`](docs/).

## GitHub Action

Create `.github/workflows/danger-go.yml`:

```yaml
name: danger-go

on:
  pull_request:

permissions:
  contents: read
  pull-requests: write
  issues: write

jobs:
  danger-go:
    runs-on: ubuntu-latest
    steps:
      - uses: harryvince/danger-go@v1
```

Then add `.danger.yaml` or `.danger.yml` to the repository.

The action checks out the repository, installs Go, runs `danger-go ci`, reads pull request metadata, evaluates the configured rules, and creates or updates a pull request comment when permissions allow it.

### Action Inputs

| Input | Default | Description |
| --- | --- | --- |
| `config` | auto-detect | Path to `.danger.yaml` or `.danger.yml`. |
| `github-token` | automatic `${{ github.token }}` | Optional token used to read pull request metadata and post comments. |
| `go-version` | `stable` | Go version passed to `actions/setup-go`. |

### Token and Permissions

You do not need to create a personal access token for normal use. GitHub automatically provides `${{ github.token }}` to every workflow run, and the action uses it by default.

The workflow permissions decide what that automatic token is allowed to do. For pull request comments, set:

```yaml
permissions:
  contents: read
  pull-requests: write
  issues: write
```

`contents: read` lets the action check out and inspect the repository. `pull-requests: write` allows pull request operations. `issues: write` is needed because GitHub pull request comments use the Issues comments API.

Pass `github-token` only if you have a specific reason to use a different token, such as a GitHub App token or a repository policy that prevents the automatic token from doing what you need.

For pull requests from forks, GitHub may restrict `GITHUB_TOKEN` permissions.

## Local Usage

```sh
go run ./cmd/danger-go local
```

Print the CLI version:

```sh
go run ./cmd/danger-go version
```

Release binaries are attached to GitHub releases for Linux, macOS, and Windows on amd64 and arm64.

## Configuration

```yaml
rules:
  max_changed_files: 50
  require_pr_title_pattern: ".+"
  required_files:
    - go.mod
  forbidden_files:
    - "*.tmp"
```

Supported config filenames:

- `.danger.yaml`
- `.danger.yml`
