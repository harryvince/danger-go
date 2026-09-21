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
- Documentation in `docs/`, built with Zensical and published to GitHub Pages.

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

## Change Workflow

For new work, create a branch, push it, and raise a GitHub pull request instead of committing directly to `main`. This keeps the dogfood workflow meaningful and verifies the reusable action against real PR events.

Use direct pushes to `main` only for explicitly requested emergency fixes or repository maintenance where the user has clearly asked for it.

Suggested flow:

```sh
git switch -c <type>/<short-description>
# make changes
go test ./...
git push -u origin <type>/<short-description>
gh pr create --title "JIRA-123: short description" --body "..."
```

After opening the PR, watch the dogfood workflow and confirm the `danger-go` comment when the change could affect CI, the action, provider behavior, docs publishing, or config evaluation.

## Core Commands

This repo uses mise for tool and task management. Install tools with:

```sh
mise install
```

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

Validate the sample config against the JSON Schema:

```sh
mise run schema:check
```

Build docs:

```sh
mise run docs:build
```

Preview docs locally:

```sh
mise run docs:serve
```

Build release archives locally:

```sh
mise run release:build
```

## Current Configuration Schema

Config files:

- `.danger.yaml`
- `.danger.yml`

The JSON Schema for config linting is `schema/danger-go.schema.json`. Update it whenever config settings change. The schema is published with the docs at:

```text
https://harryvince.github.io/danger-go/schema/danger-go.schema.json
```

Supported settings live under `rules`:

```yaml
$schema: https://harryvince.github.io/danger-go/schema/danger-go.schema.json
rules:
  max_changed_files: 50
  max_changed_lines: 500
  require_pr_title_pattern: "^JIRA-[0-9]+: .+"
  require_linked_issue_pattern: "JIRA-[0-9]+"
  required_labels:
    - ready
  required_files:
    - go.mod
  required_changed_files:
    - docs/**
  forbidden_files:
    - "*.tmp"
  warn_files:
    - generated/**
  warn_dependency_changes: true
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

## Pipeline and Integration Tests

The repo's own `.danger.yaml` is intentionally small, so dogfooding alone will not cover every feature a downstream repository may rely on.

As functionality grows, add pipeline-oriented tests for consumer-facing behavior even when this repo does not need that behavior itself. Good candidates include:

- action inputs such as custom `config` and `go-version`;
- failing-rule behavior and non-zero exit codes;
- comment posting and comment update behavior;
- missing or restricted GitHub token permissions;
- alternate config filenames;
- GitHub event payload edge cases;
- rules intended for repositories with different layouts.

Prefer fast Go unit tests for pure behavior and real GitHub Actions PR checks for workflow/action behavior. If a feature is mainly useful to consumers, add a targeted workflow or fixture so it is still exercised somewhere in CI.

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

Upstream projects:

- `googleapis/release-please`
- `googleapis/release-please-action`

Do not manually edit generated release PR content unless necessary. Prefer making normal conventional commits to `main` and let Release Please update the PR.

## Docs

Docs live under `docs/` and are built with Zensical.

Current docs:

- `docs/index.md`
- `docs/quickstart.md`
- `docs/configuration.md`
- `docs/github-action.md`
- `docs/examples.md`
- `docs/changelog.md`
- `docs/roadmap.md`

Keep README concise and link deeper docs rather than duplicating everything.

Zensical/mise files:

- `.mise.toml`
- `pyproject.toml`
- `uv.lock`
- `zensical.toml`
- `.github/workflows/docs.yml`
- `scripts/build-docs.sh`

The docs workflow publishes the generated `site` directory to GitHub Pages using GitHub Actions. If Pages deployment fails with a Pages setup error, confirm the repository Pages source is set to GitHub Actions.

`scripts/build-docs.sh` runs Zensical and copies `schema/danger-go.schema.json` into `site/schema/` so the public schema URL stays available.

## Release Artifacts

Release binaries are built by `.github/workflows/release-artifacts.yml` when a GitHub release is published. The workflow runs `scripts/build-release-artifacts.sh`, then uploads `dist/*` to the release.

The script builds these targets:

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`
- `windows/amd64`
- `windows/arm64`

Archives include the `danger-go` binary, `README.md`, and `CHANGELOG.md`. The workflow also uploads `checksums.txt`.

The script sets version metadata with Go ldflags. Keep `internal/version` variable names stable unless you update the script too.

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
- Attempt to create or update a PR comment.
- Comment posting errors are printed but do not fail the check.
- Rule failures do fail the check.

The comment behavior uses a hidden marker to update an existing `danger-go` bot comment when one is present.

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
