# conjur-oss-suite-release
Latest stable releases of the Conjur OSS suite

**THIS REPO IS UNDER CONSTRUCTION, AND THERE ARE NO SUITE RELEASES AVAILABLE YET**

_For more info on the planned work, please see our [design doc](https://github.com/cyberark/conjur/blob/master/design/oss_suite_release.md)
and monitor the [github issues](https://github.com/cyberark/conjur-oss-suite-release/issues)._

#### This repo's metrics:
![Tests](https://github.com/cyberark/conjur-oss-suite-release/workflows/Tests/badge.svg) [![Test Coverage](https://api.codeclimate.com/v1/badges/31060f348b29c7f5d02b/test_coverage)](https://codeclimate.com/repos/5e2b43bf92af05714c00b172/test_coverage) [![Maintainability](https://api.codeclimate.com/v1/badges/31060f348b29c7f5d02b/maintainability)](https://codeclimate.com/repos/5e2b43bf92af05714c00b172/maintainability)

## Usage

- Edit the file `repositories.yml` as needed
- Run the CHANGELOG generator:
```
./parse-changelogs
```
- Resulting changelog will be placed in `CHANGELOG.md`

## Advanced usage

The CLI accepts the following arguments/parameters:
```
  -f string
        Repository YAML file to parse (default "repositories.yml")
  -o string
        Output filename (default "CHANGELOG.md")
  -p string
        GitHub API token
  -t string
        Output type. Only accepts 'changelog' and 'release'. (default "changelog")
  -v string
        Version to embed in the changelog (default "Unreleased")
```

## Development
We welcome contributions of all kinds. For instructions on how to get started and
descriptions of our development workflows, please see our [contributing guide](CONTRIBUTING.md).

## License

This repository is licensed under Apache License 2.0 - see [`LICENSE`](LICENSE) for more details.
