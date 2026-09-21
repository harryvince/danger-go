# Roadmap

`danger-go` is intentionally small today. This roadmap captures likely next steps without promising a specific timeline.

## Near-Term

- Improve GitHub comment behavior so repeated runs update an existing comment instead of adding a new one.
- Add more built-in rules.
- Add better documentation examples.
- Add a release artifact for users who prefer a binary instead of `go run`.
- Add a version command.

## Possible Rule Ideas

- Require changed documentation for selected paths.
- Limit pull request size by line count.
- Require labels.
- Require linked issue keys.
- Warn on generated files.
- Warn on dependency manifest changes.

## Provider Support

GitHub Actions is the first supported provider.

Future provider candidates:

- GitLab merge requests.
- Bitbucket pull requests.
- Azure DevOps pull requests.

## Extension Models

Ruby Danger gets much of its power from custom scripts and plugins. `danger-go` may eventually support an extension model, but it should stay safe and easy to run in CI.

Possible approaches:

- More built-in YAML rules.
- Starlark scripts.
- WASM plugins.
- Go plugin packages.

Ruby `Dangerfile` compatibility is not a current goal.
