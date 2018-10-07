package main

import (
	"os"
	"context"

	"github.com/macrat/cookfs/cooklib"
	"github.com/macrat/cookfs/plugins"
)

func Nodes() []*cooklib.Node {
	ns := []*cooklib.Node{}

	for _, x := range os.Args[1:] {
		ns = append(ns, cooklib.MustParseNode(x))
	}

	return ns
}

func main() {
	ctx := context.Background()

	h := &plugins.HTTPHandler{}

	c := cooklib.NewCookFS(h, Nodes, cooklib.DefaultConfig)
	go c.RunFollower(ctx)

	h.Listen(ctx, Nodes()[0], c)
}
