FROM ubuntu:14.04

MAINTAINER virifi <virifi129@gmail.com>

RUN \
  apt-get update && \
  apt-get install -y ca-certificates build-essential gcc-multilib g++-multilib curl unzip git dos2unix \
    --no-install-recommends

RUN curl -LO https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz
RUN tar xzvf go1.4.2.linux-amd64.tar.gz && mv /go /go_amd64

RUN curl -LO https://storage.googleapis.com/golang/go1.4.2.linux-386.tar.gz
RUN tar xzvf go1.4.2.linux-386.tar.gz && mv /go /go_386

RUN git clone https://github.com/golang/go.git /go

ADD build.sh /build.sh
RUN dos2unix /build.sh
