FROM golang:1.20 AS builder
ARG NO_GO_PROXY
RUN if [ "$NO_GO_PROXY" != "true" ]; then go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct; fi

RUN go install github.com/go-delve/delve/cmd/dlv@latest

FROM envoyproxy/envoy:v1.25.0 AS envoy
ARG UBUNTUBASE=ubuntu:latest
FROM ${UBUNTUBASE}
ARG NO_UBUNTU_MIRROR
ARG DOCKER_TIMEZONE=Asia/Shanghai
ARG NO_DOCKER_TIMEZONE

RUN if [ "$NO_UBUNTU_MIRROR" != "true" ]; then \
      sed -i s@/security.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list && \
      sed -i s@/archive.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list; \
    fi
RUN apt-get clean && apt-get update && apt-get install -y wget dnsutils vim curl  \
    net-tools iptables iputils-ping lsof iproute2 tcpdump binutils traceroute conntrack socat iperf3

ENV TZ=$DOCKER_TIMEZONE \
    DEBIAN_FRONTEND=noninteractive
RUN if [ -z "$NO_DOCKER_TIMEZONE" ]; then \
      apt update \
      && apt install -y tzdata \
      && ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime \
      && echo ${TZ} > /etc/timezone \
      && dpkg-reconfigure --frontend noninteractive tzdata \
      && rm -rf /var/lib/apt/lists/*; \
    fi

WORKDIR /app

COPY bin/kubevpn /usr/local/bin/kubevpn
COPY --from=envoy /usr/local/bin/envoy /usr/local/bin/envoy
COPY --from=builder /go/bin/dlv /usr/local/bin/dlv