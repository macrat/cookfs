package main

import (
	"os"
)

func main() {
	self := ForceParseNode(os.Args[1])

	nodes := []*Node{self}
	for _, u := range os.Args[2:] {
		nodes = append(nodes, ForceParseNode(u))
	}

	store := DummyStore{}
	discover := SimpleDiscoverPlugin{self, nodes}
	transmit := &HTTPTransmitPlugin{}
	receive := &HTTPReceivePlugin{}

	cookfs := NewCookFS(store, discover, transmit, receive)

	stop := make(chan struct{})
	cookfs.Run(stop)
	for {
		select {
		case <-stop:
			panic("stopped")
			return
		}
	}
}
