package main

import (
	"fmt"
	"bytes"
	"strconv"
	"net/url"
	"encoding/json"
	"time"
	"net/http"
//	"database/sql"
//	_ "github.com/mattn/go-sqlite3"
	"github.com/gorilla/websocket"
)
type EmailRequest struct {
	Addr string
}
type Tracker struct {
	clients map[*Client]bool
	register chan *Client
	unregister chan *Client
}
type Client struct {
	tracker *Tracker
	conn *websocket.Conn
	send chan []byte
	username string `omitempty`
}
type FetchQuery struct {
	from string `omitempty`
	limit int `omitempty`
	start int `omitempty`
	end int `omitempty`
}

func newTracker() *Tracker {
	return &Tracker{
		register: make(chan *Client),
		unregister: make(chan *Client),
		clients: make(map[*Client]bool),
	}
}
func (t *Tracker) run(){
	for {
		select {
		case client := <-t.register:
			t.clients[client] = true
		case client := <-t.unregister:
			if _, ok := t.clients[client]; ok {
				delete(t.clients, client)
				close(client.send)
			}
		}
	}
}

func (c *Client) readPump(){
	defer func(){
		c.tracker.unregister <-c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60*time.Second))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(60*time.Second)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,websocket.CloseGoingAway){
				fmt.Println(err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message,[]byte{'\n'},[]byte{' '},-1))
		//Authorize
		auth := new(UserAuth)
		authErr := json.Unmarshal([]byte(message),auth)
		if authErr == nil {
			//TODO: check db for login
			c.username=auth.Username
			continue
		}
		if c.username != "" {
			//get connection info
			email := new(EmailRequest)
			err := json.Unmarshal([]byte(message),email)
			if err == nil {
				e := getAddr(email.Addr)
				json,_ := json.Marshal(e)
				c.send<-json
				continue
			}
			//fetch messages
			fetch := new(FetchQuery)
			fetchErr:= json.Unmarshal([]byte(message),fetch)
			if fetchErr == nil {
				//c.send<-message
				//TODO: from,limit,start,end
				continue
			}
			//TODO: delete messages
		}
		//recieve message
		msg:=parseMessage(string(message))
		if msg != nil {
			if checkMessage(msg) {
				//TODO: allow multchat
				//TODO: pull keys from database
				keys:=[2]string{"90636FD8C5C8B273","030E11640B5F1EA1"}
				for _,key := range keys {
					if msg.To.Key == key {
						for client,_ := range c.tracker.clients {
							if client.username == msg.From.Name+"@"+msg.From.Domain {
								client.send<-message
							}
						}
						//TODO: add to queue
					}else if msg.From.Key == key {
						if msg.From.Name+"@"+msg.From.Domain == c.username {
							//TODO: add to queue
							c.send<-message
							sendMessage(msg)
						}
					}
				}
			}
		}
	}
}
func (c *Client) writePump() {
	ticker := time.NewTicker(54*time.Second)
	defer func(){
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
			case message,ok := <-c.send:
				c.conn.SetWriteDeadline(time.Now().Add(10*time.Second))
				if !ok {
					c.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				w, err := c.conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				w.Write(message)
				n := len(c.send)
				for i := 0; i < n; i++ {
					w.Write([]byte{'\n'})
					w.Write(<-c.send)
				}
				if err := w.Close(); err != nil {
					return
				}
			case <-ticker.C:
				c.conn.SetWriteDeadline(time.Now().Add(10*time.Second))
				if err := c.conn.WriteMessage(websocket.PingMessage,[]byte{}); err != nil {
					return
				}
		}
	}
}
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}
func serveWebsocket(tracker *Tracker, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w,r,nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := &Client{tracker: tracker,conn: conn,send: make(chan []byte, 256)}
	client.tracker.register <- client
	go client.writePump()
	client.readPump()
}
func sendMessage(msg* Message) (bool) {
	var data,_=json.Marshal(msg)
	u:=url.URL{Scheme: "ws",Host:msg.To.Domain+":"+strconv.Itoa(msg.To.Port), Path:"/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(),nil)
	if err != nil {
		fmt.Printf(err.Error())
		return false
	}
	defer c.Close()

	err = c.WriteMessage(websocket.TextMessage,data)
	if err != nil {
		fmt.Printf(err.Error())
		return false
	}
	c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	return false
}
