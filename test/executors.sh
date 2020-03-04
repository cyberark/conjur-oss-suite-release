#!/usr/bin/env bash

. ./store.sh

function exec_conjur_client() {
  kubectl --namespace "$(store_get TEST_NAMESPACE)" \
   exec -i "$(store_get CONJUR_CLIENT_POD_NAME)" \
    -- "$@"
}

function exec_conjur() {
  kubectl --namespace "$(store_get TEST_NAMESPACE)" \
   exec -i "$(store_get CONJUR_POD_NAME)" --container=conjur-oss \
    -- "$@"
}

function exec_app() {
  kubectl --namespace "$(store_get TEST_NAMESPACE)" \
   exec -i "$(store_get APP_POD_NAME)" --container=app \
    -- "$@"
}
