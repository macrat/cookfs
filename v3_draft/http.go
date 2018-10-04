package main

import (
	"bytes"
	"context"
	"net/http"
	"net/url"

	"github.com/vmihailenco/msgpack"
)

type HTTPHandler http.Client

func newMux(ctx context.Context, ch chan RequestResponse) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		req := Request{Path: r.URL.Path}

		if r.Method == "POST" {
			data := NewRequestStruct(req.Path)
			if data != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			err := msgpack.NewDecoder(r.Body).Decode(&data)
			r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			req.Data = data
		}

		res := make(chan Response)
		ch <- RequestResponse{ctx, req, res}

		response := <-res
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

func (h *HTTPHandler) Listen(ctx context.Context, ch chan RequestResponse) {
	srv := &http.Server{
		Addr: ":8080",
		Handler: newMux(ctx, ch),
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
