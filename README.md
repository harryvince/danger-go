# danger-go

A small Go-native experiment inspired by Danger.

`danger-go` reads `.danger.yaml` or `.danger.yml`, inspects the repository, and reports policy failures. The first version is intentionally config-driven so it can run as a single binary without Ruby.

This repository dogfoods `danger-go` in GitHub Actions.

PR comment posting is verified with a same-repository test pull request.

## Usage

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
