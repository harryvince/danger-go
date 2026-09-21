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

Add a YAML language server annotation to a config file for editor support:

```yaml
# yaml-language-server: $schema=https://harryvince.github.io/danger-go/schema/danger-go.schema.json
level: fail
rules:
  max_changed_files: 50
```

This is a comment, so `danger-go` ignores it while editors can still provide validation and completion. You can also configure your editor/YAML language server to associate the schema with `.danger.yaml` and `.danger.yml`.

This repository validates its sample config with:

```sh
mise run schema:check
```

You can also use the built-in validation command:

```sh
danger-go validate
```

Validate a custom config path:

```sh
danger-go validate --config .github/danger.yaml
```

## Full Example

```yaml
# yaml-language-server: $schema=https://harryvince.github.io/danger-go/schema/danger-go.schema.json
level: fail
rules:
  max_changed_files: 50
  max_changed_lines: 500
  require_pr_title_pattern: "^JIRA-[0-9]+: .+"
  require_linked_issue_pattern: "JIRA-[0-9]+"
  require_conventional_commits: true
  require_squashed_commits:
    enabled: true
    level: warn
  required_labels:
    - ready
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
| `level` | `warn` or `fail` | no | Default level for configured rules. Defaults to each rule's historical behavior when omitted. |
| `rules` | object | no | Rule configuration. If omitted, no rules are evaluated. |

## Rule Levels

Every rule can either use its shorthand value or an object form with `level`.

```yaml
level: warn
rules:
  max_changed_files:
    value: 25
    level: fail
  required_labels:
    values:
      - ready
    level: warn
  warn_dependency_changes:
    enabled: true
    level: fail
```

Allowed levels:

- `warn` reports a warning and does not fail the check.
- `fail` reports a failure and fails the check.

If a rule has its own `level`, that wins. Otherwise `danger-go` uses the top-level `level`. If neither is set, rules keep their original default: most rules fail, while `warn_files`, `warn_dependency_changes`, and `require_squashed_commits: warn` warn.

## Rule Settings

### `max_changed_files`

Type: integer or object

Fails when the number of changed files is greater than the configured value.

```yaml
rules:
  max_changed_files: 50
```

With an explicit level:

```yaml
rules:
  max_changed_files:
    value: 50
    level: warn
```

Set to `0` or omit the setting to disable this rule.

### `max_changed_lines`

Type: integer or object

Fails when the number of added plus deleted lines is greater than the configured value.

```yaml
rules:
  max_changed_lines: 500
```

Set to `0` or omit the setting to disable this rule.

### `require_pr_title_pattern`

Type: string or object

Fails when the pull request title does not match the configured regular expression.

```yaml
rules:
  require_pr_title_pattern: "^JIRA-[0-9]+: .+"
```

With an explicit level:

```yaml
rules:
  require_pr_title_pattern:
    value: "^JIRA-[0-9]+: .+"
    level: warn
```

This uses Go regular expression syntax.

In local mode, the title comes from `DANGER_PR_TITLE`:

```sh
DANGER_PR_TITLE="JIRA-123: add checkout validation" danger-go local
```

In GitHub Actions, the title comes from the pull request event payload.

### `require_linked_issue_pattern`

Type: string or object

Fails when the configured regular expression does not match the pull request title, body, or branch name.

```yaml
rules:
  require_linked_issue_pattern: "JIRA-[0-9]+"
```

This is useful for requiring issue keys such as `JIRA-123` somewhere in the pull request metadata.

In local mode, the searched values come from:

- `DANGER_PR_TITLE`
- `DANGER_PR_BODY`
- `DANGER_PR_BRANCH`

In GitHub Actions, values come from the pull request event payload.

### `require_conventional_commits`

Type: boolean or object

Fails when any pull request commit subject does not follow Conventional Commits.

```yaml
rules:
  require_conventional_commits: true
```

With an explicit level:

```yaml
rules:
  require_conventional_commits:
    enabled: true
    level: warn
```

Accepted examples:

```text
feat: add checkout validation
fix(api): handle missing token
docs!: rewrite configuration guide
```

In local mode, commit subjects come from `DANGER_PR_COMMITS` when set, otherwise `danger-go` reads commits from `origin/main..HEAD` when available:

```sh
DANGER_PR_COMMITS=$'feat: add checkout validation\ntest: cover checkout validation' danger-go local
```

In GitHub Actions, commit subjects come from the pull request commits API.

### `require_squashed_commits`

Type: string or object

Allowed values:

- `warn`
- `fail`

Warns or fails when a pull request has more than one commit.

```yaml
rules:
  require_squashed_commits: warn
```

Use `fail` to deny multi-commit pull requests:

```yaml
rules:
  require_squashed_commits: fail
```

Object form:

```yaml
rules:
  require_squashed_commits:
    enabled: true
    level: warn
```

In local mode, commit subjects come from the same sources as `require_conventional_commits`. In GitHub Actions, the count comes from the pull request commits API.

### `required_labels`

Type: list of strings or object

Fails when a listed label is missing from the pull request.

```yaml
rules:
  required_labels:
    - ready
    - area/docs
```

With an explicit level:

```yaml
rules:
  required_labels:
    values:
      - ready
      - area/docs
    level: warn
```

Label matching is case-insensitive.

In local mode, labels come from comma-separated `DANGER_PR_LABELS`:

```sh
DANGER_PR_LABELS="ready,area/docs" danger-go local
```

In GitHub Actions, labels come from the pull request event payload.

### `required_files`

Type: list of strings or object

Fails when a listed file is missing from the repository.

```yaml
rules:
  required_files:
    - go.mod
    - README.md
```

This is useful for keeping baseline repository files present.

### `required_changed_files`

Type: list of strings or object

Fails when no changed file matches a listed path or glob pattern.

```yaml
rules:
  required_changed_files:
    - docs/**
    - README.md
```

This is useful for policies like requiring documentation updates. Each listed pattern must match at least one changed file.

### `forbidden_files`

Type: list of strings or object

Fails when a changed file matches one of the configured paths or glob patterns.

```yaml
rules:
  forbidden_files:
    - "*.tmp"
    - "secrets.json"
```

Patterns use Go `filepath.Match` behavior. Exact path matches are also supported.

### `warn_files`

Type: list of strings or object

Adds a warning when a changed file matches one of the configured paths or glob patterns.

```yaml
rules:
  warn_files:
    - generated/**
    - "*.lock"
```

Warnings are reported in the PR comment but do not fail the check. Use object form with `level: fail` to turn watched files into failures:

```yaml
rules:
  warn_files:
    values:
      - generated/**
      - "*.lock"
    level: fail
```

### `warn_dependency_changes`

Type: boolean or object

Adds a warning when common dependency manifests or lockfiles change.

```yaml
rules:
  warn_dependency_changes: true
```

Use object form with `level: fail` to fail dependency manifest changes:

```yaml
rules:
  warn_dependency_changes:
    enabled: true
    level: fail
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

`danger-go validate` exits with:

- `0` when the config matches the JSON Schema.
- non-zero when the config is missing, malformed, or does not match the schema.
