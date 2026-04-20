FROM alpine:latest

RUN apk add --no-cache \
    ca-certificates \
    go \
    nodejs \
    npm \
    git \
    bash

ENV GOPATH=/go
ENV PATH=$PATH:/go/bin

WORKDIR /workspace
