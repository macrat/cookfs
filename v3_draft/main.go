package main

import (
	"net/url"
	"sync"
	"context"
)

type Request struct {
	Path string
	Data interface{}
}

type Response struct {
	StatusCode int
	Data       interface{}
}

type RequestResponse struct {
	Request  Request
	Response chan Response
}

type Follower struct {
	Leader *url.URL
	TermID int64
	Staet *State
}

func (f Follower) handle(request Request) Response {
}

func (f Follower) Run(ctx context.Context, ch chan RequestResponse) {
	for {
		switch {
		case reqres := <-ch:
			reqres.Response <-f.handle(reqres.Request)

		case <-ctx.Done():
			return
	}
}

func Candidacy(ctx context.Context, termID int64) {
	// TODO: do something
}

func Leader(ctx context.Context, termID int64) {
	// TODO: do something
}
