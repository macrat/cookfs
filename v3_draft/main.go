package main

import (
	"context"
)

func main() {
	ctx := context.Background()

	f := NewFollower()
	go f.Run(ctx)

	h := HTTPHandler{}
	h.Listen(ctx, f)
}
