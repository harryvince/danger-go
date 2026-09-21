# AGENTS.md

Guidance for future coding agents working on `danger-go`.

## Project Summary

`danger-go` is a Go-native pull request policy runner inspired by Danger. It is intentionally not a Ruby Danger compatibility layer.

The current product shape is:

- Config-driven rules from `.danger.yaml` or `.danger.yml`.
- A Go CLI at `cmd/danger-go`.
- A reusable composite GitHub Action in `action.yml`.
- GitHub Actions support first.
- Release automation through Release Please.
- Documentation in `docs/`.

The public repo is:

```text
https://github.com/harryvince/danger-go
```

## Important Conventions

Use Conventional Commits. Release Please depends on them.

Examples:

```text
feat: add reusable github action
fix: handle missing pull request payload
docs: clarify default github token usage
ci: update dogfood workflow
test: verify github action comment posting
```

Avoid ticket-prefix commit subjects like `JIRA-123: ...` unless they come after the conventional prefix in the body or scope. PR titles may still need to satisfy `.danger.yaml`.

## Core Commands

Run tests:

```sh
go test ./...
```

Run local checks:

```sh
DANGER_PR_TITLE="JIRA-123: local check" go run ./cmd/danger-go local
```

Run local checks with explicit config:

```sh
DANGER_PR_TITLE="JIRA-123: local check" go run ./cmd/danger-go local --config .danger.yaml
```

The sample config requires a non-empty PR title, so set `DANGER_PR_TITLE` for local smoke tests.

## Current Configuration Schema

Config files:

- `.danger.yaml`
- `.danger.yml`

Supported settings live under `rules`:

```yaml
rules:
  max_changed_files: 50
  require_pr_title_pattern: "^JIRA-[0-9]+: .+"
  required_files:
    - go.mod
  forbidden_files:
    - "*.tmp"
```

Implementation references:

- `internal/config/config.go`
- `internal/danger/danger.go`

Update `docs/configuration.md` whenever config settings change.

## GitHub Action Behavior

The reusable action is defined in `action.yml`.

Users should be able to use:

```yaml
steps:
  - uses: harryvince/danger-go@v1
```

They do not need to pass a token for normal same-repository pull requests. The action defaults to `${{ github.token }}`:

```yaml
GITHUB_TOKEN: ${{ inputs.github-token || github.token }}
```

Required workflow permissions for PR comments:

```yaml
permissions:
  contents: read
  pull-requests: write
  issues: write
```

Reasoning:

- `contents: read` lets the action check out and inspect the repository.
- `pull-requests: write` allows PR operations.
- `issues: write` is required because GitHub PR comments use the Issues comments API.

Forked PRs may still have restricted token permissions. Do not promise comment posting always works for forks.

## Dogfood Workflow

The repo dogfoods the action in `.github/workflows/danger-go.yml`.

It intentionally uses:

```yaml
- uses: ./
```

This validates the local composite action on PRs. Keep it aligned with the README recommended usage.

When changing `action.yml` or workflow behavior, prefer testing through a real PR:

1. Create a branch.
2. Make the change.
3. Push and open a PR with a title that passes `.danger.yaml`, for example `JIRA-127: verify default token`.
4. Watch the run:

```sh
gh run list --branch <branch> --limit 3
gh run watch <run-id> --interval 10 --exit-status
```

5. Confirm the PR comment:

```sh
gh pr view <number> --comments
```

Expected comment body:

```md
## danger-go

No issues found.
```

## Known GitHub Actions Gotchas

The repo setting "Allow GitHub Actions to create and approve pull requests" was required for Release Please to create release PRs. It was enabled with:

```sh
gh api --method PUT repos/harryvince/danger-go/actions/permissions/workflow \
  -f default_workflow_permissions=write \
  -F can_approve_pull_request_reviews=true
```

If Release Please starts failing with:

```text
GitHub Actions is not permitted to create or approve pull requests.
```

check that repo setting first.

GitHub may show runner annotations about Node 20 deprecation or `ubuntu-latest` migration. Those are currently not project failures.

## Release Please

Release Please is configured with:

- `.github/workflows/release-please.yml`
- `release-please-config.json`
- `.release-please-manifest.json`

Open release PR at the time this file was written:

```text
https://github.com/harryvince/danger-go/pull/3
```

Do not manually edit generated release PR content unless necessary. Prefer making normal conventional commits to `main` and let Release Please update the PR.

## Docs

Docs live under `docs/` and are intended to be friendly to a future static docs site.

Current docs:

- `docs/index.md`
- `docs/quickstart.md`
- `docs/configuration.md`
- `docs/github-action.md`
- `docs/roadmap.md`

Keep README concise and link deeper docs rather than duplicating everything.

The user has said they may later ask to set up Zensical for publishing docs to GitHub Pages.

## Implementation Notes

Key packages:

- `internal/cli`: command dispatch and mode selection.
- `internal/config`: config discovery and YAML parsing.
- `internal/danger`: rule evaluation and reports.
- `internal/git`: local repository inspection.
- `internal/github`: GitHub Actions context, GitHub API reads, and PR comments.

Current `ci` behavior:

- If running in GitHub Actions, read the PR event payload.
- Fetch changed files from the GitHub API.
- Merge local repository file information where available for file-existence rules.
- Evaluate rules.
- Attempt to post a PR comment.
- Comment posting errors are printed but do not fail the check.
- Rule failures do fail the check.

The comment behavior currently creates a new comment. A likely improvement is updating an existing `danger-go` comment instead.

## History Notes

Early commit history was rewritten to conventional subjects while the repo was brand new. Avoid rewriting public history again unless explicitly requested.

Recent validated behaviors:

- The reusable action works through `uses: ./` in the dogfood PR workflow.
- The action can use the default `${{ github.token }}` without passing `github-token`.
- Same-repository PRs can post comments when workflow permissions are set correctly.

## Before Finishing Work

For code changes:

```sh
go test ./...
```

For action/workflow changes, also validate through a PR when feasible.

For docs/config changes, update the relevant files under `docs/` and README links if needed.

Check status before and after:

```sh
git status --short --branch
```
