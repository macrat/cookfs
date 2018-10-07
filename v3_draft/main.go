package main

import (
	"os"
	"context"
)

func Nodes() []*Node {
	ns := []*Node{}

	for _, x := range os.Args[1:] {
		ns = append(ns, MustParseNode(x))
	}

	return ns
}

func main() {
	ctx := context.Background()

	h := &HTTPHandler{}

	c := NewCookFS(h, Nodes)
	go c.RunFollower(ctx)

	h.Listen(ctx, Nodes()[0], c)
}
