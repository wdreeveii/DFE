package builder

import (
	"fmt"
	"reflect"
)

var actors map[string]actor

type actor struct {
	Func  interface{}
	Ports []reflect.Type
}

func init() {
	actors = make(map[string]actor)
}

func Register(v interface{}) {
	var p actor
	p.Func = v
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumIn(); i++ {
		in := t.In(i)
		if in.Kind() == reflect.Chan {
			p.Ports = append(p.Ports, in)
		}
	}
	actors[t.String()] = p
	/*var test chan float64
	taskt := reflect.TypeOf(v)

	fmt.Println("typeof", taskt)
	fmt.Println("string", taskt.String())
	fmt.Println("name", taskt.Name())
	fmt.Println("pkg", taskt.PkgPath())
	for i := 0; i < taskt.NumIn(); i++ {
		p := taskt.In(i)
		if p.Kind() == reflect.Chan {
			fmt.Println("param", i, p, p.ChanDir())
			if reflect.TypeOf(test).ConvertibleTo(p) {
				fmt.Println("convertible")
			}
		}
	}*/
}

func GetActors() []string {
	var p []string
	for k, _ := range actors {
		p = append(p, k)
	}
	return p
}

type ActorManagement struct {
}

type NodeID int64

type FlowGraph struct {
	Nodes   map[NodeID]Node
	AutoInc NodeID

	Edges []Edge
}

type Node struct {
	Name  string
	actor actor
}

type PortID int

type PortAddress struct {
	NodeID NodeID
	PortID PortID
}

type Edge struct {
	Caller PortAddress
	Callee PortAddress
}

func NewFlowGraph() *FlowGraph {
	var p = new(FlowGraph)
	p.Nodes = make(map[NodeID]Node)
	return p
}

func (graph *FlowGraph) StartFlowGraph() error {
	var params = make(map[NodeID][]reflect.Value)

	for k, v := range graph.Nodes {
		callType := reflect.TypeOf(v.actor.Func)
		var p []reflect.Value
		for i := 0; i < callType.NumIn(); i++ {
			p = append(p, reflect.Zero(callType.In(i)))
		}
		params[k] = p
	}

	var channels []interface{}
	for _, v := range graph.Edges {
		callerPortType := graph.Nodes[v.Caller.NodeID].actor.Ports[v.Caller.PortID]
		transportType := reflect.ChanOf(reflect.BothDir, callerPortType.Elem())
		newchan := reflect.MakeChan(transportType, 0)
		channels = append(channels, newchan)

		params[v.Caller.NodeID][v.Caller.PortID+1] = newchan
		params[v.Callee.NodeID][v.Callee.PortID+1] = newchan
	}

	for k, v := range params {
		actorfunc := reflect.ValueOf(graph.Nodes[k].actor.Func)
		go actorfunc.Call(v)
	}

	for k, v := range graph.Nodes {
		fmt.Println("Node", k, ":", v, ":", reflect.ValueOf(v.actor.Func))
	}
	return nil
}

func (graph *FlowGraph) StopFlowGraph() error {
	return nil
}

func (graph *FlowGraph) AddNode(n string) (NodeID, error) {
	f, exists := actors[n]
	if !exists {
		return 0, fmt.Errorf("Unrecognized actor: %s", n)
	}

	var p = Node{Name: n, actor: f}
	var newid = graph.AutoInc
	graph.Nodes[newid] = p
	graph.AutoInc++
	return newid, nil
}

func (graph *FlowGraph) DeleteNode(id NodeID) error {
	delete(graph.Nodes, id)
	return nil
}

func (graph *FlowGraph) AddEdge(caller, callee PortAddress) error {
	callerNode, exists := graph.Nodes[caller.NodeID]
	if !exists {
		return fmt.Errorf("Caller Node not found: %v", caller)
	}

	calleeNode, exists := graph.Nodes[callee.NodeID]
	if !exists {
		return fmt.Errorf("Callee Node not found: %v", callee)
	}

	//fmt.Println("comparison", callerNode.actor.Ports[caller.PortID], calleeNode.actor.Ports[callee.PortID])

	if caller.PortID >= PortID(len(callerNode.actor.Ports)) {
		return fmt.Errorf("Caller Port not found: %v", caller)
	}

	if callee.PortID >= PortID(len(calleeNode.actor.Ports)) {
		return fmt.Errorf("Callee Port not found: %v", callee)
	}

	callerPortType := callerNode.actor.Ports[caller.PortID]
	calleePortType := calleeNode.actor.Ports[callee.PortID]
	transportType := reflect.ChanOf(reflect.BothDir, callerPortType.Elem())

	if !transportType.ConvertibleTo(calleePortType) {
		return fmt.Errorf("Type missmatch transport: %v callee: %v", transportType, calleePortType)
	}

	combinedDir := callerPortType.ChanDir() | calleePortType.ChanDir()
	if combinedDir != reflect.BothDir {
		return fmt.Errorf("Type missmatch: edge does not have a sender and a receiver")
	}
	graph.Edges = append(graph.Edges, Edge{caller, callee})
	return nil
}
