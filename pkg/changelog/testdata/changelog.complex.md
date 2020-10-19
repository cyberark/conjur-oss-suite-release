# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Improved flows and rules around user creation
- Kubernetes authenticator now returns 403 on unpermitted hosts instead of a 401

## [1.4.6] - 2020-01-21

### Changed
- K8s hosts' resource restrictions is extracted from annotations or id. If it is
  defined in annotations it will taken from there and if not, it will be taken
  from the id.
- Another change ABC!@#$%

## [1.4.5] - 2019-12-22

### Added
- Added API endpoint to enable and disable authenticators. See
  [design/authenticator_whitelist_api.md](design/authenticator_whitelist_api.md)
  for details.

### Changed
- The k8s host id does not use the "{@account}:host:conjur/authn-k8s/#{@service_name}/apps"
  prefix and takes the full host-id from the CSR. We also handle backwards-compatibility and use
  the prefix in case of an older client.

## [1.4.4] - 2019-12-19

### Added
- Early validation of account existence during OIDC authentication
- Code coverage reporting and collection

### Changed
- Bumped `puma` from 3.12.0 to 3.12.2
- Bumped `rack` from 1.6.11 to 1.6.12
- Bumped `excon` from 0.62.0 to 0.71.0

### Fixed
- Fixed password rotation of blank password
- Fixed bug with multi-cert CA chains in Kubernetes service accounts
- Fixed build issues with creating namespaces with multiple values

### Removed
- Removed follower env configuration

## [1.4.3] - 2019-11-26

### Added
- Flattening of OSS container layers.

### Changed
- Upgraded Nokogiri to 1.10.5.
- Upgrade base image of OSS to `ubuntu:20.20`.
- Enablement work to get OSS container to work on OpenShift as-is.

## [1.4.2] - 2019-09-13

### Fixed
- An unset initContainer field in a deployment config pod spec will no
  longer cause the k8s authenticator to fail with `undefined method` ([#1182](https://github.com/cyberark/conjur/issues/1182)).

## [1.4.1] - 2019-06-24
### Fixed
- Make sure the authentication framework only caches Role lookups for the
  duration of a single request. Reusing stale lookups was leading to
  authentication failures.

## [1.4.0] - 2019-04-23
### Added
- Kubernetes authentication can now work externally from Kubernetes

### Changed
- Moved changelog validation up in CI pipeline

## 0.1.0 - 2017-12-04
### Added
- The first tagged version.

[Unreleased]: https://github.com/cyberark/conjur/compare/v1.4.6...HEAD
[1.4.6]: https://github.com/cyberark/conjur/compare/v1.4.5...v1.4.6
[1.4.5]: https://github.com/cyberark/conjur/compare/v1.4.4...v1.4.5
[1.4.4]: https://github.com/cyberark/conjur/compare/v1.4.3...v1.4.4
[1.4.3]: https://github.com/cyberark/conjur/compare/v1.4.2...v1.4.3
[1.4.2]: https://github.com/cyberark/conjur/compare/v1.4.1...v1.4.2
[1.4.1]: https://github.com/cyberark/conjur/compare/v1.4.0...v1.4.1
[1.4.0]: https://github.com/cyberark/conjur/compare/v0.1.0...v1.4.0
