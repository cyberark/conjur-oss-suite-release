# conjur-oss-suite-release

Building the official releases of the [Conjur OSS suite](https://cyberark.github.io/conjur)

**THIS REPO IS UNDER CONSTRUCTION, AND THERE ARE NO SUITE RELEASES AVAILABLE YET**

_For more info on the planned work, please see our [design doc](https://github.com/cyberark/conjur/blob/master/design/oss_suite_release.md)
and monitor the [github issues](https://github.com/cyberark/conjur-oss-suite-release/issues)._

Once we have a first release with release notes, we'll be adding a link to the
official release notes in the Conjur documentation. At that time, we will recommend:
> You've found the repository that we use for building the official Conjur OSS suite
> releases. To keep track of upcoming releases and plans, please see the
> [github issues](https://github.com/cyberark/conjur-oss-suite-release/issues).
> To view the latest release and relevant documentation, please see _the official_
> _release notes_.

#### This repo's metrics:
![Tests](https://github.com/cyberark/conjur-oss-suite-release/workflows/Tests/badge.svg)
[![Test Coverage](https://api.codeclimate.com/v1/badges/31060f348b29c7f5d02b/test_coverage)](https://codeclimate.com/repos/5e2b43bf92af05714c00b172/test_coverage)
[![Maintainability](https://api.codeclimate.com/v1/badges/31060f348b29c7f5d02b/maintainability)](https://codeclimate.com/repos/5e2b43bf92af05714c00b172/maintainability)

## What is the Conjur OSS Suite?

[CyberArk Conjur](https://github.com/cyberark/conjur) is a RESTful web service that
can be used to securely authenticate, control and audit non-human access across
tools, applications, containers and cloud environments via robust secrets management.

To ensure that Conjur is easy to use no matter your application's language, the
tools you use to build or deploy it, or the platform you deploy it to, we've built
an extensive set of additional plugins, libraries, and software that extend Conjur
to work natively with a variety of tools and in a wide range of environments.

The **Conjur OSS Suite** aggregates this set of tools to provide a _single place_
where you can:
- Find out about all of the open source tools and integrations that work with Conjur
- Learn about new features in existing components
- Learn about new integrations or tools that work with Conjur
- Find out which version of Conjur to use, and which corresponding versions of
  the suite components are compatible with it

If you're using Conjur OSS, we **strongly recommend** that you reference the latest
suite release (link TBA) to determine which version of Conjur to use, and which corresponding
versions of the suite are compatible with it.

### A note on versioning
The core of each Conjur OSS Suite release is the [Conjur server](https://github.com/cyberark/conjur).
In light of this, we decided to tie our versioning of the suite to the _Conjur
server version_.

This means that if Conjur is included in the suite at version `1.x.y`, then the
suite release will be versioned as `1.x.y+suite.z` (following the [semver standard](https://semver.org/)
for build metadata). Most of the time, there will be one suite release per Conjur
version - so that the suite will be versioned as `1.x.y+suite.1`. On occasions
where there is an update to a suite component or a valuable _new_ suite component
that becomes available between Conjur releases, we _may_ choose to release a new
version of the suite that corresponds to the same Conjur version as in the previous
suite release, and increment the last digit of the suite version (e.g. `1.x.y+suite.2`).

**Importantly**, if you are using Conjur version `1.x.y`, then you can use any
suite release with version starting with `1.x.y` with confidence.

## Development
We welcome contributions of all kinds to this project. For instructions on how to
get started, instructions for using the suite release tooling in this project, and
descriptions of our development workflows - please see our [contributing guide](CONTRIBUTING.md).

## License

This repository is licensed under Apache License 2.0 - see [`LICENSE`](LICENSE)
for more details.
