package main

import (
	"net/http"
	"os"
	"fmt"
)

func main() {
	self := ForceParseNode(os.Args[1])

	nodes := []*Node{self}
	for _, u := range os.Args[2:] {
		nodes = append(nodes, ForceParseNode(u))
	}

	p := NewHTTPPolling(self, nodes)

	p.StartLoop()

	http.ListenAndServe(fmt.Sprintf(":%d", self.Port()), p)
}
