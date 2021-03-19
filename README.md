# 🗄️ slack-git-compare

[![PkgGoDev](https://pkg.go.dev/badge/github.com/mvisonneau/slack-git-compare)](https://pkg.go.dev/mod/github.com/mvisonneau/slack-git-compare)
[![Go Report Card](https://goreportcard.com/badge/github.com/mvisonneau/slack-git-compare)](https://goreportcard.com/report/github.com/mvisonneau/slack-git-compare)
[![test](https://github.com/mvisonneau/slack-git-compare/actions/workflows/test.yml/badge.svg)](https://github.com/mvisonneau/slack-git-compare/actions/workflows/test.yml)
[![Coverage Status](https://coveralls.io/repos/github/mvisonneau/slack-git-compare/badge.svg?branch=main)](https://coveralls.io/github/mvisonneau/slack-git-compare?branch=main)
[![release](https://github.com/mvisonneau/slack-git-compare/actions/workflows/release.yml/badge.svg)](https://github.com/mvisonneau/slack-git-compare/actions/workflows/release.yml)
[![slack-git-compare](https://snapcraft.io/slack-git-compare/badge.svg)](https://snapcraft.io/slack-git-compare)

This is a slack command handler to compare git refs from whether `GitHub` or `GitLab`, within **Slack**

## Develop / Test

```bash
~$ make build
~$ ./slack-git-compare
```

## Build / Release

If you want to build and/or release your own version of `slack-git-compare`, you need the following prerequisites :

- [git](https://git-scm.com/)
- [golang](https://golang.org/)
- [make](https://www.gnu.org/software/make/)
- [goreleaser](https://goreleaser.com/)

```bash
~$ git clone git@github.com:mvisonneau/slack-git-compare.git && cd slack-git-compare

# Build the binaries locally
~$ make build

# Build the binaries and release them (you will need a GITHUB_TOKEN and to reconfigure .goreleaser.yml)
~$ make release
```

## Contribute

Contributions are more than welcome! Feel free to submit a [PR](https://github.com/mvisonneau/slack-git-compare/pulls).
