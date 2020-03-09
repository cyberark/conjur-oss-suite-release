# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased] - 2020-02-19

### Added
- `cyberark/conjur@1.4.4`: Early validation of account existence during OIDC authentication
- `cyberark/conjur@1.4.4`: Code coverage reporting and collection
- `cyberark/conjur-api-python3@0.0.5`: Added ability to delete policies [#23](https://github.com/cyberark/conjur-api-python3/issues/23)

### Changed
- `cyberark/conjur@1.3.6`: Reduced IAM authentication logging
- `cyberark/conjur@1.3.6`: Refactored authentication strategies
- `cyberark/conjur@1.4.4`: Bumped puma from 3.12.0 to 3.12.2
- `cyberark/conjur@1.4.4`: Bumped rack from 1.6.11 to 1.6.12
- `cyberark/conjur@1.4.4`: Bumped excon from 0.62.0 to 0.71.0
- `cyberark/conjur@1.4.6`: K8s hosts' application identity is extracted from annotations or id. If it is
defined in annotations it will taken from there and if not, it will be taken
from the id.
- `cyberark/conjur-oss-helm-chart@1.3.7`: Server ciphers have been upgraded to TLS1.2 levels.

### Fixed
- `cyberark/conjur@1.4.4`: Fixed password rotation of blank password
- `cyberark/conjur@1.4.4`: Fixed bug with multi-cert CA chains in Kubernetes service accounts
- `cyberark/conjur@1.4.4`: Fixed build issues with creating namespaces with multiple values

### Removed
- `cyberark/conjur@1.3.6`: Removed OIDC APIs public access
- `cyberark/conjur@1.4.4`: Removed follower env configuration


