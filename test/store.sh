#!/usr/bin/env bash

function b64_decode() {
  base64 -d "$@"
}
if base64 -D <(echo "Cg==") &> /dev/null || false; then
    function b64_decode() {
        base64 -D "$@"
    }
fi

function init_store() {
  rm -rf ./store.json
  echo "{}" > ./store.json
}

function get_val() {
  jq -r -e ".[\"${1}\"]" ./store.json | b64_decode
}

function set_val() {
  jq --arg newval "$(echo "${2}" | base64)" '.["'${1}'"] = $newval' ./store.json > ./tmp.json
  rm ./store.json
  mv ./tmp.json ./store.json
}
