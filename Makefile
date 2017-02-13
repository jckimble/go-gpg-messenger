all: depends server client
.PHONY: depends
depends:
	go get github.com/gorilla/websocket
	go get golang.org/x/crypto/openpgp
	go get golang.org/x/crypto/bcrypt
	go get github.com/mattn/go-sqlite3
	go get github.com/go-sql-driver/mysql
	go get github.com/lib/pq
	go get github.com/spf13/viper
server: main.go dns.go messages.go websocket.go
	go build -o $@ $^
client: client.go dns.go messages.go
	go build -o $@ $^
.PHONY: clean
clean:
	rm -rf server client pkg src
