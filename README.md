# 🗄️ slack-git-compare

[![PkgGoDev](https://pkg.go.dev/badge/github.com/mvisonneau/slack-git-compare)](https://pkg.go.dev/mod/github.com/mvisonneau/slack-git-compare)
[![Go Report Card](https://goreportcard.com/badge/github.com/mvisonneau/slack-git-compare)](https://goreportcard.com/report/github.com/mvisonneau/slack-git-compare)
[![test](https://github.com/mvisonneau/slack-git-compare/actions/workflows/test.yml/badge.svg)](https://github.com/mvisonneau/slack-git-compare/actions/workflows/test.yml)
[![Coverage Status](https://coveralls.io/repos/github/mvisonneau/slack-git-compare/badge.svg?branch=main)](https://coveralls.io/github/mvisonneau/slack-git-compare?branch=main)
[![release](https://github.com/mvisonneau/slack-git-compare/actions/workflows/release.yml/badge.svg)](https://github.com/mvisonneau/slack-git-compare/actions/workflows/release.yml)
[![slack-git-compare](https://snapcraft.io/slack-git-compare/badge.svg)](https://snapcraft.io/slack-git-compare)

This is a slack command handler to compare git refs from whether `GitHub` or `GitLab`, within **Slack**

## How it works

This repositories holds the code of a daemon which communicates with git providers and can handle Slack interactions.

![architecture](/docs/images/architecture.png)

## Limitations / Known issues

- For readability purposes, we currently only display up to 10 commits info per diff
- On repositories with large amount of references, API calls may not be returned in less than 3s which causes Slack to dismiss the response.
- There is currently no linking being done between commiters and slack users
- Repositories are only loaded/refreshed at the start of the process

## Install

### Go

```bash
~$ go get -u github.com/mvisonneau/slack-git-compare/cmd/slack-git-compare
```

### Snapcraft

```bash
~$ snap install slack-git-compare
```

### Homebrew

```bash
~$ brew install mvisonneau/tap/slack-git-compare
```

### Docker

```bash
~$ docker run -it --rm docker.io/mvisonneau/slack-git-compare
or
~$ docker run -it --rm ghcr.io/mvisonneau/slack-git-compare
```

### Scoop

```bash
~$ scoop bucket add https://github.com/mvisonneau/scoops
~$ scoop install slack-git-compare
```

### Binaries, DEB and RPM packages

Have a look onto the [latest release page](https://github.com/mvisonneau/slack-git-compare/releases/latest) to pick your flavor and version. Here is an helper to fetch the most recent one:

```bash
~$ export SGC_VERSION=$(curl -s "https://api.github.com/repos/mvisonneau/slack-git-compare/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
```

```bash
# Binary (eg: linux/amd64)
~$ wget https://github.com/mvisonneau/slack-git-compare/releases/download/${SGC_VERSION}/slack-git-compare_${SGC_VERSION}_linux_amd64.tar.gz
~$ tar zxvf slack-git-compare_${SGC_VERSION}_linux_amd64.tar.gz -C /usr/local/bin

# DEB package (eg: linux/386)
~$ wget https://github.com/mvisonneau/slack-git-compare/releases/download/${SGC_VERSION}/slack-git-compare_${SGC_VERSION}_linux_386.deb
~$ dpkg -i slack-git-compare_${SGC_VERSION}_linux_386.deb

# RPM package (eg: linux/arm64)
~$ wget https://github.com/mvisonneau/slack-git-compare/releases/download/${SGC_VERSION}/slack-git-compare_${SGC_VERSION}_linux_arm64.rpm
~$ rpm -ivh slack-git-compare_${SGC_VERSION}_linux_arm64.rpm
```

### HELM

If you want to make it run on [kubernetes](https://kubernetes.io/), there is a [helm chart](https://github.com/mvisonneau/helm-charts/tree/main/charts/slack-git-compare) available for this purpose.

You can check the chart's [values.yml](https://github.com/mvisonneau/helm-charts/blob/main/charts/slack-git-compare/values.yaml) for complete configuration options.

```bash
# Add the helm repository to your local client
~$ helm repo add mvisonneau https://charts.visonneau.fr

# Configure a minimal configuration for the exporter
~$ cat <<EOF > values.yml
config:
  SGC_SLACK_TOKEN: xoxb-123456789-xxx-xxx
  SGC_SLACK_SIGNING_SECRET: 123456789xxxxxx
  SGC_GITHUB_TOKEN: xxxxx
  SGC_GITHUB_ORG: cilium,foo,bar
  SGC_GITLAB_TOKEN: xxxx
  SGC_GITLAB_GROUP: gitlab-org,foo,bar
EOF

# Release the chart on your Kubernetes cluster
~$ helm upgrade -i slack-git-compare mvisonneau/slack-git-compare -f values.yml
```

## Examples / Getting Started

Here is documentation about [how to get started](examples/quickstart) with the tool.

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
