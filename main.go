package main

import (
	"os"

	"./cookfs"

	"net/http"
	_ "net/http/pprof"
)

func main() {
	self := cookfs.ForceParseNode(os.Args[1])

	nodes := []*cookfs.Node{self}
	for _, u := range os.Args[2:] {
		nodes = append(nodes, cookfs.ForceParseNode(u))
	}

	recipe := cookfs.NewInMemoryRecipeStore()
	chunk := cookfs.NewInMemoryChunkStore()
	discover := cookfs.SimpleDiscoverPlugin{self, nodes}
	transmit := &cookfs.HTTPTransmitPlugin{}
	receive := cookfs.NewHTTPReceivePlugin()

	c := cookfs.NewCookFS(recipe, chunk, discover, transmit, receive)

	go func() {
		http.ListenAndServe(":3000", nil)
	}()

	stop := make(chan struct{})
	c.Run(stop)
	for {
		select {
		case <-stop:
			panic("stopped")
			return
		}
	}
}
