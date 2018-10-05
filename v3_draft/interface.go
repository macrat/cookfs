package main

import (
	"context"
	"net/url"
)

type Request struct {
	Node *url.URL
	Path string
	Data interface{}
}

type Response struct {
	StatusCode int
	Data       interface{}
}

type CommunicationHandler interface {
	Listen(context.Context, Follower)
	Send(context.Context, Request) Response
}

type AliveMessage struct {
	Leader  *url.URL `json:"leader"`
	Term    int64    `json:"term"`
	PatchID PatchID  `json:"patch_id"`
}
