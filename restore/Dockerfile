# Use the offical golang image to create a binary.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.20.3@sha256:403f48633fb5ebd49f9a2b6ad6719f912df23dae44974a0c9445be331e72ff5e as builder

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

FROM google/cloud-sdk@sha256:d07ec17c1e737633a5487579bef51c346b0674cc44b137b00e4599d7bc8b1753
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
