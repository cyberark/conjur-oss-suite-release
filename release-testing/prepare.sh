#!/usr/bin/env bash
set -euo pipefail

. ./store.sh
. ./executors.sh

function deploy_pg {
  local DB_PASSWORD PG_RELEASE_NAME TEST_NAMESPACE
  DB_PASSWORD="$(store_get DB_PASSWORD)"
  PG_RELEASE_NAME="$(store_get PG_RELEASE_NAME)"
  TEST_NAMESPACE="$(store_get TEST_NAMESPACE)"

  helm install --namespace "${TEST_NAMESPACE}" "${PG_RELEASE_NAME}" \
    --set postgresqlPassword="${DB_PASSWORD}" \
    --set persistence.enabled="false" \
    bitnami/postgresql
}

function deploy_conjur {
  local AUTHENTICATOR_ID
  local CONJUR_ACCOUNT
  local CONJUR_OSS_RELEASE_VERSION
  local CONJUR_RELEASE_NAME
  local DATA_KEY
  local DB_PASSWORD
  local HELM_CHART_RELEASE_VERSION
  local PG_RELEASE_NAME
  local TEST_NAMESPACE

  AUTHENTICATOR_ID="$(store_get AUTHENTICATOR_ID)"
  CONJUR_ACCOUNT="$(store_get CONJUR_ACCOUNT)"
  CONJUR_OSS_RELEASE_VERSION="$(store_get CONJUR_OSS_RELEASE_VERSION)"
  CONJUR_RELEASE_NAME="$(store_get CONJUR_RELEASE_NAME)"
  DATA_KEY="$(store_get DATA_KEY)"
  DB_PASSWORD="$(store_get DB_PASSWORD)"
  HELM_CHART_RELEASE_VERSION="$(store_get HELM_CHART_RELEASE_VERSION)"
  PG_RELEASE_NAME="$(store_get PG_RELEASE_NAME)"
  TEST_NAMESPACE="$(store_get TEST_NAMESPACE)"

  helm install --namespace "${TEST_NAMESPACE}" "${CONJUR_RELEASE_NAME}" \
    --set account.name="${CONJUR_ACCOUNT}" \
    --set authenticators="authn-k8s/${AUTHENTICATOR_ID}\,authn" \
    --set databaseUrl="postgres://postgres:${DB_PASSWORD}@${PG_RELEASE_NAME}-postgresql.${TEST_NAMESPACE}.svc.cluster.local/postgres" \
    --set dataKey="${DATA_KEY}" \
    --set image.pullPolicy="Always" \
    --set image.tag="${CONJUR_OSS_RELEASE_VERSION}" \
    "https://github.com/cyberark/conjur-oss-helm-chart/releases/download/v${HELM_CHART_RELEASE_VERSION}/conjur-oss-${HELM_CHART_RELEASE_VERSION}.tgz"
}

function register_conjur_pod() {
  local CONJUR_RELEASE_NAME POD_LABEL_SELECTOR TEST_NAMESPACE
  CONJUR_RELEASE_NAME="$(store_get CONJUR_RELEASE_NAME)"
  POD_LABEL_SELECTOR="app=conjur-oss,release=${CONJUR_RELEASE_NAME}"
  TEST_NAMESPACE="$(store_get TEST_NAMESPACE)"

  # Wait for deployment
  kubectl --namespace "${TEST_NAMESPACE}" \
    rollout status deployment "${CONJUR_RELEASE_NAME}" --watch --timeout=150s

  # Wait for pod
  kubectl --namespace "${TEST_NAMESPACE}" \
    wait pod --for=condition=ready -l "${POD_LABEL_SELECTOR}" --timeout 150s

  local CONJUR_POD_NAME
  CONJUR_POD_NAME=$(kubectl --namespace "${TEST_NAMESPACE}" get pods \
                      -l "${POD_LABEL_SELECTOR}" \
                      -o jsonpath="{.items[0].metadata.name}")
  store_set CONJUR_POD_NAME "${CONJUR_POD_NAME}"
}

