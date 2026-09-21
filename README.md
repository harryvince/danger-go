# danger-go

A small Go-native experiment inspired by Danger.

`danger-go` reads `.danger.yaml` or `.danger.yml`, inspects the repository, and reports policy failures. The first version is intentionally config-driven so it can run as a single binary without Ruby.

This repository dogfoods `danger-go` in GitHub Actions.

PR comment posting is verified with a same-repository test pull request.

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
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

Then add `.danger.yaml` or `.danger.yml` to the repository.

The action checks out the repository, installs Go, runs `danger-go ci`, reads pull request metadata, evaluates the configured rules, and posts a pull request comment when permissions allow it.

### Action Inputs

| Input | Default | Description |
| --- | --- | --- |
| `config` | auto-detect | Path to `.danger.yaml` or `.danger.yml`. |
| `github-token` | `${{ github.token }}` | Token used to read pull request metadata and post comments. |
| `go-version` | `stable` | Go version passed to `actions/setup-go`. |

For pull request comments, the workflow needs:

```yaml
permissions:
  contents: read
  pull-requests: write
  issues: write
```

For pull requests from forks, GitHub may restrict `GITHUB_TOKEN` permissions.

## Local Usage

```sh
go run ./cmd/danger-go local
```

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
