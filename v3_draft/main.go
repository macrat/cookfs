package main

import (
	"context"
)

func main() {
	ctx := context.Background()

	ch := make(chan RequestResponse)

	f := Follower{}
	go f.Run(ctx, ch)

	h := HTTPHandler{}
	h.Listen(ctx, ch)
}
