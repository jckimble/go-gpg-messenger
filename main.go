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
func main(){
	//TODO: add multible database options
	db,err:=sql.Open("sqlite3","./ggm.db")
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
