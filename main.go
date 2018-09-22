package main

import (
	"os"

	"net/http"
	_ "net/http/pprof"
)

func main() {
	self := ForceParseNode(os.Args[1])

	nodes := []*Node{self}
	for _, u := range os.Args[2:] {
		nodes = append(nodes, ForceParseNode(u))
	}

	recipe := NewInMemoryRecipeStore()
	chunk := NewInMemoryChunkStore()
	discover := SimpleDiscoverPlugin{self, nodes}
	transmit := &HTTPTransmitPlugin{}
	receive := NewHTTPReceivePlugin()

	cookfs := NewCookFS(recipe, chunk, discover, transmit, receive)

	go func() {
		http.ListenAndServe(":3000", nil)
	}()

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
