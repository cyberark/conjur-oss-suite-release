#!/usr/bin/env bash
set -euo pipefail

. ./store.sh
. ./executors.sh

function deploy_pg {
  local TEST_NAMESPACE PG_RELEASE_NAME DB_PASSWORD
  TEST_NAMESPACE="$(store_get TEST_NAMESPACE)"
  PG_RELEASE_NAME="$(store_get PG_RELEASE_NAME)"
  DB_PASSWORD="$(store_get DB_PASSWORD)"

  helm install --namespace "${TEST_NAMESPACE}" "${PG_RELEASE_NAME}" \
   --set postgresqlPassword="${DB_PASSWORD}" \
   --set persistence.enabled="false" \
   stable/postgresql
}

function deploy_conjur {
  local TEST_NAMESPACE
  local CONJUR_RELEASE_NAME
  local DATA_KEY
  local CONJUR_ACCOUNT
  local CONJUR_OSS_RELEASE_VERSION
  local AUTHENTICATOR_ID
  local DB_PASSWORD
  local PG_RELEASE_NAME
  local HELM_CHART_RELEASE_VERSION

  TEST_NAMESPACE="$(store_get TEST_NAMESPACE)"
  CONJUR_RELEASE_NAME="$(store_get CONJUR_RELEASE_NAME)"
  DATA_KEY="$(store_get DATA_KEY)"
  CONJUR_ACCOUNT="$(store_get CONJUR_ACCOUNT)"
  CONJUR_OSS_RELEASE_VERSION="$(store_get CONJUR_OSS_RELEASE_VERSION)"
  AUTHENTICATOR_ID="$(store_get AUTHENTICATOR_ID)"
  DB_PASSWORD="$(store_get DB_PASSWORD)"
  PG_RELEASE_NAME="$(store_get PG_RELEASE_NAME)"
  HELM_CHART_RELEASE_VERSION="$(store_get HELM_CHART_RELEASE_VERSION)"

  helm install --namespace "${TEST_NAMESPACE}" "${CONJUR_RELEASE_NAME}" \
    --set dataKey="${DATA_KEY}" \
    --set account="${CONJUR_ACCOUNT}" \
    --set image.tag="${CONJUR_OSS_RELEASE_VERSION}" \
    --set image.pullPolicy="Always" \
    --set authenticators="authn-k8s/${AUTHENTICATOR_ID}\,authn" \
    --set databaseUrl="postgres://postgres:${DB_PASSWORD}@${PG_RELEASE_NAME}-postgresql.${TEST_NAMESPACE}.svc.cluster.local/postgres" \
    "https://github.com/cyberark/conjur-oss-helm-chart/releases/download/v${HELM_CHART_RELEASE_VERSION}/conjur-oss-${HELM_CHART_RELEASE_VERSION}.tgz"
}

function register_conjur_pod() {
  local TEST_NAMESPACE CONJUR_RELEASE_NAME
  TEST_NAMESPACE="$(store_get TEST_NAMESPACE)"
  CONJUR_RELEASE_NAME="$(store_get CONJUR_RELEASE_NAME)"

  local CONJUR_POD_NAME
  CONJUR_POD_NAME=$(kubectl --namespace "${TEST_NAMESPACE}" get pods \
                   -l "app=conjur-oss,release=${CONJUR_RELEASE_NAME}" \
                   -o jsonpath="{.items[0].metadata.name}")

  store_set CONJUR_POD_NAME "${CONJUR_POD_NAME}"
}

function register_conjur_client_pod() {
  local TEST_NAMESPACE CONJUR_URL CONJUR_ACCOUNT CONJUR_ADMIN_API_KEY
  TEST_NAMESPACE="$(store_get TEST_NAMESPACE)"
  CONJUR_URL="$(store_get CONJUR_URL)"
  CONJUR_ACCOUNT="$(store_get CONJUR_ACCOUNT)"
  CONJUR_ADMIN_API_KEY="$(store_get CONJUR_ADMIN_API_KEY)"

  local CONJUR_CLIENT_POD_NAME="conjur-client-pod"

  # start the CLI pod
  kubectl --namespace "${TEST_NAMESPACE}" run "${CONJUR_CLIENT_POD_NAME}" \
   --restart='Never' \
   --image cyberark/conjur-cli:5 \
   --command -- sleep infinity

  # wait for CLI pod to be ready
  kubectl --namespace "${TEST_NAMESPACE}" \
   wait --for=condition=ready "pod/${CONJUR_CLIENT_POD_NAME}" --timeout 150s

  # login to conjur
  kubectl --namespace "${TEST_NAMESPACE}" \
    exec -i "${CONJUR_CLIENT_POD_NAME}" -- \
     bash -xce "
yes yes | conjur init -u '${CONJUR_URL}' -a '${CONJUR_ACCOUNT}'

# API key here is the key that creation of the account provided you in step #2
conjur authn login -u admin -p '${CONJUR_ADMIN_API_KEY}'

# Check that you are identified as the admin user
conjur authn whoami
"

  store_set CONJUR_CLIENT_POD_NAME "${CONJUR_CLIENT_POD_NAME}"
}

