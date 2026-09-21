# Configuration

`danger-go` reads one of these files from the repository root:

- `.danger.yaml`
- `.danger.yml`

You can also pass an explicit path:

```sh
danger-go local --config .github/danger.yaml
```

## JSON Schema

`danger-go` publishes a JSON Schema for editor validation and CI linting:

```text
https://harryvince.github.io/danger-go/schema/danger-go.schema.json
```

YAML schema plugins generally consume JSON Schema, so this schema can validate `.danger.yaml` and `.danger.yml` files.

Add `$schema` to a config file for editor support:

```yaml
$schema: https://harryvince.github.io/danger-go/schema/danger-go.schema.json
rules:
  max_changed_files: 50
```

Or configure your editor/YAML language server to associate the schema with `.danger.yaml` and `.danger.yml`.

This repository validates its sample config with:

```sh
mise run schema:check
```

## Full Example

```yaml
rules:
  max_changed_files: 50
  max_changed_lines: 500
  require_pr_title_pattern: "^JIRA-[0-9]+: .+"
  required_files:
    - go.mod
    - README.md
  required_changed_files:
    - docs/**
  forbidden_files:
    - "*.tmp"
    - "secrets.json"
  warn_files:
    - generated/**
  warn_dependency_changes: true
```

## Top-Level Settings

| Setting | Type | Required | Description |
| --- | --- | --- | --- |
| `rules` | object | no | Rule configuration. If omitted, no rules are evaluated. |

## Rule Settings

### `max_changed_files`

Type: integer

Fails when the number of changed files is greater than the configured value.

```yaml
rules:
  max_changed_files: 50
```

Set to `0` or omit the setting to disable this rule.

### `max_changed_lines`

Type: integer

Fails when the number of added plus deleted lines is greater than the configured value.

```yaml
rules:
  max_changed_lines: 500
```

Set to `0` or omit the setting to disable this rule.

### `require_pr_title_pattern`

Type: string

Fails when the pull request title does not match the configured regular expression.

```yaml
rules:
  require_pr_title_pattern: "^JIRA-[0-9]+: .+"
```

This uses Go regular expression syntax.

In local mode, the title comes from `DANGER_PR_TITLE`:

```sh
DANGER_PR_TITLE="JIRA-123: add checkout validation" danger-go local
```

In GitHub Actions, the title comes from the pull request event payload.

### `required_files`

Type: list of strings

Fails when a listed file is missing from the repository.

```yaml
rules:
  required_files:
    - go.mod
    - README.md
```

This is useful for keeping baseline repository files present.

### `required_changed_files`

Type: list of strings

Fails when no changed file matches a listed path or glob pattern.

```yaml
rules:
  required_changed_files:
    - docs/**
    - README.md
```

This is useful for policies like requiring documentation updates. Each listed pattern must match at least one changed file.

### `forbidden_files`

Type: list of strings

Fails when a changed file matches one of the configured paths or glob patterns.

```yaml
rules:
  forbidden_files:
    - "*.tmp"
    - "secrets.json"
```

Patterns use Go `filepath.Match` behavior. Exact path matches are also supported.

### `warn_files`

Type: list of strings

Adds a warning when a changed file matches one of the configured paths or glob patterns.

```yaml
rules:
  warn_files:
    - generated/**
    - "*.lock"
```

Warnings are reported in the PR comment but do not fail the check.

### `warn_dependency_changes`

Type: boolean

Adds a warning when common dependency manifests or lockfiles change.

```yaml
rules:
  warn_dependency_changes: true
```

Currently watched files include common Go, Node, Ruby, Rust, and Python dependency files such as `go.mod`, `package.json`, `Gemfile.lock`, `Cargo.lock`, and `requirements.txt`.

## Path Patterns

Rules that accept file patterns support:

- exact paths like `README.md`;
- standard glob patterns like `*.tmp`;
- directory-prefix patterns ending in `/**`, such as `docs/**`.

## Exit Codes

`danger-go` exits with:

- `0` when no configured rules fail.
- non-zero when one or more rules fail.
- non-zero when configuration or provider setup cannot be read.

GitHub comment posting errors are reported but do not fail the check. Rule results decide whether the check passes or fails.
