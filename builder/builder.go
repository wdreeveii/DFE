package builder

import (
	"fmt"
	"reflect"
	"sync"
)

var actors map[string]Actor

type Actor struct {
	Func  interface{}
	Ports []reflect.Type
}

func init() {
	actors = make(map[string]Actor)
}

func Register(v interface{}) {
	var p Actor
	p.Func = v
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumIn(); i++ {
		in := t.In(i)
		if in.Kind() == reflect.Chan {
			p.Ports = append(p.Ports, in)
		}
	}
	actors[t.String()] = p
}

func GetActors() []string {
	var p []string
	for k, _ := range actors {
		p = append(p, k)
	}
	return p
}

type ActorConfig map[string]interface{}
type ActorState interface{}

type ActorManagement struct {
	Config ActorConfig
	State  ActorState
	Done   chan chan ActorState
}

type NodeID int64

type FlowGraph struct {
	Nodes   map[NodeID]Node
	AutoInc NodeID
	nmu     sync.Mutex

	Edges []Edge
	emu   sync.Mutex
}

type Node struct {
	Name  string
	Actor Actor
	ActorManagement
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
	graph.nmu.Lock()
	for k, v := range graph.Nodes {
		callType := reflect.TypeOf(v.Actor.Func)
		var p []reflect.Value
		for i := 0; i < callType.NumIn(); i++ {
			ptype := callType.In(i)

			if reflect.TypeOf(v.ActorManagement).AssignableTo(ptype) {
				p = append(p, reflect.ValueOf(v.ActorManagement))
			} else {
				p = append(p, reflect.Zero(ptype))
			}
		}
		params[k] = p
	}

	var channels []interface{}
	for _, v := range graph.Edges {
		callerPortType := graph.Nodes[v.Caller.NodeID].Actor.Ports[v.Caller.PortID]
		transportType := reflect.ChanOf(reflect.BothDir, callerPortType.Elem())
		newchan := reflect.MakeChan(transportType, 1)
		channels = append(channels, newchan)

		params[v.Caller.NodeID][v.Caller.PortID+1] = newchan
		params[v.Callee.NodeID][v.Callee.PortID+1] = newchan
	}

	for k, v := range params {
		actorfunc := reflect.ValueOf(graph.Nodes[k].Actor.Func)
		go actorfunc.Call(v)
	}

	for k, v := range graph.Nodes {
		fmt.Println("Node", k, ":", v, ":", reflect.ValueOf(v.Actor.Func))
	}
	graph.nmu.Unlock()
	return nil
}

func (graph *FlowGraph) StopFlowGraph() error {
	graph.nmu.Lock()

	for k, v := range graph.Nodes {
		fmt.Println("killing:", v)
		dchan := make(chan ActorState)
		v.ActorManagement.Done <- dchan
		v.State = <-dchan
		graph.Nodes[k] = v
	}

	graph.nmu.Unlock()
	return nil
}

func (graph *FlowGraph) AddNode(n string) (NodeID, error) {
	f, exists := actors[n]
	if !exists {
		return 0, fmt.Errorf("Unrecognized actor: %s", n)
	}
	var a = ActorManagement{Done: make(chan chan ActorState)}
	var p = Node{Name: n, Actor: f, ActorManagement: a}

	graph.nmu.Lock()

	var newid = graph.AutoInc
	graph.Nodes[newid] = p
	graph.AutoInc++

	graph.nmu.Unlock()

	return newid, nil
}

func (graph *FlowGraph) DeleteNode(id NodeID) error {
	graph.nmu.Lock()

	// delete node
	delete(graph.Nodes, id)

	graph.nmu.Unlock()

	graph.emu.Lock()
	//delete edges connected to the deleted node
	for k, v := range graph.Edges {
		if v.Caller.NodeID == id || v.Callee.NodeID == id {
			graph.Edges[k], graph.Edges = graph.Edges[len(graph.Edges)-1], graph.Edges[:len(graph.Edges)-1]
		}
	}

	graph.emu.Unlock()

	return nil
}

func (graph *FlowGraph) AddEdge(caller, callee PortAddress) error {
	graph.nmu.Lock()
	callerNode, exists := graph.Nodes[caller.NodeID]
	if !exists {
		graph.nmu.Unlock()
		return fmt.Errorf("Caller Node not found: %v", caller)
	}

	calleeNode, exists := graph.Nodes[callee.NodeID]
	if !exists {
		graph.nmu.Unlock()
		return fmt.Errorf("Callee Node not found: %v", callee)
	}

	if caller.PortID >= PortID(len(callerNode.Actor.Ports)) {
		graph.nmu.Unlock()
		return fmt.Errorf("Caller Port not found: %v", caller)
	}

	if callee.PortID >= PortID(len(calleeNode.Actor.Ports)) {
		graph.nmu.Unlock()
		return fmt.Errorf("Callee Port not found: %v", callee)
	}

	callerPortType := callerNode.Actor.Ports[caller.PortID]
	calleePortType := calleeNode.Actor.Ports[callee.PortID]
	transportType := reflect.ChanOf(reflect.BothDir, callerPortType.Elem())

	if !transportType.ConvertibleTo(calleePortType) {
		graph.nmu.Unlock()
		return fmt.Errorf("Type missmatch transport: %v callee: %v", transportType, calleePortType)
	}

	combinedDir := callerPortType.ChanDir() | calleePortType.ChanDir()
	if combinedDir != reflect.BothDir {
		graph.nmu.Unlock()
		return fmt.Errorf("Type missmatch: edge does not have a sender and a receiver")
	}
	graph.nmu.Unlock()
	graph.emu.Lock()

	for _, v := range graph.Edges {
		if v.Caller == caller || v.Callee == caller {
			graph.emu.Unlock()
			return fmt.Errorf("Edge using caller: %v already exists", caller)
		}
		if v.Caller == callee || v.Callee == callee {
			graph.emu.Unlock()
			return fmt.Errorf("Edge using callee: %v already exists", callee)
		}
	}

	graph.Edges = append(graph.Edges, Edge{caller, callee})

	graph.emu.Unlock()

	return nil
}
