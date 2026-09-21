# Examples

These examples show common ways to configure `danger-go`.

## Small Pull Requests

Fail pull requests that change more than 25 files:

```yaml
rules:
  max_changed_files: 25
```

This is useful for teams that prefer smaller, easier-to-review changes.

## Large Line Diffs

Fail pull requests that add or delete more than 500 lines:

```yaml
rules:
  max_changed_lines: 500
```

## Ticket-Based Titles

Require a ticket key at the start of every pull request title:

```yaml
rules:
  require_pr_title_pattern: "^JIRA-[0-9]+: .+"
```

Matching titles:

```text
JIRA-123: add checkout validation
JIRA-9876: update release workflow
```

Non-matching titles:

```text
Add checkout validation
JIRA-123 add checkout validation
```

## Required Repository Files

Require baseline files to exist:

```yaml
rules:
  required_files:
    - README.md
    - go.mod
```

This checks the repository contents, not just the pull request diff.

## Required Documentation Changes

Require documentation to change with the pull request:

```yaml
rules:
  required_changed_files:
    - docs/**
    - README.md
```

Each listed pattern must match at least one changed file.

## Forbidden Changed Files

Fail when generated, local, or sensitive files are changed:

```yaml
rules:
  forbidden_files:
    - "*.tmp"
    - ".env"
    - "secrets.json"
```

Patterns use Go `filepath.Match` behavior. Exact path matches are also supported.

## Warning-Only Files

Warn when generated files change without failing the check:

```yaml
rules:
  warn_files:
    - generated/**
```

## Dependency Manifest Warnings

Warn when common dependency manifests or lockfiles change:

```yaml
rules:
  warn_dependency_changes: true
```

## Combined Policy

```yaml
rules:
  max_changed_files: 50
  max_changed_lines: 500
  require_pr_title_pattern: "^JIRA-[0-9]+: .+"
  required_files:
    - README.md
    - go.mod
  required_changed_files:
    - docs/**
  forbidden_files:
    - "*.tmp"
    - ".env"
    - "secrets.json"
  warn_files:
    - generated/**
  warn_dependency_changes: true
```

## Custom Config Location

Use a non-default config path:

```yaml
jobs:
  danger-go:
    runs-on: ubuntu-latest
    steps:
      - uses: harryvince/danger-go@v1
        with:
          config: .github/danger.yaml
```

## Different Go Version

The action installs Go before running `danger-go`. Override the version when needed:

```yaml
jobs:
  danger-go:
    runs-on: ubuntu-latest
    steps:
      - uses: harryvince/danger-go@v1
        with:
          go-version: "1.26"
```

## Local Validation

Run the same config locally:

```sh
DANGER_PR_TITLE="JIRA-123: local validation" go run github.com/harryvince/danger-go/cmd/danger-go@latest local
```

Use an explicit config path locally:

```sh
DANGER_PR_TITLE="JIRA-123: local validation" go run github.com/harryvince/danger-go/cmd/danger-go@latest local --config .github/danger.yaml
```
