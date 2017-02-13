package main

import (
	"fmt"
	"net/http"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)
func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not Found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("location","/ws")
}
const (
	MySQL = iota
	Postgres
	SQLite
)
var driver = map[int]string{
	0: "mysql",
	1: "postgres",
	2: "sqlite3",
}
var CREATEUSERS = map[int]string{
	MySQL: "CREATE TABLE IF NOT EXISTS users (id INT UNSIGNED NOT NULL AUTO_INCREMENT, name VARCHAR(255), domain VARCHAR(255), password VARCHAR(255), key VARCHAR(16), port INT UNSIGNED NOT NULL, PRIMARY KEY (id))",
	Postgres: "CREATE SEQUENCE users_seq; CREATE TABLE IF NOT EXISTS users (id INT CHECK (id > 0) NOT NULL DEFAULT NEXTVAL ('users_seq'), name VARCHAR(255), domain VARCHAR(255), password VARCHAR(255), port INT CHECK (port > 0) NOT NULL, PRIMARY KEY (id)) create index VARCHAR on users(16);",
	SQLite: "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, name VARCHAR(255), domain VARCHAR(255), password VARCHAR(255), key VARCHAR(16), port INTEGER UNSIGNED NOT NULL)",
}
var CREATEMESSAGES = map[int]string{
	MySQL: "CREATE TABLE IF NOT EXISTS messages (`id` INT UNSIGNED NOT NULL AUTO_INCREMENT, `to` INT UNSIGNED NOT NULL, `from` INT UNSIGNED NOT NULL, `time` INT UNSIGNED NOT NULL, `type` VARCHAR(255), `message` longtext NULL, PRIMARY KEY (id))",
	Postgres: "CREATE SEQUENCE messages_seq; CREATE TABLE IF NOT EXISTS messages (id INT CHECK (id > 0) NOT NULL DEFAULT NEXTVAL ('messages_seq'), to INT CHECK (to > 0) NOT NULL, from INT CHECK (from > 0) NOT NULL, time INT CHECK (time > 0) NOT NULL, type VARCHAR(255), message longtext NULL, PRIMARY KEY (id))",
	SQLite: "CREATE TABLE IF NOT EXISTS messages (`id` INTEGER PRIMARY KEY AUTOINCREMENT, `to` INTEGER UNSIGNED NOT NULL, `from` INTEGER UNSIGNED NOT NULL, `time` INTEGER UNSIGNED NOT NULL, `type` VARCHAR(255), `message` longtext NULL)",
}
func main(){
	//TODO: add multible database options
	db,err:=sql.Open(driver[SQLite],"./ggm.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	_,err = db.Exec(CREATEUSERS[SQLite])
	if err != nil {
		fmt.Println(err)
		return
	}
	_,err = db.Exec(CREATEMESSAGES[SQLite])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()
	tracker := newTracker()
	go tracker.run()
	http.HandleFunc("/",serveHome)
	http.HandleFunc("/ws",func(w http.ResponseWriter, r *http.Request) {
		serveWebsocket(db,tracker,w,r)
	})
	err = http.ListenAndServe(":8080",nil)
	if err != nil {
		fmt.Println(err)
	}
}
