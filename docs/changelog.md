# Changelog

`danger-go` uses Release Please to generate changelogs from Conventional Commits.

The changelog will be published in `CHANGELOG.md` after the first Release Please release PR is merged.

Until then, release preparation is tracked in the open Release Please PR:

https://github.com/harryvince/danger-go/pull/3

## How Releases Work

Release Please watches commits on `main`, groups Conventional Commits, and opens or updates a release pull request.

When the release PR is merged, Release Please will:

- update `CHANGELOG.md`;
- update `.release-please-manifest.json`;
- create a GitHub release;
- tag the release.

Use Conventional Commits so changes appear in the right changelog section.
