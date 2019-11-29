FROM golang:1.12-stretch AS builder

ENV GO111MODULE=on

# Download tools
RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin v1.17.1

# Fetch dependencies
WORKDIR /app
COPY go.mod go.sum Makefile /app/
RUN make deps

# Now pull in our code
COPY . .

RUN make build
# Copy binary to alpine
FROM alpine:3.10
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/sync-ssh-keys /bin/sync-ssh-keys

ENTRYPOINT ["/bin/sync-ssh-keys"]
