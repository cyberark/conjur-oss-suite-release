#!/usr/bin/env bash

. ./store.sh

function exec_conjur_client() {
  local TEST_NAMESPACE CONJUR_CLIENT_POD_NAME
  TEST_NAMESPACE="$(store_get TEST_NAMESPACE)"
  CONJUR_CLIENT_POD_NAME="$(store_get CONJUR_CLIENT_POD_NAME)"

  kubectl --namespace "${TEST_NAMESPACE}" \
   exec -i "${CONJUR_CLIENT_POD_NAME}" \
    -- "$@"
}

function exec_conjur() {
  local TEST_NAMESPACE CONJUR_POD_NAME
  TEST_NAMESPACE="$(store_get TEST_NAMESPACE)"
  CONJUR_POD_NAME="$(store_get CONJUR_POD_NAME)"

  kubectl --namespace "${TEST_NAMESPACE}" \
   exec -i "${CONJUR_POD_NAME}" --container=conjur-oss \
    -- "$@"
}

function exec_app() {
  local TEST_NAMESPACE APP_POD_NAME
  TEST_NAMESPACE="$(store_get TEST_NAMESPACE)"
  APP_POD_NAME="$(store_get APP_POD_NAME)"

  kubectl --namespace "${TEST_NAMESPACE}" \
   exec -i "${APP_POD_NAME}" --container=app \
    -- "$@"
}
