#!/usr/bin/env bash
set -euo pipefail

. ./store.sh

CONJUR_AUTHN_URL="$(store_get CONJUR_URL)/authn-k8s/$(store_get AUTHENTICATOR_ID)"
CONJUR_APPLIANCE_URL="$(store_get CONJUR_URL)"
CONJUR_ACCOUNT="$(store_get CONJUR_ACCOUNT)"
CONJUR_AUTHN_LOGIN="host/conjur/authn-k8s/$(store_get AUTHENTICATOR_ID)/apps/$(store_get TEST_NAMESPACE)/service_account/$(store_get APP_SERVICE_ACCOUNT)"
CONJUR_SSL_CERTIFICATE_SECRET="$(store_get CONJUR_RELEASE_NAME)-conjur-ssl-cert"
APP_SERVICE_ACCOUNT="$(store_get APP_SERVICE_ACCOUNT)"
SECRETLESS_RELEASE_VERSION="$(store_get SECRETLESS_RELEASE_VERSION)"

# language=YAML
echo "
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: secretless-config
data:
  secretless.yml: |
    version: '2'
    services:
      http:
        connector: generic_http
        listenOn: tcp://0.0.0.0:8080
        credentials:
          username:
            get: test-app-secrets/username
            from: conjur
        config:
          credentialPatterns:
            username: '[^:]+'
          headers:
            Authorization: '{{ .username }}'
          authenticateURLsMatching:
            - ^http
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: test-app
  name: test-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-app
  template:
    metadata:
      labels:
        app: test-app
    spec:
      serviceAccountName: ${APP_SERVICE_ACCOUNT}
      containers:
      - image: cyberark/conjur-cli:5
        imagePullPolicy: Always
        name: app
        command: ['sleep', 'infinity']
      - image: cyberark/secretless-broker:${SECRETLESS_RELEASE_VERSION}
        imagePullPolicy: Always
        name: authenticator
        args: ['-f', '/etc/secretless/secretless.yml', '--debug']
        env:
          - name: MY_POD_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          - name: MY_POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: MY_POD_IP
            valueFrom:
              fieldRef:
                fieldPath: status.podIP
          - name: CONJUR_APPLIANCE_URL
            value: '${CONJUR_APPLIANCE_URL}'
          - name: CONJUR_AUTHN_URL
            value: '${CONJUR_AUTHN_URL}'
          - name: CONJUR_ACCOUNT
            value: '${CONJUR_ACCOUNT}'
          - name: CONJUR_AUTHN_LOGIN
            value: '${CONJUR_AUTHN_LOGIN}'
          - name: CONJUR_SSL_CERTIFICATE
            valueFrom:
              secretKeyRef:
                name: '${CONJUR_SSL_CERTIFICATE_SECRET}'
                key: 'tls.crt'
        volumeMounts:
          - mountPath: /etc/secretless
            name: config
            readOnly: true
      volumes:
        - name: conjur-access-token
          emptyDir:
            medium: Memory
        - name: config
          configMap:
            name: secretless-config
            defaultMode: 420
"
