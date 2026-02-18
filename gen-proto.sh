#!/bin/bash

PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
PROTO_INCLUDES="${PROJECT_ROOT}/protocol"
GOOGLE_APIS="${PROJECT_ROOT}/third_party/googleapis"

echo "Generating protos from: ${PROTO_INCLUDES}"

# Generate core protos
protoc -I=${PROTO_INCLUDES} \
    -I${GOOGLE_APIS} \
    --go_out=. \
    --go_opt=paths=source_relative \
    ${PROTO_INCLUDES}/core/*.proto

# Generate api.proto with all services
protoc -I=${PROTO_INCLUDES} \
    -I${GOOGLE_APIS} \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    --grpc-gateway_out=. \
    --grpc-gateway_opt=paths=source_relative \
    ${PROTO_INCLUDES}/api/api.proto

# Generate swagger definitions
protoc -I=${PROTO_INCLUDES} \
    -I${GOOGLE_APIS} \
    --openapiv2_out=. \
    --openapiv2_opt=logtostderr=true \
    ${PROTO_INCLUDES}/api/api.proto

echo "Proto generation complete!"