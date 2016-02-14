package glimmer

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	conn *websocket.Conn

	delay     int
	delayLock sync.RWMutex
)

func init() {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	var err error

	conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err = conn.WriteMessage(messageType, p); err != nil {
			return err
		}
	}
}

func ProcessRecieve(ch, value interface{}) {
	fmt.Println("recieve", &ch, &value)
}

func ProcessSend(ch, value interface{}) {
	fmt.Println("send", &ch, &value)
}

func Sleep() {
	delayLock.RLock()
	amount := delay
	delayLock.RUnlock()

	time.Sleep(amount * time.Millisecond)
}
