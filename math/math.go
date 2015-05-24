package math

import (
	"DFE/builder"
)

type Add func(m builder.ActorManagement, ach, bch <-chan float64, cch chan<- float64)

func init() {
	var p Add = AddImpl
	builder.Register(p)
}

type AddState struct {
	a float64
	b float64
	c float64
}

func AddImpl(m builder.ActorManagement, ach, bch <-chan float64, cch chan<- float64) {
	var ctmp float64

	state, _ := m.State.(AddState)

	for {
		select {
		case d := <-m.Done:
			d <- state
			return
		case state.a = <-ach:
		case state.b = <-bch:
		}
		ctmp = state.a + state.b
		if state.c != ctmp {
			state.c = ctmp
			cch <- state.c
		}
	}
}