function register_conjur_client_pod() {
  local TEST_NAMESPACE CONJUR_ACCOUNT CONJUR_ADMIN_API_KEY CONJUR_URL
  TEST_NAMESPACE="$(store_get TEST_NAMESPACE)"
  CONJUR_ACCOUNT="$(store_get CONJUR_ACCOUNT)"
  CONJUR_ADMIN_API_KEY="$(store_get CONJUR_ADMIN_API_KEY)"
  CONJUR_URL="$(store_get CONJUR_URL)"

  local CONJUR_CLIENT_POD_NAME="conjur-client-pod"

  # Start the CLI pod
  kubectl --namespace "${TEST_NAMESPACE}" run --generator=run-pod/v1 \
    "${CONJUR_CLIENT_POD_NAME}" \
    --restart='Never' \
    --image cyberark/conjur-cli:5 \
    --command -- sleep infinity

  # Wait for CLI pod to be ready
  kubectl --namespace "${TEST_NAMESPACE}" \
    wait --for=condition=ready "pod/${CONJUR_CLIENT_POD_NAME}" --timeout 150s

  # Login to conjur
  kubectl --namespace "${TEST_NAMESPACE}" \
    exec -i "${CONJUR_CLIENT_POD_NAME}" -- \
      bash -exc "
yes yes | conjur init -u '${CONJUR_URL}' -a '${CONJUR_ACCOUNT}'

# API key here is the key that creation of the account provided you in step #2
conjur authn login -u admin -p '${CONJUR_ADMIN_API_KEY}'

# Check that you are identified as the admin user
conjur authn whoami
"

  store_set CONJUR_CLIENT_POD_NAME "${CONJUR_CLIENT_POD_NAME}"
}

function register_conjur_admin_key() {
  local ADMIN_USER_ID CONJUR_ADMIN_API_KEY
  ADMIN_USER_ID="$(store_get CONJUR_ACCOUNT):user:admin"
  CONJUR_ADMIN_API_KEY=$(exec_conjur \
    conjurctl role retrieve-key "${ADMIN_USER_ID}")

  store_set CONJUR_ADMIN_API_KEY "${CONJUR_ADMIN_API_KEY}"
}

function setup_conjur() {
  local CONJUR_ACCOUNT
  CONJUR_ACCOUNT="$(store_get CONJUR_ACCOUNT)"

  exec_conjur conjurctl account create "${CONJUR_ACCOUNT}"
}

function run_policy() {
  local AUTHENTICATOR_ID CONJUR_ACCOUNT
  AUTHENTICATOR_ID="$(store_get AUTHENTICATOR_ID)"
  CONJUR_ACCOUNT="$(store_get CONJUR_ACCOUNT)"

  local POLICY
  POLICY="$(./policy.yml.sh)"

  store_set "POLICY" "${POLICY}"
  echo -n "${POLICY}" | exec_conjur_client conjur policy load --replace root /dev/stdin

  exec_conjur_client bash -exc "
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
  local APP_SERVICE_ACCOUNT TEST_NAMESPACE
  APP_SERVICE_ACCOUNT="$(store_get APP_SERVICE_ACCOUNT)"
  TEST_NAMESPACE="$(store_get TEST_NAMESPACE)"

  echo -n "Create namespace: ${TEST_NAMESPACE}." | tee -a info-log.txt
  kubectl create namespace "${TEST_NAMESPACE}"
  echo " ✅ " | tee -a info-log.txt

  echo -n "Deploy Postgres." | tee -a info-log.txt
  deploy_pg
  echo " ✅ " | tee -a info-log.txt

  echo -n "Deploy Conjur." | tee -a info-log.txt
  deploy_conjur
  register_conjur_pod
  echo " ✅ " | tee -a info-log.txt

  echo "Wait for Conjur." | tee -a info-log.txt
  exec_conjur conjurctl wait | tee -a info-log.txt
  echo " ✅ " | tee -a info-log.txt

  echo -n "Setup Conjur account." | tee -a info-log.txt
  setup_conjur
  register_conjur_admin_key
  echo " ✅ " | tee -a info-log.txt

  echo -n "Apply policy and populate variables on Conjur." | tee -a info-log.txt
  register_conjur_client_pod
  run_policy
  populate_variables
  echo " ✅ " | tee -a info-log.txt

  echo -n "Deploy App With Secretless." | tee -a info-log.txt
  kubectl --namespace "${TEST_NAMESPACE}" \
    create sa "${APP_SERVICE_ACCOUNT}"
  echo " ✅ " | tee -a info-log.txt

  local APP_DEPLOYMENT
  APP_DEPLOYMENT="$(./app_secretless_deployment.yml.sh)"
  store_set "APP_DEPLOYMENT" "${APP_DEPLOYMENT}"
  echo -n "${APP_DEPLOYMENT}" | kubectl --namespace "${TEST_NAMESPACE}" apply -f -

  local APP_POD_NAME
  APP_POD_NAME=$(kubectl --namespace "${TEST_NAMESPACE}" get pods \
                   -l "app=test-app" \
                   -o jsonpath="{.items[0].metadata.name}")
  store_set APP_POD_NAME "${APP_POD_NAME}"
  kubectl --namespace "${TEST_NAMESPACE}" wait \
    --for=condition=ready "pod/${APP_POD_NAME}" --timeout 150s
}

main
