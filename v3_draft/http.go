package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/vmihailenco/msgpack"
)

type HTTPHandler http.Client

func processSet(c *CookFS, path string, body io.Reader) Response {
	data := NewRequestStruct(path)
	if data == nil {
		return Response{StatusCode: 404}
	}

	if err := msgpack.NewDecoder(body).Decode(data); err != nil {
		return Response{StatusCode: 404}
	}

	return c.HandleRequest(Request{c.Nodes()[0], path, data})
}

func processGet(c *CookFS, path string) Response {
	return c.HandleRequest(Request{c.Nodes()[0], path, nil})
}

func newMux(ctx context.Context, c *CookFS) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var response Response

		if r.Method == "POST" {
			response = processSet(c, r.URL.Path, r.Body)
			r.Body.Close()
		} else {
			response = processGet(c, r.URL.Path)
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

func (h *HTTPHandler) Listen(ctx context.Context, node *Node, c *CookFS) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", node.Port()),
		Handler: newMux(ctx, c),
	}

	go srv.ListenAndServe()

	<-ctx.Done()
	srv.Shutdown(ctx)
}

func (h *HTTPHandler) Send(ctx context.Context, req Request) Response {
	u := *req.Node
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
