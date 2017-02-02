package main

import (
	"net/url"
	"os"
	"flag"
	"fmt"
	"sync"
	"os/signal"
	"time"
	"strconv"
	"encoding/json"
	"github.com/gorilla/websocket"
)



func main() {
	host := flag.String("host","","Host Server To Login To(optional, will try to autodetect if not provided)")
	username := flag.String("username","","Login Username(Will be asked for if not provided)")
	password := flag.String("password","","Login Password(Will be asked for if not provided)")
	flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        flag.PrintDefaults()
	}
	flag.Parse()
	if *username == "" {
		fmt.Printf("Username: ")
		fmt.Scanln(username)
	}
	if *password == "" {
		fmt.Printf("Password: ")
		fmt.Scanln(password)
	}
	if *username == "" || *password == "" {
		return
	}
	addr := ""
	if *host == "" {
		info := getAddr(*username)
		if info != nil {
			addr=info.Domain+"@"+strconv.Itoa(info.Port)
		}else{
			fmt.Println("Unable to AutoDetect Server Information")
			return
		}
	}else{
		addr=*host
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	u:= url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err)
	}
	defer c.Close()

	msgqueue := make(chan []byte)
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			//Email
			//Message
			fmt.Printf("recv: %s",message)
		}
	}()

	go func(){
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case msg := <-msgqueue:
				w, err := c.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				w.Write(msg)
				n := len(msgqueue)
				for i := 0; i < n; i++ {
					w.Write([]byte{'\n'})
					w.Write(<-msgqueue)
				}
				if err := w.Close(); err != nil {
					return
				}
			case t := <-ticker.C:
				err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
				if err != nil {
					fmt.Println(err)
					return
				}
			case <-interrupt:
				err := c.WriteMessage(websocket.CloseMessage,websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					fmt.Println(err)
					return
				}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				c.Close()
				return
			}
		}
	}()


	auth := new(UserAuth)
	auth.Username=*username
	auth.Password=*password
	msg, _:=json.Marshal(auth)
	msgqueue<-msg

/*
	email := new(EmailRequest)
	email.Addr=
	msg, _:=json.Marshal(email)
*/
/*
	fetch := new(FetchQuery)
	//from,limit,start,end
	msg, _:=json.Marshal(fetch)
*/
/*
	message := new(Message)
	message.From=
	message.To=
	message.Message=
	message.Time=time.Now()
	msg, _:=json.Marshal(message)
*/
	wg.Wait()
}
