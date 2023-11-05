ARG NAMESPACE=naison
ARG REPOSITORY=kubevpn
FROM ${NAMESPACE}/${REPOSITORY}:latest

WORKDIR /app

COPY bin/kubevpn /usr/local/bin/kubevpn