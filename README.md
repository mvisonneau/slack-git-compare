# slack-git-compare

[![GoDoc](https://godoc.org/github.com/mvisonneau/slack-git-compare?status.svg)](https://godoc.org/github.com/mvisonneau/slack-git-compare)
[![Go Report Card](https://goreportcard.com/badge/github.com/mvisonneau/slack-git-compare)](https://goreportcard.com/report/github.com/mvisonneau/slack-git-compare)
[![Docker Pulls](https://img.shields.io/docker/pulls/mvisonneau/slack-git-compare.svg)](https://hub.docker.com/r/mvisonneau/slack-git-compare/)
[![Build Status](https://github.com/mvisonneau/slack-git-compare/workflows/test/badge.svg?branch=main)](https://github.com/mvisonneau/slack-git-compare/actions)
[![Coverage Status](https://coveralls.io/repos/github/mvisonneau/slack-git-compare/badge.svg?branch=main)](https://coveralls.io/github/mvisonneau/slack-git-compare?branch=main)

This is a slack command handler to compare git refs from whether GitHub or GitLab, within Slack

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

## Terminology

`slack-git-compare` is a conjugation of the verb [approuver](https://www.larousse.fr/conjugaison/francais/approuver/518) in French 🇫🇷, equivalent to `approve` in English 🇬🇧
