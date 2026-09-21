# danger-go

`danger-go` is a Go-native pull request policy runner inspired by Danger.

It gives a repository a small, versioned rules file, runs those rules in CI, and reports the result back to the pull request. The goal is to keep the useful parts of Danger while avoiding a Ruby runtime requirement.

## What It Does

`danger-go` can currently:

- Read `.danger.yaml` or `.danger.yml`.
- Inspect changed files in a local Git repository.
- Read pull request title and changed files from GitHub Actions.
- Fail a CI check when configured rules fail.
- Post a pull request comment through GitHub Actions when permissions allow it.
- Run as a reusable GitHub Action.

## Project Shape

The project is intentionally config-first. Instead of executing a Ruby `Dangerfile`, it uses YAML rules that are easy to review and safe to run in CI.

Current rule categories are deliberately small:

- Pull request size limits.
- Pull request title validation.
- Required repository files.
- Forbidden changed file patterns.

See [Configuration](configuration.md) for every supported setting.

## Design Goals

- Single Go toolchain, no Ruby runtime.
- Clear repository-local policy in `.danger.yaml`.
- Good GitHub Actions support first.
- Predictable behavior in CI.
- Changelog-friendly development through Conventional Commits and Release Please.

## Current Limitations

`danger-go` is early. It is not a drop-in replacement for Ruby Danger.

It does not currently support:

- Ruby `Dangerfile` execution.
- Danger plugins.
- Inline review comments.
- GitLab, Bitbucket, or Azure DevOps.
- Custom user-defined scripts.

Those are possible future directions, but the current project is a practical Go-native subset rather than a compatibility layer.

## Next Steps

- [Quickstart](quickstart.md)
- [GitHub Action](github-action.md)
- [Configuration](configuration.md)
- [Roadmap](roadmap.md)
