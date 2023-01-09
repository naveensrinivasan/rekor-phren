# Use the offical golang image to create a binary.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.19.4@sha256:660f138b4477001d65324a51fa158c1b868651b44e43f0953bf062e9f38b72f3 as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY invoke.go ./

# Build the binary.
RUN go build -mod=readonly -v -o server

FROM google/cloud-sdk@sha256:0b5e8d676210c484217afa10f69e30b9d7250f52bbc09ae164fea1766a377cb3
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    --no-install-recommends \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Create and change to the app directory.
WORKDIR /

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /app/server
COPY restore.sh ./

# Run the web service on container startup.
CMD ["/app/server"]
