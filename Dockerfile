# Build stage
FROM golang:1.16-alpine AS builder

RUN apk add --no-cache git make protobuf

WORKDIR /go/src/github.com/lindaprotocol/grpc-gateway

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate protos
RUN make generate

# Build main gateway
RUN go build -o /go/bin/grpc-gateway ./cmd/grpc-gateway

# Build scan server
RUN go build -o /go/bin/scan-server ./cmd/scan-server

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binaries
COPY --from=builder /go/bin/grpc-gateway /app/
COPY --from=builder /go/bin/scan-server /app/

# Copy configuration
COPY configs /app/configs

EXPOSE 18889 50051

# Default command (can be overridden)
CMD ["/app/grpc-gateway"]