package glimmer

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type MessageEvent struct {
	Func  string
	Type  bool // true is recieving, false is sending
	Chan  string
	Value string
}

var (
	forSendingQueue chan *MessageEvent

	delay     int = 1000
	delayLock sync.RWMutex
)

func init() {
	// 1024 seems a reasonable buffer size for this
	// TODO: consider using the channels with infinite buffers
	// from https://github.com/eapache/channels
	forSendingQueue = make(chan *MessageEvent, 1024)
}

func handler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()

	for {
		select {
		// TODO: read
		// write message events to the websocket
		case m := <-forSendingQueue:
			if err := conn.WriteJSON(<-forSendingQueue); err != nil {
				log.Error(err)
				return
			}
		}
	}
}

func ProcessRecieve(ch, value interface{}) {
	// get the caller of the caller of this function
	// we get two levels below the current because
	// this function is being called by the function literal
	// that substitutes the recieve expression
	programCounter, _, _, ok := runtime.Caller(2)
	if !ok {
		panic("Can't read the stack trace to find who called the ProcessRecieve function. Have no idea how to handle this.")
	}

	caller := runtime.FuncForPC(programCounter)
	fmt.Println("Recieve called from", caller.Name())

	sendMessageEvent(caller.Name(), fmt.Sprintf("%d", &ch), fmt.Sprintf("%d", &value), true)
}

func ProcessSend(ch, value interface{}) {
	// get the caller of the caller of this function
	// we get two levels below the current because
	// this function is being called by the function literal
	// that substitutes the recieve expression
	programCounter, _, _, ok := runtime.Caller(2)
	if !ok {
		panic("Can't read the stack trace to find who called the ProcessRecieve function. Have no idea how to handle this.")
	}

	caller := runtime.FuncForPC(programCounter)
	fmt.Println("Send called from", caller.Name())

	sendMessageEvent(caller.Name(), fmt.Sprintf("%d", &ch), fmt.Sprintf("%d", &value), false)
}

func sendMessageEvent(funcName, ch, value string, eventType bool) {
	messageEvent := &MessageEvent{
		Func:  funcName,
		Type:  eventType,
		Chan:  ch,
		Value: value,
	}

	forSendingQueue <- messageEvent
}

func Sleep() {
	delayLock.RLock()
	amount := delay
	delayLock.RUnlock()

	time.Sleep(time.Duration(amount) * time.Millisecond)
}
