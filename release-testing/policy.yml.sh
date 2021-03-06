#!/usr/bin/env bash
set -euo pipefail

. ./store.sh

AUTHENTICATOR_ID="$(store_get AUTHENTICATOR_ID)"
TEST_NAMESPACE="$(store_get TEST_NAMESPACE)"
APP_SERVICE_ACCOUNT="$(store_get APP_SERVICE_ACCOUNT)"

echo "
---
# This policy defines an authn-k8s endpoint, CA creds and a layer for whitelisted identities permitted to authenticate to it
- !policy
  id: conjur/authn-k8s/${AUTHENTICATOR_ID}
  owner: !user admin
  annotations:
    description: Namespace defs for the Conjur cluster in dev
  body:
  - !webservice
    annotations:
      description: authn service for cluster

  - !policy
    id: ca
    body:
    - !variable
      id: cert
      annotations:
        description: CA cert for Kubernetes Pods.
    - !variable
      id: key
      annotations:
        description: CA key for Kubernetes Pods.


# This policy defines a layer of whitelisted identities permitted to authenticate to the authn-k8s endpoint.
- !policy
  id: conjur/authn-k8s/${AUTHENTICATOR_ID}/apps
  owner: !user admin
  annotations:
    description: Identities permitted to authenticate
  body:

  # Define layer of whitelisted authn ids permitted to call authn service
  - !layer
    annotations:
      description: Layer of authenticator identities permitted to call authn svc

  - &hosts
    - !host
      id: ${TEST_NAMESPACE}/service_account/${APP_SERVICE_ACCOUNT}
      annotations:
        kubernetes/authentication-container-name: authenticator
        kubernetes: true

  - !grant
    role: !layer
    members: *hosts

- !policy
  id: test-app
  owner: !user admin
  annotations:
    description: This policy connects authn identities to an application identity. It defines a layer named for an application that contains the whitelisted identities that can authenticate to the authn-k8s endpoint. Any permissions granted to the application layer will be inherited by the whitelisted authn identities, thereby granting access to the authenticated identity.
  body:
  - !layer

- !policy
  id: test-app-secrets
  owner: !user admin
  annotations:
    description: This policy contains our test app creds

  body:
    - &variables
      - !variable password
      - !variable url
      - !variable port
      - !variable host
      - !variable username

    - !permit
      role: !layer /test-app
      privileges: [ read, execute ]
      resources: *variables

# Add authn identities to application layer so authn roles inherit app's permissions
- !grant
  role: !layer test-app
  members:
  - !layer conjur/authn-k8s/${AUTHENTICATOR_ID}/apps

- !permit
  resource: !webservice conjur/authn-k8s/${AUTHENTICATOR_ID}
  privilege: [ read, authenticate ]
  role: !layer conjur/authn-k8s/${AUTHENTICATOR_ID}/apps
"
