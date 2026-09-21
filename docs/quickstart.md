# Quickstart

## 1. Add Configuration

Create `.danger.yaml` in your repository:

```yaml
rules:
  max_changed_files: 50
  require_pr_title_pattern: ".+"
  required_files:
    - go.mod
  forbidden_files:
    - "*.tmp"
```

`.danger.yml` is also supported.

## 2. Add the GitHub Action

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

## 3. Open a Pull Request

When a pull request opens or updates, `danger-go` will:

1. Read the repository config.
2. Fetch pull request metadata from GitHub.
3. Evaluate the configured rules.
4. Create or update a pull request comment when permitted.
5. Fail the check if any rule fails.

## Local Runs

You can run checks locally:

```sh
go run github.com/harryvince/danger-go/cmd/danger-go@latest local
```

Local mode inspects the current Git repository. If your config validates pull request titles, set `DANGER_PR_TITLE`:

```sh
DANGER_PR_TITLE="JIRA-123: update billing flow" go run github.com/harryvince/danger-go/cmd/danger-go@latest local
```

## Custom Config Path

Use `config` with the action:

```yaml
- uses: harryvince/danger-go@v1
  with:
    config: .github/danger.yaml
```

Or use `--config` locally:

```sh
go run github.com/harryvince/danger-go/cmd/danger-go@latest local --config .github/danger.yaml
```

## Version

Print the installed version:

```sh
go run github.com/harryvince/danger-go/cmd/danger-go@latest version
```

## Release Binaries

GitHub releases include archived binaries for:

- Linux amd64 and arm64
- macOS amd64 and arm64
- Windows amd64 and arm64

Download them from the latest release:

https://github.com/harryvince/danger-go/releases/latest
