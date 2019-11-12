
SRC=main.go
BIN=github-team-ssh-keys

all:
	GO111MODULE=on go build -o $(BIN) $(SRC)

run:
	GO111MODULE=on go run ${SRC} --github-org epitech --github-team sysadmin

clean:
	rm -f $(BIN)
