package main

import (
	"context"
	"net/url"
	"time"
)

type Follower struct {
	Leader *url.URL
	Term   int64
	State  *State

	alive chan struct{}
}

func NewFollower() Follower {
	return Follower{alive: make(chan struct{})}
}

func (f Follower) AliveMessage(alive AliveMessage) Response {
	if f.Leader == alive.Leader && f.Term == alive.Term || f.Term < alive.Term {
		f.Leader = alive.Leader
		f.Term = alive.Term
		f.alive <- struct{}{}

		return Response{StatusCode: 200}
	} else {
		return Response{StatusCode: 409}
	}
}

func (f Follower) Run(ctx context.Context) {
	for {
		select {
		case <-f.alive:
			continue

		case <-time.After(500 * time.Millisecond):
			c, _ := context.WithTimeout(ctx, 1*time.Second)
			go Candidacy(c, f.Term+1)

		case <-ctx.Done():
			return
		}
	}
}

func Candidacy(ctx context.Context, term int64) {
	println("been candidacy")

	req, resp := StartWorkers(ctx)

	for {
		select {
		case r := <-resp:
			if r.StatusCode == 200 {
				poll++
			}

		case <-ctx.Done():
			return
		}
	}
}

func Leader(ctx context.Context, term int64) {
	// TODO: do something
}

func CommunicateWorker(ctx context.Context, req chan Request, resp chan Response) {
	for {
		select {
		case r := <-req:
			resp <- handler.Send(ctx, req)

		case <-ctx.Done():
			return
		}
	}
}

func StartWorkers(ctx context.Context) (chan Request, chan Response) {
	req := make(chan Request)
	resp := make(chan Response)

	for i := 0; i < 10; i++ {
		go CommunicateWorker(ctx, req, resp)
	}
}
