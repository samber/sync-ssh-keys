
SRC=main.go
BIN=github-team-ssh-keys
VERSION=$(shell cat VERSION)

all:
	GO111MODULE=on go build -o $(BIN) $(SRC)

run:
	GO111MODULE=on go run ${SRC} --github-org epitech --github-team sysadmin

release:
	GO111MODULE=on go build -o $(BIN)_$(VERSION)_linux-amd64 $(SRC)

clean:
	rm -f $(BIN)