function register_conjur_admin_key() {
  local admin_user_id
  admin_user_id="$(store_get CONJUR_ACCOUNT):user:admin"

  local CONJUR_ADMIN_API_KEY
  CONJUR_ADMIN_API_KEY=$(exec_conjur \
   conjurctl role retrieve-key "${admin_user_id}")

  store_set CONJUR_ADMIN_API_KEY "${CONJUR_ADMIN_API_KEY}"
}

function setup_conjur() {
  local CONJUR_ACCOUNT
  CONJUR_ACCOUNT="$(store_get CONJUR_ACCOUNT)"

  exec_conjur conjurctl account create "${CONJUR_ACCOUNT}"
}

function run_policy() {
  local CONJUR_ACCOUNT AUTHENTICATOR_ID
  CONJUR_ACCOUNT="$(store_get CONJUR_ACCOUNT)"
  AUTHENTICATOR_ID="$(store_get AUTHENTICATOR_ID)"

  local policy
  policy="$(./policy.yml.sh)"

  store_set "policy" "${policy}"
  echo -n "${policy}" | exec_conjur_client conjur policy load --replace root /dev/stdin

  exec_conjur_client bash -xce "
# Generate OpenSSL private key
openssl genrsa -out ca.key 2048

echo -n '
[ req ]
distinguished_name = dn
x509_extensions = v3_ca
[ dn ]
[ v3_ca ]
basicConstraints = critical,CA:TRUE
subjectKeyIdentifier   = hash
authorityKeyIdentifier = keyid:always,issuer:always
' > ca.config

# Generate root CA certificate
openssl req -x509 -new -nodes -key ca.key -sha1 -days 3650 -set_serial 0x0 -out ca.cert \
  -subj '/CN=conjur.authn-k8s.${AUTHENTICATOR_ID}/OU=Conjur Kubernetes CA/O=${CONJUR_ACCOUNT}' \
  -config ca.config

# Verify cert
openssl x509 -in ca.cert -text -noout

# Set cert values
conjur variable values add conjur/authn-k8s/${AUTHENTICATOR_ID}/ca/key < ca.key
conjur variable values add conjur/authn-k8s/${AUTHENTICATOR_ID}/ca/cert < ca.cert
"
}

function populate_variables() {
    exec_conjur_client conjur variable values add test-app-secrets/username 'meow meow meow'
}

function main() {
  local TEST_NAMESPACE APP_SERVICE_ACCOUNT
  TEST_NAMESPACE="$(store_get TEST_NAMESPACE)"
  APP_SERVICE_ACCOUNT="$(store_get APP_SERVICE_ACCOUNT)"

  kubectl create namespace "${TEST_NAMESPACE}"

  echo "Deploy Postgres."
  deploy_pg

  echo "Deploy Conjur."
  deploy_conjur

  echo "Wait for Conjur."
  register_conjur_pod

  local CONJUR_POD_NAME
  CONJUR_POD_NAME="$(store_get CONJUR_POD_NAME)"
  kubectl --namespace "${TEST_NAMESPACE}" \
   wait --for=condition=ready "pod/${CONJUR_POD_NAME}" --timeout 150s
  exec_conjur conjurctl wait


  echo "Setup Conjur."
  setup_conjur
  register_conjur_admin_key

  echo "Run policy on Conjur."
  register_conjur_client_pod
  run_policy
  populate_variables

  echo "Deploy app."
  kubectl --namespace "${TEST_NAMESPACE}" \
    create sa "${APP_SERVICE_ACCOUNT}"

  local app_deployment
  app_deployment="$(./app_secretless_deployment.yml.sh)"
  store_set "app_deployment" "${app_deployment}"
  echo -n "${app_deployment}" | kubectl --namespace "${TEST_NAMESPACE}" apply -f -

  local APP_POD_NAME
  APP_POD_NAME=$(kubectl --namespace "${TEST_NAMESPACE}" get pods \
                   -l "app=test-app" \
                   -o jsonpath="{.items[0].metadata.name}")
  store_set APP_POD_NAME "${APP_POD_NAME}"
  kubectl --namespace "${TEST_NAMESPACE}" wait \
   --for=condition=ready "pod/${APP_POD_NAME}" --timeout 150s
}

main
