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

### Usage

- Edit the file `suite.yml` as needed
- Run the CHANGELOG generator:
```
./parse-changelogs
```
- Resulting changelog will be placed in `CHANGELOG.md`

### Advanced usage

The CLI accepts the following arguments/parameters:
```
  -f string
        Repository YAML file to parse (default "suite.yml")
  -o string
        Output filename
  -p string
        GitHub API token. This can also be passed in as the 'GITHUB_TOKEN' environment variable. The flag takes precedence.
  -r string
        Directory of releases (containinng 'suite_<semver>.yml') files. Set this to empty string to skip suite version diffing. (default "releases")
  -t string
        Output type. Only accepts 'changelog', 'docs-release', 'release', and 'unreleased'. (default "changelog")
  -v string
        Version to embed in the changelog (default "Unreleased")
```

## Testing

### Prerequisites

- Docker

### Running all tests

```sh-session
$ go test -v ./...
```

Note: if you're running all the tests a lot during local development, you
may want to run the tests after setting the `GITHUB_TOKEN` env var, so that
you won't run up against GitHub API limits.

If you have your GitHub API token saved in your keychain, you may also want
to use [Summon](https://cyberark.github.io/summon) with the Keychain provider
and run the test command instead as something like:
```sh-session
summon -p keyring.py \
  --yaml 'GITHUB_TOKEN: !var github/api_token' \
  bash -c 'go test -v ./...'
```

### Running only unit (short) tests

```sh-session
$ go test -v -short ./...
```

## Releasing

1. Determine whether there are component changes since the last suite release
   that merit a new suite release.

   - Check the [wiki](https://github.com/cyberark/conjur-oss-suite-release/wiki/Unreleased-Changes)
     to see the daily report on which components have had new tagged
     versions since the last release, and which components have unreleased changes.
     - Note: Entries for components with unreleased changes (changes on the master
       branch that are not yet available in a GitHub release) show in this report as
       `org/repo @HEAD`. The links take you to the commit history for all commits on
       master that are not included in the latest GitHub release.

   - If there are any components with unreleased changes that should be tagged,
     open an issue in that component's repository for adding a new tag.

1. Ensure the components have green builds.

   - Check the [Jenkins dashboard](https://jenkins.conjur.net/view/OSS%20Suite%20Components/)
     to make sure there are no ongoing build failures for any of the OSS suite
     components.

     Note: The Jenkins dashboard does not include the following components at
     this time:
     - [Jenkins plugin](https://github.com/cyberark/conjur-credentials-plugin)

1. Update the versions included in the suite release.

   - Edit the [suite release config](https://github.com/cyberark/conjur-oss-suite-release/blob/master/suite.yml)
     to bump the versions of any components with new tags and/or to add any new components
     to the next suite release.

   - [Submit your changes in a pull request (PR)](https://docs.joomla.org/Using_the_Github_UI_to_Make_Pull_Requests)
     as per our [contributor guidelines](https://github.com/cyberark/community).
     - **Important:** the PR description **must** include the suite release version (following
       [semantic versioning](https://semver.org/) of the new suite. The maintainers
       of this project will use this info to complete the release.
     - The PR to modify the `suite.yml` will automatically kick off the end-to-end
       tests for the suite against the pinned suite component versions. If the tests
       don't pass, they'll need to be fixed before the new suite release can be created.
     - To see the status of the automated tests, you can check the
       [status tab](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/about-status-checks)
       in the pull request.

   - A maintainer of this project will review the PR to make sure the release is
     ready to move forward. In particular, they will do the following before
     approving and merging the PR:
     - Check that the PR description includes the desired suite release version.
     - Review the status of the automated tests, to make sure they are passing.
     - Check the [Jenkins dashboard](https://jenkins.conjur.net/view/OSS%20Suite%20Components/)
       to make sure there are no ongoing build failures for any of the OSS suite
       components.

   - Once the changes to update the suite are approved and merged to master, the
     maintainer will create a new git tag. Creating a git tag (as outlined in the
     [maintainer docs](https://github.com/cyberark/community/blob/master/Conjur/CONTRIBUTING.md#tagging)):
     - Re-runs the automated end-to-end tests against the current pinned versions
       in `suite.yml`
     - Auto-generates HTML release notes for the docs website
     - Auto-generates a draft GitHub release; that is, the automated process:
       - Creates the suite `CHANGELOG.md`
       - Creates GitHub release notes
       - Creates draft GitHub release populated with release notes and with attached
         artifacts, including `CHANGELOG.md`, `suite.yml`, HTML release notes, etc
     - Note: To view the progress of the GitHub actions that automatically run post-tag,
       you can take a look at
       [this page](https://github.com/cyberark/conjur-oss-suite-release/actions).

 1. [Publish](https://help.github.com/en/github/administering-a-repository/managing-releases-in-a-repository)
    the [draft GitHub release](https://github.com/cyberark/conjur-oss-suite-release/releases/).
    If additional validation is needed, it can be initially published as a pre-release
    and promoted to a full release once final validation is complete.

    Publishing the release:
    - Runs the end-to-end test suite again
    - Archives the current `suite.yml` as `releases/suite-x.y.z.yml` where `x.y.z` is the
       new suite release version.
    - Manual publishing of the release will make it public for consumers of this repo.
