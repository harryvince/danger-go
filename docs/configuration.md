# Configuration

`danger-go` reads one of these files from the repository root:

- `.danger.yaml`
- `.danger.yml`

You can also pass an explicit path:

```sh
danger-go local --config .github/danger.yaml
```

## Full Example

```yaml
rules:
  max_changed_files: 50
  require_pr_title_pattern: "^JIRA-[0-9]+: .+"
  required_files:
    - go.mod
    - README.md
  forbidden_files:
    - "*.tmp"
    - "secrets.json"
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

## Exit Codes

`danger-go` exits with:

- `0` when no configured rules fail.
- non-zero when one or more rules fail.
- non-zero when configuration or provider setup cannot be read.

GitHub comment posting errors are reported but do not fail the check. Rule results decide whether the check passes or fails.
