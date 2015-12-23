package main

var runtimeSrc string = `
package glimmer_runtime

// import (
// 	"reflect"
// )

type Glimmer {
	// map with keys that are the original chans and values that are their corresponding mocks
	mockChannels map[interface{}]interface{}
}

func init() {
	glimmer := new(Glimmer)
}

func (g *Glimmer) AddChan(ch chan interface{}) chan interface{} {
	
}

`
