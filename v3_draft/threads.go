package main

import (
	"net/url"
	"context"
)

type Follower struct {
	Leader *url.URL
	Term   int64
	State *State
}

func (f Follower) handle(request Request) Response {
	return Response{200, "hello world"}
}

func (f Follower) Run(ctx context.Context, ch chan RequestResponse) {
	for {
		select {
		case reqres := <-ch:
			reqres.Response <-f.handle(reqres.Request)

		case <-ctx.Done():
			return
		}
	}
}

func Candidacy(ctx context.Context, termID int64) {
	// TODO: do something
}

func Leader(ctx context.Context, termID int64) {
	// TODO: do something
}
