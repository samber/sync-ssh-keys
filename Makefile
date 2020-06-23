
SRC=main.go output.go
BIN=sync-ssh-keys
DOCKER_IMAGE=samber/sync-ssh-keys
VERSION=$(shell cat VERSION)
BUILD_ID := $(shell git rev-parse --short HEAD)

LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD_ID)"

all: build

build:
	CGO_ENABLED=0 GO111MODULE=on go build -v $(LDFLAGS) -o $(BIN) $(SRC)

docker-build:
	docker build -t $(DOCKER_IMAGE):${VERSION} .

release: clean deps
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build $(LDFLAGS) -o $(BIN)_$(VERSION)_linux-amd64 $(SRC)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GO111MODULE=on go build $(LDFLAGS) -o $(BIN)_$(VERSION)_linux-arm64 $(SRC)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GO111MODULE=on go build $(LDFLAGS) -o $(BIN)_$(VERSION)_linux-arm $(SRC)
	CGO_ENABLED=0 GOOS=freebsd GOARCH=386 GO111MODULE=on go build $(LDFLAGS) -o $(BIN)_$(VERSION)_freebsd-386 $(SRC)

docker-release: docker-build
	docker push $(DOCKER_IMAGE):${VERSION}

run-dev:
	GO111MODULE=on go run -v $(LDFLAGS) ${SRC} --github-username samber

clean:
	rm -f $(BIN) ${BIN}_*

deps:
	GO111MODULE=on go mod download

re: clean all
