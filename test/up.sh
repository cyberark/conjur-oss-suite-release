#!/usr/bin/env bash
set -euo pipefail

. ./store.sh

function init_state() {
  init_store

  set_val HELM_CHART_RELEASE_VERSION 1.3.8
  set_val CONJUR_OSS_RELEASE_VERSION 1.4.1
  set_val CONJUR_KUBERNETES_AUTHENTICATOR_RELEASE_VERSION 0.15.0

  set_val TEST_NAMESPACE test-kumbi

  set_val PG_RELEASE_NAME test-conjur-oss-db
  set_val DB_PASSWORD databasepassword

  set_val CONJUR_RELEASE_NAME test-conjur-oss
  set_val DATA_KEY c+KONZjTUk9zWib4IKeAX5kUltghDtEH7JJOJYxjc9A=

  set_val CONJUR_ACCOUNT another
  set_val CONJUR_ADMIN_API_KEY "to be retrieved dynamically"
  set_val AUTHENTICATOR_ID testing
  set_val CONJUR_URL "https://$(get_val CONJUR_RELEASE_NAME).$(get_val TEST_NAMESPACE).svc.cluster.local"

  set_val APP_SERVICE_ACCOUNT test-app

  set_val CLI_POD_NAME "test-client"
}

function cleanup {
  kubectl delete clusterrolebinding --ignore-not-found "$(get_val CONJUR_RELEASE_NAME)-conjur-authenticator"
  kubectl delete clusterrole --ignore-not-found "$(get_val CONJUR_RELEASE_NAME)-conjur-authenticator"
  kubectl delete namespace "$(get_val TEST_NAMESPACE)" --ignore-not-found
  kubectl create namespace "$(get_val TEST_NAMESPACE)"
}

function deploy_pg {
  helm install --namespace "$(get_val TEST_NAMESPACE)" "$(get_val PG_RELEASE_NAME)" \
   --set postgresqlPassword="$(get_val DB_PASSWORD)"\
   stable/postgresql
}

function deploy_conjur {
  helm install --namespace "$(get_val TEST_NAMESPACE)" "$(get_val CONJUR_RELEASE_NAME)" \
    --set dataKey="$(get_val DATA_KEY)" \
    --set account="$(get_val CONJUR_ACCOUNT)" \
    --set image.tag="$(get_val CONJUR_OSS_RELEASE_VERSION)" \
    --set image.pullPolicy="Always" \
    --set authenticators="authn-k8s/$(get_val AUTHENTICATOR_ID)\,authn" \
    --set databaseUrl="postgres://postgres:$(get_val DB_PASSWORD)@$(get_val PG_RELEASE_NAME)-postgresql.$(get_val TEST_NAMESPACE).svc.cluster.local/postgres" \
    "https://github.com/cyberark/conjur-oss-helm-chart/releases/download/v$(get_val HELM_CHART_RELEASE_VERSION)/conjur-oss-$(get_val HELM_CHART_RELEASE_VERSION).tgz"
}

function register_conjur_pod() {
  local CONJUR_POD_NAME=$(kubectl --namespace "$(get_val TEST_NAMESPACE)" get pods \
                   -l "app=conjur-oss,release=$(get_val CONJUR_RELEASE_NAME)" \
                   -o jsonpath="{.items[0].metadata.name}")

  set_val CONJUR_POD_NAME "${CONJUR_POD_NAME}"
}

function exec_conjur() {
  kubectl --namespace "$(get_val TEST_NAMESPACE)" \
   exec -i "$(get_val CONJUR_POD_NAME)" --container=conjur-oss \
    -- "$@"
}

