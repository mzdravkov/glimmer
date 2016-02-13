package glimmer

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
)

type ChanGuard struct {
	Chan      interface{}
	Type      reflect.Type
	semaphore chan struct{}
	RecvLock  sync.Mutex
	SendLock  sync.Mutex
}

func MakeChanGuard(ch interface{}) ChanGuard {
	return ChanGuard{
		Chan:      ch,
		Type:      reflect.ValueOf(ch).Type(),
		semaphore: make(chan struct{}),
	}
}

func (cg *ChanGuard) Recieve() interface{} {
	if cg.Cap() == 0 {
		<-cg.semaphore
	}

	// TODO: PROCESS RECIEVE EVENT

	cg.RecvLock.Lock()
	defer cg.RecvLock.Unlock()

	// get the caller of the Recieve function
	programCounter, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("Can't read the stack trace to find who called the Recieve function. Have no idea how to handle this.")
	}

	caller := runtime.FuncForPC(programCounter)
	fmt.Println("Recieve called from", caller.Name())

	result, _ := reflect.ValueOf(cg.Chan).Recv()

	return result
}

func (cg *ChanGuard) RecieveWithBool() (interface{}, bool) {
	if cg.Cap() == 0 {
		<-cg.semaphore
	}

	// TODO: PROCESS RECIEVE EVENT

	cg.RecvLock.Lock()
	defer cg.RecvLock.Unlock()

	// get the caller of the RecieveWiwSithBool function
	programCounter, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("Can't read the stack trace to find who called the RecieveWithBool function. Have no idea how to handle this.")
	}

	caller := runtime.FuncForPC(programCounter)
	fmt.Println("RecieveWithBool called from", caller.Name())

	return reflect.ValueOf(cg.Chan).Recv()
}

func (cg *ChanGuard) Send(value interface{}) {
	if cg.Cap() == 0 {
		cg.semaphore <- struct{}{}
	}

	// TODO: PROCESS SEND EVENT

	cg.SendLock.Lock()
	defer cg.SendLock.Unlock()

	// get the caller of the Send function
	programCounter, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("Can't read the stack trace to find who called the Send function. Have no idea how to handle this.")
	}

	caller := runtime.FuncForPC(programCounter)
	fmt.Println("Send called from", caller.Name())

	reflect.ValueOf(cg.Chan).Send(reflect.ValueOf(value))
}

func (cg *ChanGuard) Len() int {
	return reflect.ValueOf(cg.Chan).Len()
}

func (cg *ChanGuard) Cap() int {
	return reflect.ValueOf(cg.Chan).Cap()
}

func (cg *ChanGuard) Close() {
	reflect.ValueOf(cg.Chan).Close()
}

//TODO
func ProcessMessage(message reflect.Value) reflect.Value {
	return message
}
