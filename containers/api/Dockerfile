# Dynamic Builds
ARG BUILDER_IMAGE=golang:1.21-buster
ARG FINAL_IMAGE=debian:buster-slim

# Build stage
FROM --platform=${BUILDPLATFORM} ${BUILDER_IMAGE} AS builder

# Build Args
ARG GIT_REVISION=""

# Ensure ca-certificates are up to date on the image
RUN update-ca-certificates

# Use modules for dependencies
WORKDIR $GOPATH/src/github.com/bbengfort/epistolary

COPY go.mod .
COPY go.sum .

ENV CGO_ENABLED=0
ENV GO111MODULE=on
RUN go mod download
RUN go mod verify

# Copy package
COPY . .

# Build binary
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -o /go/bin/epistolary -ldflags="-X 'github.com/bbengfort/epistolary/pkg.GitVersion=${GIT_REVISION}'" ./cmd/epistolary

# Final Stage
FROM --platform=${BUILDPLATFORM} ${FINAL_IMAGE} AS final

LABEL maintainer="Benjamin Bengfort <benjamin@bengfort.com>"
LABEL description="Epistolary API Server"

# Ensure ca-certificates are up to date
RUN set -x && apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates postgresql-client && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage
COPY --from=builder /go/bin/epistolary /usr/local/bin/epistolary

CMD [ "/usr/local/bin/epistolary", "serve" ]