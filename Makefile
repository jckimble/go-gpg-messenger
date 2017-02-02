all: depends server client
.PHONY: depends
depends:
	go get github.com/gorilla/websocket
	go get golang.org/x/crypto/openpgp
server: main.go dns.go messages.go websocket.go
	go build -o $@ $^
client: client.go dns.go messages.go
	go build -o $@ $^
.PHONY: clean
clean:
	rm -rf server client pkg src
