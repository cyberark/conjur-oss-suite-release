#!/usr/bin/env bash

CURRENT_DIR="$(dirname "${BASH_SOURCE[0]}")"
if [[ -f "${CURRENT_DIR}/store-port" ]]; then
    export STORE_PORT="$(cat ${CURRENT_DIR}/store-port)"
fi

# store_init initialises a store and waits for it to come online.
# store_init also stores the port the store is listening in "${CURRENT_DIR}/store-port"
function store_init() {
  rm -rf "${CURRENT_DIR}/store-logs" "${CURRENT_DIR}/store-port"

  go build -o "${CURRENT_DIR}/store" "${CURRENT_DIR}/kv/cmd"
  "${CURRENT_DIR}/store" serve &> "${CURRENT_DIR}/store-logs" &

  sleep 1
  cat "${CURRENT_DIR}/store-logs" | \
   grep "Using port:" | \
    awk '{ print $5 }' | \
     xargs echo -n > "${CURRENT_DIR}/store-port"
  export STORE_PORT="$(cat ${CURRENT_DIR}/store-port)"
}

# store_destroy instructs the store, if one exists, to terminate.
function store_destroy() {
  "${CURRENT_DIR}/store" destroy
}

# store_get fetches the value, in the store, of the key provided as a command line argument.
# arg[1] = key
function store_get() {
  "${CURRENT_DIR}/store" get -k "${1}"
}

# store_set records into the store the key-value pair provided as command line arguments.
# arg[1] = key
# arg[2] = value
function store_set() {
  "${CURRENT_DIR}/store" set -k "${1}" -v "${2}"
}
