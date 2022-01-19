FROM golang:1.17-buster

RUN mkdir -p /src
WORKDIR /src

RUN curl -sSL https://sdk.cloud.google.com | bash
ENV PATH $PATH:/root/google-cloud-sdk/bin

RUN gcloud components install kubectl

# Install Docker client
RUN apt-get update -y && \
  apt-get install -y apt-transport-https ca-certificates curl gnupg2 \
  software-properties-common wget && \
  curl -fsSL \
  https://download.docker.com/linux/$(. /etc/os-release; echo "$ID")/gpg \
  | apt-key add - && \
  add-apt-repository "deb [arch=amd64] \
  https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") \
  $(lsb_release -cs) stable" && \
  apt-get update && \
  apt-get install -y docker-ce && \
  rm -rf /var/lib/apt/lists/*

# Install Helm
RUN HELM_VERSION=v3.2.0 && \
  wget https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz && \
  tar xvf helm-${HELM_VERSION}-linux-amd64.tar.gz && \
  mv linux-amd64/helm /usr/local/bin/ && \
  rm helm-${HELM_VERSION}-linux-amd64.tar.gz && \
  rm -rf linux-amd64

# Install Helm bitnami repo
RUN helm repo add bitnami https://charts.bitnami.com/bitnami && \
  helm repo update
