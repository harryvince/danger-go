# Changelog

`danger-go` uses Google's Release Please project to generate changelogs from Conventional Commits.

The canonical changelog is maintained in [`CHANGELOG.md`](https://github.com/harryvince/danger-go/blob/main/CHANGELOG.md).

Latest release:

- [danger-go v1.0.0](https://github.com/harryvince/danger-go/releases/tag/danger-go-v1.0.0)

Release automation:

- [googleapis/release-please](https://github.com/googleapis/release-please)
- [googleapis/release-please-action](https://github.com/googleapis/release-please-action)

## How Releases Work

Release Please watches commits on `main`, groups Conventional Commits, and opens or updates a release pull request.

When the release PR is merged, Release Please will:

- update `CHANGELOG.md`;
- update `.release-please-manifest.json`;
- create a GitHub release;
- tag the release.

Use Conventional Commits so changes appear in the right changelog section.

## Current Changelog

See [`CHANGELOG.md`](https://github.com/harryvince/danger-go/blob/main/CHANGELOG.md) for the current generated changelog.
