package cooklib

import (
	"context"
)

type Request struct {
	Node *Node
	Path string
	Data interface{}
}

type Response struct {
	StatusCode int
	Data       interface{}
}

type CommunicationHandler interface {
	Listen(context.Context, *Node, *CookFS)
	Send(context.Context, Request) Response
}

type AliveMessage struct {
	Leader  *Node   `json:"leader"`
	Term    int64   `json:"term"`
	PatchID PatchID `json:"patch_id"`
}

type PollRequest struct {
	Node    *Node   `json:"node"`
	Term    int64   `json:"term"`
	PatchID PatchID `json:"patch_id"`
}

func NewRequestStruct(path string) interface{} {
	switch path {
	case "/term":
		return &AliveMessage{}

	case "/term/poll":
		return &PollRequest{}

	case "/journal":
		return &Patch{}

	default:
		return nil
	}
}
