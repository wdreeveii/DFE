package util

import (
	"DFE/builder"
	"time"
)

type Counter func(m builder.ActorManagement, in <-chan float64, out chan<- float64)
type SquareWave func(m builder.ActorManagement, out chan<- float64)

func init() {
	var c Counter = CounterImplementation
	builder.Register(c)
	var sq SquareWave = SquareWaveImplementation
	builder.Register(sq)
}

func CounterImplementation(m builder.ActorManagement, in <-chan float64, out chan<- float64) {
	var a float64
	var inval float64

	out <- a
	for {
		v := <-in
		if inval <= 0 && v > 0 {
			a++
			out <- a
		}
		inval = v
	}
}

func SquareWaveImplementation(m builder.ActorManagement, out chan<- float64) {
	ticker := time.Tick(500 * time.Millisecond)
	var a float64
	out <- a
	for {
		<-ticker
		if a == 0 {
			a = 1
		} else {
			a = 0
		}
		out <- a
	}
}
