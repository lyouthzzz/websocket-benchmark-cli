package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"time"
)

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade: ", err)
	}

	defer c.Close()

	log.Println("new connection")
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			if err = c.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("write:", err)
				break
			}

			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		if err = c.WriteMessage(mt, message); err != nil {
			log.Println("write:", err)
			break
		}
	}
	log.Println("disconnect connection")
}

func server() {
	http.HandleFunc("/ws", echo)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func call() {
	uri := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(uri.String(), nil)
	if err != nil {
		panic(err)
	}
	//if err := c.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
	//	panic(err)
	//}

	c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	//if err := c.Close(); err != nil {
	//	panic(err)
	//}

	time.Sleep(time.Hour)
}

func main() {
	go server()

	call()
}
