package util

import (
	"DFE/builder"
	//"fmt"
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

type CounterState struct {
	a     float64
	inval float64
}

func CounterImplementation(m builder.ActorManagement, in <-chan float64, out chan<- float64) {
	state, _ := m.State.(CounterState)

	out <- state.a
	for {
		select {
		case d := <-m.Done:
			d <- state
			return
		case v := <-in:
			if state.inval <= 0 && v > 0 {
				state.a++
				out <- state.a
			}
			state.inval = v
		}
	}
}

func SquareWaveImplementation(m builder.ActorManagement, out chan<- float64) {
	ticker := time.Tick(500 * time.Millisecond)

	a, _ := m.State.(float64)

	out <- a
	for {
		select {
		case d := <-m.Done:
			d <- a
			return
		case _ = <-ticker:
			if a == 0 {
				a = 1
			} else {
				a = 0
			}
			out <- a
		}
	}
}