function register_conjur_client_pod() {
  local CONJUR_CLIENT_POD_NAME="conjur-client-pod"

  # start the CLI pod
  kubectl --namespace "$(get_val TEST_NAMESPACE)" run "${CONJUR_CLIENT_POD_NAME}" \
   --restart='Never' \
   --image cyberark/conjur-cli:5 \
   --command -- sleep infinity

  # wait for CLI pod to be ready
  kubectl --namespace "$(get_val TEST_NAMESPACE)" \
   wait --for=condition=ready "pod/${CONJUR_CLIENT_POD_NAME}" --timeout 150s

  # login to conjur
  kubectl --namespace "$(get_val TEST_NAMESPACE)" \
    exec -i "${CONJUR_CLIENT_POD_NAME}" -- \
     bash -xce "
yes yes | conjur init -u '$(get_val CONJUR_URL)' -a '$(get_val CONJUR_ACCOUNT)'

# API key here is the key that creation of the account provided you in step #2
conjur authn login -u admin -p '$(get_val CONJUR_ADMIN_API_KEY)'

# Check that you are identified as the admin user
conjur authn whoami
"

  set_val CONJUR_CLIENT_POD_NAME "${CONJUR_CLIENT_POD_NAME}"
}

function exec_conjur_client() {
  kubectl --namespace "$(get_val TEST_NAMESPACE)" \
   exec -i "$(get_val CONJUR_CLIENT_POD_NAME)" \
    -- "$@"
}

function register_conjur_admin_key() {
  local CONJUR_ADMIN_API_KEY=$(exec_conjur \
   conjurctl role retrieve-key "$(get_val CONJUR_ACCOUNT):user:admin")

  set_val CONJUR_ADMIN_API_KEY "${CONJUR_ADMIN_API_KEY}"
}

function setup_conjur() {
  exec_conjur conjurctl account create "$(get_val CONJUR_ACCOUNT)"
}

function run_policy() {
  set_val "policy" "$(./policy.yml.sh)"
  get_val "policy" | exec_conjur_client conjur policy load --replace root /dev/stdin

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
  -subj '/CN=conjur.authn-k8s.$(get_val AUTHENTICATOR_ID)/OU=Conjur Kubernetes CA/O=$(get_val CONJUR_ACCOUNT)' \
  -config <(echo \"\${CONFIG}\")

# Verify cert
openssl x509 -in ca.cert -text -noout

# Set cert values
conjur variable values add conjur/authn-k8s/$(get_val AUTHENTICATOR_ID)/ca/key \"\$(cat ca.key)\"
conjur variable values add conjur/authn-k8s/$(get_val AUTHENTICATOR_ID)/ca/cert \"\$(cat ca.cert)\"
"
}

function populate_variables() {
    exec_conjur_client conjur variable values add test-app-secrets/username 'meow meow meow'
}

function main() {
  echo "Initialise state."
  init_state

  echo "Clean up."
  cleanup

  echo "Deploy Postgres."
  deploy_pg

  echo "Deploy Conjur."
  deploy_conjur

  echo "Wait for Conjur."
  register_conjur_pod
  kubectl --namespace "$(get_val TEST_NAMESPACE)" \
   wait --for=condition=ready "pod/$(get_val CONJUR_POD_NAME)" --timeout 150s
  exec_conjur conjurctl wait


  echo "Setup Conjur."
  setup_conjur
  register_conjur_admin_key

  echo "Run policy on Conjur."
  register_conjur_client_pod
  run_policy
  populate_variables

  echo "Deploy app."
  kubectl create sa "$(get_val APP_SERVICE_ACCOUNT)"
  set_val "app_deployment" "$(./app_deployment.yml.sh)"
  get_val "app_deployment" | kubectl --namespace "$(get_val TEST_NAMESPACE)" apply -f -

  local APP_POD_NAME=$(kubectl --namespace "$(get_val TEST_NAMESPACE)" get pods \
                     -l "app=test-app" \
                     -o jsonpath="{.items[0].metadata.name}")
  set_val APP_POD_NAME "${APP_POD_NAME}"
  kubectl --namespace "$(get_val TEST_NAMESPACE)" wait \
   --for=condition=ready "pod/${APP_POD_NAME}" --timeout 150s
}

main
