#!/usr/bin/env bash
set -euo pipefail

. ./store.sh

store_set app.CONJUR_AUTHN_URL "$(store_get CONJUR_URL)/authn-k8s/$(store_get AUTHENTICATOR_ID)"
store_set app.CONJUR_APPLIANCE_URL "$(store_get CONJUR_URL)"
store_set app.CONJUR_ACCOUNT "$(store_get CONJUR_ACCOUNT)"
store_set app.CONJUR_AUTHN_LOGIN "host/conjur/authn-k8s/$(store_get AUTHENTICATOR_ID)/apps/$(store_get TEST_NAMESPACE)/service_account/$(store_get APP_SERVICE_ACCOUNT)"
store_set app.CONJUR_SSL_CERTIFICATE_SECRET "$(store_get CONJUR_RELEASE_NAME)-conjur-ssl-cert"
store_set app.APP_SERVICE_ACCOUNT "$(store_get APP_SERVICE_ACCOUNT)"
store_set app.CONJUR_KUBERNETES_AUTHENTICATOR_RELEASE_VERSION "$(store_get CONJUR_KUBERNETES_AUTHENTICATOR_RELEASE_VERSION)"

echo "
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
      serviceAccountName: $(store_get app.APP_SERVICE_ACCOUNT)
      containers:
      - image: cyberark/conjur-cli:5
        imagePullPolicy: Always
        name: app
        command: ['sleep', 'infinity']
        env:
          - name: CONJUR_APPLIANCE_URL
            value: '$(store_get app.CONJUR_APPLIANCE_URL)'
          - name: CONJUR_ACCOUNT
            value: '$(store_get app.CONJUR_ACCOUNT)'
          - name: CONJUR_AUTHN_LOGIN
            value: '$(store_get app.CONJUR_AUTHN_LOGIN)'
          - name: CONJUR_AUTHN_TOKEN_FILE
            value: /run/conjur/access-token
          - name: CONJUR_SSL_CERTIFICATE
            valueFrom:
              secretKeyRef:
                name: '$(store_get app.CONJUR_SSL_CERTIFICATE_SECRET)'
                key: 'tls.crt'
        volumeMounts:
          - mountPath: /run/conjur
            name: conjur-access-token
            readOnly: true
      - image: cyberark/conjur-kubernetes-authenticator:$(store_get app.CONJUR_KUBERNETES_AUTHENTICATOR_RELEASE_VERSION)
        imagePullPolicy: Always
        name: authenticator
        env:
          - name: CONTAINER_MODE
            value: sidecar
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
          - name: CONJUR_AUTHN_URL
            value: '$(store_get app.CONJUR_AUTHN_URL)'
          - name: CONJUR_ACCOUNT
            value: '$(store_get app.CONJUR_ACCOUNT)'
          - name: CONJUR_AUTHN_LOGIN
            value: '$(store_get app.CONJUR_AUTHN_LOGIN)'
          - name: CONJUR_SSL_CERTIFICATE
            valueFrom:
              secretKeyRef:
                name: '$(store_get app.CONJUR_SSL_CERTIFICATE_SECRET)'
                key: 'tls.crt'
        volumeMounts:
          - mountPath: /run/conjur
            name: conjur-access-token
      volumes:
        - name: conjur-access-token
          emptyDir:
            medium: Memory
"
