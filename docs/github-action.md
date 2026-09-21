# GitHub Action

`danger-go` ships as a reusable GitHub Action.

## Basic Workflow

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

## Inputs

| Input | Default | Description |
| --- | --- | --- |
| `config` | auto-detect | Path to `.danger.yaml` or `.danger.yml`. |
| `github-token` | automatic `${{ github.token }}` | Optional token used to read pull request metadata and post comments. |
| `go-version` | `stable` | Go version passed to `actions/setup-go`. |

## Token Behavior

You do not need to create a personal access token for normal same-repository pull requests.

GitHub automatically provides `${{ github.token }}` to workflow runs. The action uses that token by default. The workflow `permissions` block controls what the token can do.

For pull request comments, use:

```yaml
permissions:
  contents: read
  pull-requests: write
  issues: write
```

Why these permissions are needed:

- `contents: read` allows repository checkout and inspection.
- `pull-requests: write` allows pull request operations.
- `issues: write` allows pull request comments, because GitHub exposes PR comments through the Issues comments API.

Use `github-token` only when you intentionally want a different credential, such as a GitHub App token.

For pull requests from forks, GitHub may restrict the automatic token. In that case, `danger-go` can still run checks, but comment posting may be skipped or denied by GitHub.

## Custom Config

```yaml
jobs:
  danger-go:
    runs-on: ubuntu-latest
    steps:
      - uses: harryvince/danger-go@v1
        with:
          config: .github/danger.yaml
```

## Custom Go Version

```yaml
jobs:
  danger-go:
    runs-on: ubuntu-latest
    steps:
      - uses: harryvince/danger-go@v1
        with:
          go-version: "1.26"
```

## What the Action Does

The action:

1. Checks out the repository.
2. Installs Go.
3. Runs `danger-go ci`.
4. Reads GitHub Actions pull request metadata.
5. Evaluates rules from the config file.
6. Creates or updates a summary comment when GitHub allows it.
7. Fails the workflow if any rule fails.
