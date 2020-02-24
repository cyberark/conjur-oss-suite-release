# Contributing to the Conjur OSS Suite

## Table of Contents

- [Prerequisites](#prerequisites)
- [Pull Request Workflow](#pull-request-workflow)
- [Style Guide](#style-guide)
- [Building](#building)
- [Running](#running)
- [Testing](#testing)
- [Releasing](#releasing)

## Prerequisites

### Go version
To work in this codebase, you will want to have at least Go v1.13 installed. The
code may work on older versions but it has not been tested nor evaluated to be
compatible for those configurations. We expect at the very least that you will
need Go modules support so you will at minimum need Go v1.11+.

## Pull Request Workflow

1. Search the [open issues][issues] in GitHub to find out what has been planned
2. Select an existing issue or open an issue to propose changes or fixes
3. Add the `implementing` label to the issue as you begin to work on it
4. Run tests as described [here][tests], ensuring they pass
5. Submit a pull request, linking the issue in the description (e.g. `Connected to #123`)
6. Add the `implemented` label to the issue, and ask another contributor to review and merge your code

## Style guide

Use [this guide][style] to maintain consistent style across the project.

[issues]: https://github.com/cyberark/conjur-oss-suite-release/issues
[style]: STYLE.md
[tests]: #testing

## Building

Clone `https://github.com/cyberark/conjur-oss-suite-release`. If you are new to Go,
be aware that Go can be very selective about where the files are placed on the filesystem.
There is an environment variable called `GOPATH`, whose default value
is `~/go`. Conjur OSS Suite uses [go modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more)
which require either that you clone this repository outside of your `GOPATH` or you set the
`GO111MODULE` environment variable to `on`. We recommend cloning this repository
 outside of your `GOPATH`.

Once you've cloned the repository, you can build and/or run the included code.

## Running

Currently the only functionality included is the `CHANGELOG.md` generation which can be
done with:
```
$ ./parse-changelog
```

## Testing

### Prerequisites

- Docker

### Running all tests

```sh-session
$ go test -v ./...
```

### Running only unit (short) tests

```sh-session
$ go test -v -short ./...
```

## Releasing

Releases are automatically prepared using GitHub actions [here](.github/workflows/release.yml).
In general terms, whenever a tag with `v*` pattern is pushed up, the following steps
are automatically run:
- Tests are re-run
- If tests fail, release process is halted
- Release notes and changelog files are generated
- Draft release is created on GitHub
- Release notes and the changelog are attached as assets to that release

To view the progress of the actions, you can take a look at [this](https://github.com/cyberark/conjur-oss-suite-release/actions) page.

After these steps run and they encounter no problems, manual publishing of release will make it
public for consumers of this repo.
