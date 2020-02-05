## https://github.com/cyberark/conjur-oss-helm-chart
### Changed
- Server ciphers have been upgraded to TLS1.2 levels.

## https://github.com/cyberark/conjur-api-go
* Converted to Golang 1.12
* Started using os.UserHomeDir() built-in instead of go-homedir module

## https://github.com/cyberark/conjur-api-java
### Added
- License updated to Apache v2 - PR [8](https://github.com/cyberark/conjur-api-java/pull/8)
### Changed
- Authn tokens now use the new Conjur 5 format - PR [21](https://github.com/cyberark/conjur-api-java/pull/21)
- Configuration change. When using environment variables, use `CONJUR_AUTHN_LOGIN` and `CONJUR_AUTHN_API_KEY` now
    instead of `CONJUR_CREDENTIALS` - https://github.com/cyberark/conjur-api-java/commit/60344308fc48cb5380c626e612b91e1e720c03fb

## https://github.com/cyberark/conjur-api-python3
### Added

- Added ability to delete policies [#23](https://github.com/cyberark/conjur-api-python3/issues/23)

## https://github.com/cyberark/conjur-api-ruby
* Updates URI path parameter escaping to consistently encode resource ids

## https://github.com/cyberark/conjur
### Changed
- K8s hosts' application identity is extracted from annotations or id. If it is
  defined in annotations it will taken from there and if not, it will be taken 
  from the id.

