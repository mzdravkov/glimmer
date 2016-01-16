package main

import (
	"reflect"
	"sync"
)

type Channels struct {
	Input, Output reflect.Value
}

type Glimmer struct {
	Channels map[reflect.Value]*Channels
	Mutex    *sync.Mutex
}

var glimmer Glimmer

func init() {
	glimmer = Glimmer{
		Channels: make(map[reflect.Value]*Channels),
	}

	glimmer.StartTransponder()
}

//TODO: find a good way to optimize this to not take all the channels from scratch every time
func (g *Glimmer) StartTransponder() {
	go func() {
		for {
			keys := make([]reflect.SelectCase, len(g.Channels))
			for key := range g.Channels {
				selectCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: key}
				keys = append(keys, selectCase)
			}

			chosen, inputMessage, _ := reflect.Select(keys)

			outputMessage := g.ProcessMessage(inputMessage)

			g.Channels[keys[chosen].Chan].Output.Send(outputMessage)
		}
	}()
}

//TODO
func (g *Glimmer) ProcessMessage(message reflect.Value) reflect.Value {
	return message
}

//TODO shoild I make a DeleteChan method that to be invoked on delete(chan) calls?
func (g *Glimmer) AddChan(ch interface{}) *Channels {
	key := reflect.ValueOf(ch)

	g.Mutex.Lock()
	defer func() { g.Mutex.Unlock() }()

	if mockChannels, ok := g.Channels[key]; ok {
		return mockChannels
	}

	chType := reflect.TypeOf(ch).Elem()

	inputChan := reflect.MakeChan(chType, reflect.ValueOf(ch).Cap())
	outputChan := reflect.MakeChan(chType, reflect.ValueOf(ch).Cap())

	g.Channels[key] = &Channels{Input: inputChan, Output: outputChan}

	return g.Channels[key]
}

// this will be called from the modified client's code
func (g *Glimmer) Send(ch interface{}, value interface{}) {
	mockChan := g.AddChan(ch)

	println("sending ", value)
	mockChan.Input.Send(reflect.ValueOf(value))
}

// this will be called from the modified client's code
func (g *Glimmer) Recieve(ch interface{}) interface{} {
	mockChan := g.AddChan(ch)

	result, _ := mockChan.Output.Recv()

	return result.Interface()
}

// this will be called from the modified client's code
func (g *Glimmer) RecieveWithBool(ch interface{}) (interface{}, bool) {
	mockChan := g.AddChan(ch)

	result, ok := mockChan.Output.Recv()

	return result.Interface(), ok
}
