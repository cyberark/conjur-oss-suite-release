{{- $conjurVersion := .ComponentReleaseVersion "cyberark/conjur" -}}
{{- $helmChartVersion := .ComponentReleaseVersion "cyberark/conjur-oss-helm-chart" -}}
Installing the Suite Release Version of Conjur requires setting the container image tag. Below are more specific instructions depending on environment.

+ **Docker or docker-compose**

  Set the container image tag to `cyberark/conjur:{{$conjurVersion}}`.
  For example, make the following update to the conjur service in the [quickstart docker-compose.yml](https://github.com/cyberark/conjur-quickstart/blob/master/docker-compose.yml)
  ```
  image: cyberark/conjur:{{$conjurVersion}}
  ```

+ [**Cloud Formation templates for AWS**](https://github.com/cyberark/conjur-aws)

  Set the environment variable CONJUR_VERSION before building the AMI:
  ```
  export CONJUR_VERSION="{{$conjurVersion}}"
  ./build-ami.sh
  ```
{{- if $helmChartVersion }}

+ [**Conjur OSS Helm chart**](https://github.com/cyberark/conjur-oss-helm-chart)

  Update the `image.tag` value and use the appropriate release of the helm chart:
  ```
  helm install ... \
    --set image.tag="{{$conjurVersion}}" \
    ...
    https://github.com/cyberark/conjur-oss-helm-chart/releases/download/v{{$helmChartVersion}}/conjur-oss-{{$helmChartVersion}}.tgz
  ```
{{- end -}}
