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

// for debugging
// func init() {
// 	go func() {
// 		for {
// 			fmt.Println(<-forSendingQueue)
// 		}
// 	}()
// }

type MessageEvent struct {
	Func  string
	Type  bool // true is recieving, false is sending
	Chan  string
	Value string
}

type chanLock struct {
	Send    *sync.Mutex
	Recieve *sync.Mutex
}

var (
	locks = make(map[uintptr]*chanLock)

	forSendingQueue chan *MessageEvent

	delay     int = 1000
	delayLock sync.RWMutex
)

func Locks(ch uintptr) *chanLock {
	if chLock, ok := locks[ch]; ok {
		return chLock
	}

	locks[ch] = &chanLock{
		Send:    new(sync.Mutex),
		Recieve: new(sync.Mutex),
	}
	return locks[ch]
}

func init() {
	// 1024 seems a reasonable buffer size for this
	// TODO-min: consider using the channels with infinite buffers
	// from https://github.com/eapache/channels
	forSendingQueue = make(chan *MessageEvent, 1024)

	go func() {
		http.HandleFunc("/", handler)
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			panic(err)
		}
	}()
}

func handler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	for {
		select {
		// TODO: read
		// write message events to the websocket
		case m := <-forSendingQueue:
			if err := conn.WriteJSON(m); err != nil {
				log.Fatal(err)
				// TODO-min: should I exit here?
				return
			}
		}
	}
}

func ProcessRecieve(ch uintptr, value interface{}) {
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

	go sendMessageEvent(ch, caller.Name(), fmt.Sprintf("%d", value), true)
}

func ProcessSend(ch uintptr, value interface{}) {
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

	go sendMessageEvent(ch, caller.Name(), fmt.Sprintf("%d", value), false)
}

func sendMessageEvent(ch uintptr, funcName, value string, eventType bool) {
	messageEvent := &MessageEvent{
		Func:  funcName,
		Type:  eventType,
		Chan:  fmt.Sprintf("%d", ch),
		Value: value,
	}

	forSendingQueue <- messageEvent

	if eventType {
		Locks(ch).Recieve.Unlock()
	} else {
		Locks(ch).Send.Unlock()
	}
}

func Sleep() {
	delayLock.RLock()
	amount := delay
	delayLock.RUnlock()

	time.Sleep(time.Duration(amount) * time.Millisecond)
}
