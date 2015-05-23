package math

import (
	"DFE/builder"
)

type Add func(m builder.ActorManagement, ach, bch <-chan float64, cch chan<- float64)

func init() {
	var p Add = AddImpl
	builder.Register(p)
}

func AddImpl(m builder.ActorManagement, ach, bch <-chan float64, cch chan<- float64) {
	var a, b, c, ctmp float64

	for {
		select {
		case a = <-ach:
		case b = <-bch:
		}
		ctmp = a + b
		if c != ctmp {
			c = ctmp
			cch <- c
		}
	}
}
