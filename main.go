package main

import (
	"DFE/builder"
	_ "DFE/io"
	_ "DFE/math"
	_ "DFE/util"
	"fmt"
	"time"
)

func main() {
	/*var ach = make(chan float64)
	var bch = make(chan float64)
	var cch = make(chan float64)
	var dch = make(chan float64)
	var ech = make(chan float64)

	go math.Add(ach, bch, cch)
	go math.Add(cch, dch, ech)
	go io.Println(ech)
	ach <- 5
	dch <- 2
	<-dch*/

	fmt.Println(builder.GetActors())

	var g = builder.NewFlowGraph()
	n1, err := g.AddNode("math.Add")
	if err != nil {
		fmt.Println(err)
		return
	}
	n2, err := g.AddNode("math.Add")
	if err != nil {
		fmt.Println(err)
		return
	}
	n3, err := g.AddNode("math.Add")
	if err != nil {
		fmt.Println(err)
		return
	}
	n4, err := g.AddNode("io.Println")
	if err != nil {
		fmt.Println(err)
		return
	}
	n5, err := g.AddNode("util.Counter")
	if err != nil {
		fmt.Println(err)
		return
	}
	n6, err := g.AddNode("util.SquareWave")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = g.AddEdge(builder.PortAddress{n6, 0}, builder.PortAddress{n5, 0})
	if err != nil {
		fmt.Println(err)
		return
	}
	err = g.AddEdge(builder.PortAddress{n5, 1}, builder.PortAddress{n1, 0})
	if err != nil {
		fmt.Println(err)
		return
	}
	err = g.AddEdge(builder.PortAddress{n1, 2}, builder.PortAddress{n2, 0})
	if err != nil {
		fmt.Println(err)
		return
	}
	err = g.AddEdge(builder.PortAddress{n2, 2}, builder.PortAddress{n3, 0})
	if err != nil {
		fmt.Println(err)
		return
	}
	err = g.AddEdge(builder.PortAddress{n3, 2}, builder.PortAddress{n4, 0})
	if err != nil {
		fmt.Println(err)
		return
	}

	/*err = g.AddEdge(builder.PortAddress{n1, 0}, builder.PortAddress{n1, 1})
	if err != nil {
		fmt.Println(err)
		return
	}*/
	g.StartFlowGraph()

	time.Sleep(10 * time.Second)
}
