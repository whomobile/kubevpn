FROM ubuntu:latest

RUN if [ "$NO_UBUNTU_MIRROR" != "true" ]; then \
      sed -i s@/security.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list && \
      sed -i s@/archive.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list; \
    fi
RUN apt-get clean && apt-get update && apt-get install -y wget dnsutils vim curl  \
    net-tools iptables iputils-ping lsof iproute2 tcpdump binutils traceroute conntrack socat iperf3
