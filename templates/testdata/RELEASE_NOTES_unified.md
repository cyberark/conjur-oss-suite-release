# Release Notes
All notable changes to this project will be documented in this file.

## [11.22.33] - 2020-02-19

## Table of Contents

- [Components](#components)
- [Installation Instructions for the Suite Release Version of Conjur](#installation-instructions-for-the-suite-release-version-of-conjur)
- [Upgrade Instructions](#upgrade-instructions)
- [Changes](#changes)

## Components

These are the components that combine to create this Conjur OSS Suite release and links
to their releases:
### Conjur Core

- **[cyberark/conjur v1.4.4](https://github.com/cyberark/conjur/releases/tag/v1.4.4)** (2020-01-03) [![Certification Level](https://img.shields.io/badge/Certification%20Level-Trusted-007BFF)](https://github.com/cyberark/conjur)
### Secrets Delivery

- **[cyberark/secretless-broker v1.4.2](https://github.com/cyberark/secretless-broker/releases/tag/v1.4.2)** (2020-01-08) [![Certification Level](https://img.shields.io/badge/Certification%20Level-Certified-6C757D)](https://github.com/cyberark/secretless-broker)

## Installation Instructions for the Suite Release Version of Conjur

Installing the Suite Release Version of Conjur requires setting the container image tag. Below are more specific instructions depending on environment.

+ **Docker or docker-compose**

  Set the container image tag to `cyberark/conjur:1.4.4`.
  For example, make the following update to the conjur service in the [quickstart docker-compose.yml](https://github.com/cyberark/conjur-quickstart/blob/master/docker-compose.yml)
  ```
  image: cyberark/conjur:1.4.4
  ```

+ [**Cloud Formation templates for AWS**](https://github.com/cyberark/conjur-aws)

  Set the environment variable CONJUR_VERSION before building the AMI:
  ```
  export CONJUR_VERSION="1.4.4"
  ./build-ami.sh
  ```

## Upgrade Instructions

Upgrade instructions are available for the following components:

- [cyberark/conjur](https://conjur_upgrade_url)

## Changes
The following are changes to the constituent components since the last Conjur
OSS Suite release:
- [cyberark/conjur](#cyberarkconjur)
- [cyberark/secretless-broker](#cyberarksecretless-broker)

### cyberark/conjur

#### [v1.3.6](https://github.com/cyberark/conjur/releases/tag/v1.3.6) (2020-02-01)
* **Changed**
    - 136Change
    - 136Change2
* **Removed**
    - 136Removal
#### [v1.4.4](https://github.com/cyberark/conjur/releases/tag/v1.4.4) (2020-01-03)
* **Added**
    - 144Addition
    - 144Addition2
* **Changed**
    - 144Change
    - 144Change2
* **Fixed**
    - 144Fix
### cyberark/secretless-broker

#### [v1.4.2](https://github.com/cyberark/secretless-broker/releases/tag/v1.4.2) (2020-01-08)
* **Added**
    - Broker142Addition
* **Changed**
    - Broker142Change
* **Removed**
    - Broker142Removal
