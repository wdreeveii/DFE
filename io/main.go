package io

import (
	"DFE/builder"
	"fmt"
)

type Println func(m builder.ActorManagement, in <-chan float64)

func init() {
	var p Println = PrintlnImplementation
	builder.Register(p)
}

func PrintlnImplementation(m builder.ActorManagement, in <-chan float64) {
	for {
		select {
		case d := <-m.Done:
			d <- nil
			return
		case v := <-in:
			fmt.Println(v)
		}
	}
}
