#!/usr/bin/env bash
set -euo pipefail

. ./store.sh
. ./executors.sh

function deploy_pg {
  helm install --namespace "$(store_get TEST_NAMESPACE)" "$(store_get PG_RELEASE_NAME)" \
   --set postgresqlPassword="$(store_get DB_PASSWORD)" \
   --set persistence.enabled="false" \
   stable/postgresql
}

function deploy_conjur {
  helm install --namespace "$(store_get TEST_NAMESPACE)" "$(store_get CONJUR_RELEASE_NAME)" \
    --set dataKey="$(store_get DATA_KEY)" \
    --set account="$(store_get CONJUR_ACCOUNT)" \
    --set image.tag="$(store_get CONJUR_OSS_RELEASE_VERSION)" \
    --set image.pullPolicy="Always" \
    --set authenticators="authn-k8s/$(store_get AUTHENTICATOR_ID)\,authn" \
    --set databaseUrl="postgres://postgres:$(store_get DB_PASSWORD)@$(store_get PG_RELEASE_NAME)-postgresql.$(store_get TEST_NAMESPACE).svc.cluster.local/postgres" \
    "https://github.com/cyberark/conjur-oss-helm-chart/releases/download/v$(store_get HELM_CHART_RELEASE_VERSION)/conjur-oss-$(store_get HELM_CHART_RELEASE_VERSION).tgz"
}

function register_conjur_pod() {
  local CONJUR_POD_NAME=$(kubectl --namespace "$(store_get TEST_NAMESPACE)" get pods \
                   -l "app=conjur-oss,release=$(store_get CONJUR_RELEASE_NAME)" \
                   -o jsonpath="{.items[0].metadata.name}")

  store_set CONJUR_POD_NAME "${CONJUR_POD_NAME}"
}

function register_conjur_client_pod() {
  local CONJUR_CLIENT_POD_NAME="conjur-client-pod"

  # start the CLI pod
  kubectl --namespace "$(store_get TEST_NAMESPACE)" run "${CONJUR_CLIENT_POD_NAME}" \
   --restart='Never' \
   --image cyberark/conjur-cli:5 \
   --command -- sleep infinity

  # wait for CLI pod to be ready
  kubectl --namespace "$(store_get TEST_NAMESPACE)" \
   wait --for=condition=ready "pod/${CONJUR_CLIENT_POD_NAME}" --timeout 150s

  # login to conjur
  kubectl --namespace "$(store_get TEST_NAMESPACE)" \
    exec -i "${CONJUR_CLIENT_POD_NAME}" -- \
     bash -xce "
yes yes | conjur init -u '$(store_get CONJUR_URL)' -a '$(store_get CONJUR_ACCOUNT)'

# API key here is the key that creation of the account provided you in step #2
conjur authn login -u admin -p '$(store_get CONJUR_ADMIN_API_KEY)'

# Check that you are identified as the admin user
conjur authn whoami
"

  store_set CONJUR_CLIENT_POD_NAME "${CONJUR_CLIENT_POD_NAME}"
}

function register_conjur_admin_key() {
  local CONJUR_ADMIN_API_KEY=$(exec_conjur \
   conjurctl role retrieve-key "$(store_get CONJUR_ACCOUNT):user:admin")

  store_set CONJUR_ADMIN_API_KEY "${CONJUR_ADMIN_API_KEY}"
}

function setup_conjur() {
  exec_conjur conjurctl account create "$(store_get CONJUR_ACCOUNT)"
}

function run_policy() {
  store_set "policy" "$(./policy.yml.sh)"
  store_get "policy" | exec_conjur_client conjur policy load --replace root /dev/stdin

  exec_conjur_client bash -xce "
# Generate OpenSSL private key
openssl genrsa -out ca.key 2048

CONFIG='
[ req ]
distinguished_name = dn
x509_extensions = v3_ca
[ dn ]
[ v3_ca ]
basicConstraints = critical,CA:TRUE
subjectKeyIdentifier   = hash
authorityKeyIdentifier = keyid:always,issuer:always
'

# Generate root CA certificate
openssl req -x509 -new -nodes -key ca.key -sha1 -days 3650 -set_serial 0x0 -out ca.cert \
  -subj '/CN=conjur.authn-k8s.$(store_get AUTHENTICATOR_ID)/OU=Conjur Kubernetes CA/O=$(store_get CONJUR_ACCOUNT)' \
  -config <(echo \"\${CONFIG}\")

# Verify cert
openssl x509 -in ca.cert -text -noout

# Set cert values
conjur variable values add conjur/authn-k8s/$(store_get AUTHENTICATOR_ID)/ca/key \"\$(cat ca.key)\"
conjur variable values add conjur/authn-k8s/$(store_get AUTHENTICATOR_ID)/ca/cert \"\$(cat ca.cert)\"
"
}

function populate_variables() {
    exec_conjur_client conjur variable values add test-app-secrets/username 'meow meow meow'
}

function main() {
  kubectl create namespace "$(store_get TEST_NAMESPACE)"

  echo "Deploy Postgres."
  deploy_pg

  echo "Deploy Conjur."
  deploy_conjur

  echo "Wait for Conjur."
  register_conjur_pod
  kubectl --namespace "$(store_get TEST_NAMESPACE)" \
   wait --for=condition=ready "pod/$(store_get CONJUR_POD_NAME)" --timeout 150s
  exec_conjur conjurctl wait


  echo "Setup Conjur."
  setup_conjur
  register_conjur_admin_key

  echo "Run policy on Conjur."
  register_conjur_client_pod
  run_policy
  populate_variables

  echo "Deploy app."
  kubectl --namespace "$(store_get TEST_NAMESPACE)" \
    create sa "$(store_get APP_SERVICE_ACCOUNT)"
  store_set "app_deployment" "$(./app_deployment.yml.sh)"
  store_get "app_deployment" | kubectl --namespace "$(store_get TEST_NAMESPACE)" apply -f -

  local APP_POD_NAME=$(kubectl --namespace "$(store_get TEST_NAMESPACE)" get pods \
                     -l "app=test-app" \
                     -o jsonpath="{.items[0].metadata.name}")
  store_set APP_POD_NAME "${APP_POD_NAME}"
  kubectl --namespace "$(store_get TEST_NAMESPACE)" wait \
   --for=condition=ready "pod/${APP_POD_NAME}" --timeout 150s
}

main
