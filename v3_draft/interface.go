package main

import (
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
	Context  context.Context
	Request  Request
	Response chan Response
}

type CommunicationHandler interface {
	Listen(context.Context, chan RequestResponse)
	Send(context.Context, *url.URL, Request) Response
}
