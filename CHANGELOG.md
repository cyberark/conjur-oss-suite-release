# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v1.20.0+suite.1] - 2023-02-06

### Changed
- Updated gopkg.in/yaml.v3 indirect dependency and k8s-ci Dockerfile base image
  [cyberark/conjur-oss-suite-release#260](https://github.com/cyberark/conjur-oss-suite-release/pull/260)
- Replace deprecated Ruby CLI with Go CLI
  [cyberark/conjur-oss-suite-release#259](https://github.com/cyberark/conjur-oss-suite-release/pull/259)
- Upgraded release-testing go.mod to use github.com/stretchr/testify v1.8.1
  [cyberark/conjur-oss-suite-release#258](https://github.com/cyberark/conjur-oss-suite-release/pull/258)

## [v1.19.0+suite.1] - 2022-11-30

## [v1.18.4+suite.1] - 2022-09-16

## [v1.18.0+suite.1] - 2022-08-23

## [v1.17.6+suite.1] - 2022-05-17

### Changed
- Updated go dependencies to latest versions (github.com/gomarkdown/markdown
  -> v0.0.0-20220607163217-45f7c050e2d1, github.com/stretchr/testify -> v1.7.2,
  gopkg.in/yaml.v3 -> v3.0.1)
  [cyberark/conjur-oss-suite-release#248](https://github.com/cyberark/conjur-oss-suite-release/pull/248)

### Security
- Updated gopkg.in/yaml.v2 to resolve CVE-2019-11254
  [cyberark/conjur-oss-suite-release#245](https://github.com/cyberark/conjur-oss-suite-release/pull/245)
  [cyberark/conjur-oss-suite-release#246](https://github.com/cyberark/conjur-oss-suite-release/pull/246)

## [v1.15.0+suite.1] - 2022-01-24

### Changed
- Updated Go version to 1.17. [cyberark/conjur-oss-suite-release#241](https://github.com/cyberark/conjur-oss-suite-release/pull/241)

## [v1.14.1+suite.1] - 2021-11-12

## [v1.13.1+suite.1] - 2021-09-14

## [v1.13.0+suite.1] - 2021-08-13

## [v1.11.7+suite.1] - 2021-07-07

## [v1.11.6+suite.1] - 2021-05-11

### Added
- The Conjur OpenAPI specification is now part of the Conjur Open Source Suite.

## [v1.11.5+suite.1] - 2021-04-14

### Changed
- The Python SDK is incremented to v7.0.1 to announce the release of the v7 CLI.

## [v1.11.3+suite.1] - 2021-03-09

### Fixed
- The draft release action is updated to use valid logic for determining the
  suite version so that the draft release notes will include the correct
  interpolated version.
  [cyberark/conjur-oss-suite-release#212](https://github.com/cyberark/conjur-oss-suite-release/issues/212)

## [v1.11.2+suite.1] - 2021-02-10

### Added
- The [Conjur Service Broker](https://github.com/cyberark/conjur-service-broker)
  and [Conjur Buildpack](https://github.com/cyberark/cloudfoundry-conjur-buildpack)
  have been added to the suite.
  [PR cyberark/conjur-oss-suite-release#207](https://github.com/cyberark/conjur-oss-suite-release/pull/207)

### Fixed
- The version package now uses the build metadata of the form `{string}.{number}`
  in determining the highest release version in a directory. Previously, if
  two files had the same version and differed only in the increment in the
  build metadata, whichever file was processed last would be marked the latest
  release. With this change the file with the highest build metadata increment
  will be considered as the latest release instead.
  [cyberark/conjur-oss-suite-release#208](https://github.com/cyberark/conjur-oss-suite-release/issues/208)

## [v1.11.1+suite.2] - 2021-01-06

### Added
- The [Conjur Ansible collection](https://github.com/cyberark/ansible-conjur-collection)
  has been added to the suite.
  [PR cyberark/conjur-oss-suite-release#204](https://github.com/cyberark/conjur-oss-suite-release/pull/204)

## [v1.11.1+suite.1] - 2020-12-04

### Changed
- Updated Go version to 1.15. [cyberark/conjur-oss-suite-release#196](https://github.com/cyberark/conjur-oss-suite-release/issues/196)
- Updates links in HTML output to include `target="_blank"` so that they will
  open in a new window, unless the link points to a CyberArk docs site.
  [cyberark/conjur-oss-suite-release#199](https://github.com/cyberark/conjur-oss-suite-release/issues/199)

[Unreleased]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.19.0+suite.1...HEAD
[v1.19.0+suite.1]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.18.4+suite.1...v1.19.0+suite.1
[v1.18.4+suite.1]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.18.0+suite.1...v1.18.4+suite.1
[v1.18.0+suite.1]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.17.6+suite.1...v1.18.0+suite.1
[v1.17.6+suite.1]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.15.0+suite.1...v1.17.6+suite.1
[v1.15.0+suite.1]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.14.1+suite.1...v1.15.0+suite.1
[v1.14.1+suite.1]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.13.1+suite.1...v1.14.1+suite.1
[v1.13.1+suite.1]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.13.0+suite.1...v1.13.1+suite.1
[v1.13.0+suite.1]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.11.7+suite.1...v1.13.0+suite.1
[v1.11.7+suite.1]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.11.6+suite.1...v1.11.7+suite.1
[v1.11.6+suite.1]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.11.5+suite.1...v1.11.6+suite.1
[v1.11.5+suite.1]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.11.3+suite.1...v1.11.5+suite.1
[v1.11.3+suite.1]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.11.2+suite.1...v1.11.3+suite.1
[v1.11.2+suite.1]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.11.1+suite.2...v1.11.2+suite.1
[v1.11.1+suite.2]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.11.1+suite.1...v1.11.1+suite.2
[v1.11.1+suite.1]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.10.0+suite.1...v1.11.1+suite.1
