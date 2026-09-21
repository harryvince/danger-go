# Roadmap

`danger-go` is intentionally small today. This roadmap captures likely next steps without promising a specific timeline.

## Possible Rule Ideas

- Require labels.
- Require linked issue keys.

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
