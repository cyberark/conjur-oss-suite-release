#!/bin/bash

set -euo pipefail

# Sets additional required environment variables that aren't available in the
# secrets.yml file, and performs other preparatory steps

# Prepare Docker images
function prepareTestEnvironment() {
  # Pipe the Dockerfile into the command to avoid sending the whole
  # context to Docker
  docker build --rm --tag "gke-utils:latest" - < Dockerfile
}

function runDockerCommand() {
  docker run --rm \
    -i \
    -e GCLOUD_SERVICE_KEY="/tmp${GCLOUD_SERVICE_KEY}" \
    -e GCLOUD_CLUSTER_NAME \
    -e GCLOUD_ZONE \
    -e GCLOUD_PROJECT_NAME \
    -v "${GCLOUD_SERVICE_KEY}:/tmp${GCLOUD_SERVICE_KEY}" \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v "$PWD/..":/src \
    -w /src \
    "gke-utils:latest" \
    bash -exc "
      ./k8s-ci/platform_login > /dev/null
      $1
    "
}

function announce() {
  echo "++++++++++++++++++++++++++++++++++++++"
  echo ""
  echo "$@"
  echo ""
  echo "++++++++++++++++++++++++++++++++++++++"
}
