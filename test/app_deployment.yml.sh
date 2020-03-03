#!/usr/bin/env bash

. ./store.sh

set_val app.CONJUR_AUTHN_URL "$(get_val CONJUR_URL)/authn-k8s/$(get_val AUTHENTICATOR_ID)"
set_val app.CONJUR_APPLIANCE_URL "$(get_val CONJUR_URL)"
set_val app.CONJUR_ACCOUNT "$(get_val CONJUR_ACCOUNT)"
set_val app.CONJUR_AUTHN_LOGIN "host/conjur/authn-k8s/$(get_val AUTHENTICATOR_ID)/apps/$(get_val TEST_NAMESPACE)/service_account/$(get_val APP_SERVICE_ACCOUNT)"
set_val app.CONJUR_SSL_CERTIFICATE_SECRET "$(get_val CONJUR_RELEASE_NAME)-conjur-ssl-cert"
set_val app.APP_SERVICE_ACCOUNT "$(get_val APP_SERVICE_ACCOUNT)"
set_val app.CONJUR_KUBERNETES_AUTHENTICATOR_RELEASE_VERSION "$(get_val CONJUR_KUBERNETES_AUTHENTICATOR_RELEASE_VERSION)"

echo "
---
apiVersion: apps/v1beta1
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
      serviceAccountName: $(get_val app.APP_SERVICE_ACCOUNT)
      containers:
      - image: cyberark/conjur-cli:5
        imagePullPolicy: Always
        name: app
        command: ['sleep', 'infinity']
        env:
          - name: CONJUR_APPLIANCE_URL
            value: '$(get_val app.CONJUR_APPLIANCE_URL)'
          - name: CONJUR_ACCOUNT
            value: '$(get_val app.CONJUR_ACCOUNT)'
          - name: CONJUR_AUTHN_LOGIN
            value: '$(get_val app.CONJUR_AUTHN_LOGIN)'
          - name: CONJUR_AUTHN_TOKEN_FILE
            value: /run/conjur/access-token
          - name: CONJUR_SSL_CERTIFICATE
            valueFrom:
              secretKeyRef:
                name: '$(get_val app.CONJUR_SSL_CERTIFICATE_SECRET)'
                key: 'tls.crt'
        volumeMounts:
          - mountPath: /run/conjur
            name: conjur-access-token
            readOnly: true
      - image: cyberark/conjur-kubernetes-authenticator:$(get_val app.CONJUR_KUBERNETES_AUTHENTICATOR_RELEASE_VERSION)
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
            value: '$(get_val app.CONJUR_AUTHN_URL)'
          - name: CONJUR_ACCOUNT
            value: '$(get_val app.CONJUR_ACCOUNT)'
          - name: CONJUR_AUTHN_LOGIN
            value: '$(get_val app.CONJUR_AUTHN_LOGIN)'
          - name: CONJUR_SSL_CERTIFICATE
            valueFrom:
              secretKeyRef:
                name: '$(get_val app.CONJUR_SSL_CERTIFICATE_SECRET)'
                key: 'tls.crt'
        volumeMounts:
          - mountPath: /run/conjur
            name: conjur-access-token
      volumes:
        - name: conjur-access-token
          emptyDir:
            medium: Memory
"
