# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/cyberark/conjur-oss-suite-release/compare/v1.11.1+suite.2...HEAD
[v1.11.1+suite.2]: https://github.com/cyberark/conjur-oss-helm-chart/compare/v1.11.1+suite.1...v1.11.1+suite.2
[v1.11.1+suite.1]: https://github.com/cyberark/conjur-oss-helm-chart/compare/v1.10.0+suite.1...v1.11.1+suite.1
