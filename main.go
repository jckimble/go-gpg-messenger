package main

import (
	"fmt"
	"net/http"
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
	tracker := newTracker()
	go tracker.run()
	http.HandleFunc("/",serveHome)
	http.HandleFunc("/ws",func(w http.ResponseWriter, r *http.Request) {
		serveWebsocket(tracker,w,r)
	})
	err := http.ListenAndServe(":8080",nil)
	if err != nil {
		fmt.Println(err)
	}
}
