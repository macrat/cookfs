package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/vmihailenco/msgpack"
)

type HTTPHandler http.Client

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

func processSet(f Follower, path string, body io.Reader) Response {
	data := NewRequestStruct(path)
	if data != nil {
		return Response{StatusCode: 404}
	}

	if err := msgpack.NewDecoder(body).Decode(&data); err != nil {
		return Response{StatusCode: 404}
	}

	switch path {
	case "/term":
		return f.AliveMessage(data.(AliveMessage))
	}
	return Response{StatusCode: 404}
}

func processGet(f Follower, path string) Response {
	return Response{StatusCode: 500, Data: "this is test response"}
}

func newMux(ctx context.Context, f Follower) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var response Response

		if r.Method == "POST" {
			response = processSet(f, r.URL.Path, r.Body)
			r.Body.Close()
		} else {
			response = processGet(f, r.URL.Path)
		}

		data, err := msgpack.Marshal(response.Data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/x-msgpack")
		w.WriteHeader(response.StatusCode)
		w.Write(data)
	})

	return mux
}

func (h *HTTPHandler) Listen(ctx context.Context, f Follower) {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: newMux(ctx, f),
	}

	go srv.ListenAndServe()

	<-ctx.Done()
	srv.Shutdown(ctx)
}

func (h *HTTPHandler) Send(ctx context.Context, host *url.URL, req Request) Response {
	u := *host
	u.Path = u.Path + req.Path

	var request *http.Request
	var err error
	if req.Data == nil {
		request, err = http.NewRequest("GET", (&u).String(), nil)
	} else {
		data, err := msgpack.Marshal(req.Data)
		if err != nil {
			return Response{StatusCode: 400}
		}
		request, err = http.NewRequest("POST", (&u).String(), bytes.NewReader(data))
	}
	if err != nil {
		return Response{StatusCode: 400}
	}

	response, err := (*http.Client)(h).Do(request.WithContext(ctx))
	if err != nil {
		return Response{StatusCode: 400}
	}

	data, err := msgpack.NewDecoder(response.Body).DecodeInterface()
	if err != nil {
		return Response{StatusCode: 400}
	}
	return Response{response.StatusCode, data}
}
