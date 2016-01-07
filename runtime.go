package main

import (
	"reflect"
	"sync"
)

type Glimmer struct {
	// map with keys that are the original chans and values that are their corresponding mocks
	MockChannels map[reflect.Value]reflect.Value
	Mutex        *sync.Mutex
}

var glimmer Glimmer

func init() {
	glimmer = Glimmer{
		MockChannels: make(map[reflect.Value]reflect.Value),
	}
}

func (g *Glimmer) AddChan(ch interface{}) reflect.Value {

	chType := reflect.TypeOf(ch).Elem()

	key := reflect.ValueOf(ch)

	g.Mutex.Lock()
	defer func() { g.Mutex.Unlock() }()

	if mockChan, ok := glimmer.MockChannels[key]; ok {
		return mockChan
	}

	mockChan := reflect.MakeChan(chType, reflect.ValueOf(ch).Cap())

	glimmer.MockChannels[key] = mockChan

	return mockChan
}

func (g *Glimmer) Send(ch interface{}, value interface{}) {
	mockChan := g.AddChan(ch)

	mockChan.Send(reflect.ValueOf(value))
}

func (g *Glimmer) Recieve(ch interface{}) (interface{}, bool) {
	// chType := reflect.ValueOf(ch).Type().Elem()

	mockChan := g.AddChan(ch)

	result, ok := mockChan.Recv()

	return result.Interface(), ok
}
