#!/bin/bash

# Generate scan service protos
protoc -I=./protocol \
    -I$GOPATH/src/github.com/lindaprotocol/grpc-gateway/third_party/googleapis \
    --go_out=plugins=grpc:../../../ \
    ./protocol/api/scan_api.proto

# Generate reverse-proxy for scan service
protoc -I=./protocol \
    -I$GOPATH/src/github.com/lindaprotocol/grpc-gateway/third_party/googleapis \
    --grpc-gateway_out=logtostderr=true:../../../ \
    ./protocol/api/scan_api.proto

# Generate swagger definitions for scan service
protoc -I./protocol \
    -I$GOPATH/src \
    -I$GOPATH/src/github.com/lindaprotocol/grpc-gateway/third_party/googleapis \
    --swagger_out=logtostderr=true:../../../ \
    ./protocol/api/scan_api.proto