package main

// var runtimeSrc string = `
// package glimmer_runtime

import (
	"reflect"
)

// type Channel struct {
// 	Chan chan interface{}
// 	Type reflect.Type
// }

// type Channel reflect.Value

type Glimmer struct {
	// map with keys that are the original chans and values that are their corresponding mocks
	MockChannels map[reflect.Value]reflect.Value
}

var glimmer Glimmer

func init() {
	glimmer = Glimmer{
		MockChannels: make(map[reflect.Value]reflect.Value),
	}
}

func (g *Glimmer) AddChan(ch chan interface{}) reflect.Value {

	chType := reflect.TypeOf(ch).Elem()

	// key := Channel{}
	// key.Chan = ch
	// key.Type = chType
	key := reflect.ValueOf(ch)

	if mockChan, ok := glimmer.MockChannels[key]; ok {
		return mockChan
	}

	// mockChan := Channel{}
	// mockChan.Chan = reflect.MakeChan(chType, cap(ch))
	// mockChan.Type = chType
	mockChan := reflect.MakeChan(chType, cap(ch))

	glimmer.MockChannels[key] = mockChan

	return mockChan
}

func (g *Glimmer) Send(ch chan interface{}, value interface{}) {
	mockChan := g.AddChan(ch)

	valueAsReflectionObj := reflect.ValueOf(value)

	mockChan.Send(valueAsReflectionObj)
}

func (g *Glimmer) Recieve(ch chan interface{}) {

}
