package main

import (
	"context"
	"net/url"
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
	Context  context.Context
	Request  Request
	Response chan Response
}

type CommunicationHandler interface {
	Listen(context.Context, chan RequestResponse)
	Send(context.Context, *url.URL, Request) Response
}

type AliveMessage struct {
	Leader  *url.URL `json:"leader"`
	Term    int64    `json:"term"`
	PatchID PatchID  `json:"patch_id"`
}

func NewRequestStruct(path string) interface{} {
	switch path {
	case "/term":
		return AliveMessage{}

	case "/journal":
		return Patch{}

	default:
		return nil
	}
}
