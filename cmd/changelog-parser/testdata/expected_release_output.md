# Release Notes
All notable changes to this project will be documented in this file.

## [Unreleased] - 2020-02-19

## Table of Contents

- [Components](#components)
- [Installation Instructions for the Suite Release Version of Conjur](#installation-instructions-for-the-suite-release-version-of-conjur)
- [Upgrade Instructions](#upgrade-instructions)
- [Changes](#changes)

## Components

These are the components that combine to create this Conjur OSS Suite release and links
to their releases:

- **[cyberark/conjur v1.4.6](https://github.com/cyberark/conjur/releases/tag/v1.4.6)** (2020-01-21) [![Certification Level](https://img.shields.io/badge/Certification%20Level-Trusted-Blue)](https://github.com/cyberark/conjur)
- **[cyberark/conjur-oss-helm-chart v1.3.7](https://github.com/cyberark/conjur-oss-helm-chart/releases/tag/v1.3.7)** (2019-01-31) [![Certification Level](https://img.shields.io/badge/Certification%20Level-Certified-Green)](https://github.com/cyberark/conjur-oss-helm-chart)
- **[cyberark/conjur-api-python3 v0.0.5](https://github.com/cyberark/conjur-api-python3/releases/tag/v0.0.5)** (2019-12-06) [![Certification Level](https://img.shields.io/badge/Certification%20Level-Community-Yellow)](https://github.com/cyberark/conjur-api-python3)

## Installation Instructions for the Suite Release Version of Conjur

Installing the Suite Release Version of Conjur requires setting the container image tag. Below are more specific instructions depending on environment.

+ **Docker or docker-compose**

  Set the container image tag to `cyberark/conjur:1.4.6`.
  For example, make the following update to the conjur service in the [quickstart docker-compose.yml](https://github.com/cyberark/conjur-quickstart/blob/master/docker-compose.yml)
  ```
  image: cyberark/conjur:1.4.6
  ```

+ [**Cloud Formation templates for AWS**](https://github.com/cyberark/conjur-aws)

  Set the environment variable CONJUR_VERSION before building the AMI:
  ```
  export CONJUR_VERSION="1.4.6"
  ./build-ami.sh
  ```

+ [**Conjur OSS Helm chart**](https://github.com/cyberark/conjur-oss-helm-chart)

  Update the `image.tag` value and use the appropriate release of the helm chart:
  ```
  helm install ... \
    --set image.tag="1.4.6" \
    ...
    https://github.com/cyberark/conjur-oss-helm-chart/releases/download/v1.3.7/conjur-oss-1.3.7.tgz
  ```

## Upgrade Instructions

Upgrade instructions are available for the following components:

- [cyberark/conjur](https://docs.cyberark.com/Product-Doc/OnlineHelp/AAM-DAP/Latest/en/Content/Deployment/Upgrade/upgrade-intro.htm)

## Changes

The following are changes to the constituent components since the last Conjur
OSS Suite release:

### [cyberark/conjur v1.3.6](https://github.com/cyberark/conjur/releases/tag/v1.3.6) (2019-02-19)

#### Changed
- Reduced IAM authentication logging
- Refactored authentication strategies

#### Removed
- Removed OIDC APIs public access

### [cyberark/conjur v1.4.4](https://github.com/cyberark/conjur/releases/tag/v1.4.4) (2019-12-19)

#### Added
- Early validation of account existence during OIDC authentication
- Code coverage reporting and collection

#### Changed
- Bumped puma from 3.12.0 to 3.12.2
- Bumped rack from 1.6.11 to 1.6.12
- Bumped excon from 0.62.0 to 0.71.0

#### Fixed
- Fixed password rotation of blank password
- Fixed bug with multi-cert CA chains in Kubernetes service accounts
- Fixed build issues with creating namespaces with multiple values

#### Removed
- Removed follower env configuration

### [cyberark/conjur v1.4.6](https://github.com/cyberark/conjur/releases/tag/v1.4.6) (2020-01-21)

#### Changed
- K8s hosts' application identity is extracted from annotations or id. If it is
defined in annotations it will taken from there and if not, it will be taken
from the id.

### [cyberark/conjur-oss-helm-chart v1.3.7](https://github.com/cyberark/conjur-oss-helm-chart/releases/tag/v1.3.7) (2019-01-31)

#### Changed
- Server ciphers have been upgraded to TLS1.2 levels.

### [cyberark/conjur-api-python3 v0.0.5](https://github.com/cyberark/conjur-api-python3/releases/tag/v0.0.5) (2019-12-06)

#### Added
- Added ability to delete policies [#23](https://github.com/cyberark/conjur-api-python3/issues/23)

