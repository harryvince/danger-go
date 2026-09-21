# Roadmap

`danger-go` is intentionally small today. This roadmap captures likely next steps without promising a specific timeline.

## Provider Support

GitHub Actions is the first supported provider.

Future provider candidates:

- GitLab merge requests.
- Bitbucket pull requests.
- Azure DevOps pull requests.

## Extension Models

Ruby Danger gets much of its power from custom scripts and plugins. `danger-go` now has an initial extension model through command plugins configured in `.danger.yaml`.

Current support:

- External command plugins listed under `plugins`.
- A JSON request is sent to each plugin on standard input.
- A JSON response is read from standard output and converted into `warn` or `fail` messages.
- Optional per-plugin config, level, and timeout settings.

Future extension work may include:

- More built-in YAML rules.
- A stabilized plugin protocol with versioning.
- Example plugins and packaging guidance.
- Richer repository and provider context in plugin requests.
- Starlark scripts.
- WASM plugins.
- Go plugin packages.

Ruby `Dangerfile` compatibility is not a current goal.
