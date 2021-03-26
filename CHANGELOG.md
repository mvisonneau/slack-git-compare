# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [0ver](https://0ver.org) (more or less).

## [Unreleased]

### Added

- Map slack users and git authors for an enhanced experience
- Refresh the cached repositories and refs list from the modal
- Configure and schedule automatic cache updates

### Changed

- Moved configuration flags into a configuration file instead
- Got bigger selectors in the Slack modal
- Removed the conversation selector from the modal, fallback the location of the slash command instead

## [v0.0.1] - 2021-03-22

### Added

- GitHub support (.com & self-hosted)
- GitLab support (.com & self-hosted)
- Fuzzy search of repositories and references
- In-memory storage/caching
- Some tests
- CI and project boilerplating

[Unreleased]: https://github.com/mvisonneau/slack-git-compare/compare/v0.0.1...HEAD
[v0.0.1]: https://github.com/mvisonneau/slack-git-compare/tree/v0.0.1
