package plugins

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/vmihailenco/msgpack"

	"github.com/macrat/cookfs/cooklib"
)

type HTTPHandler http.Client

func processSet(c *cooklib.CookFS, path string, body io.Reader) cooklib.Response {
	data := cooklib.NewRequestStruct(path)
	if data == nil {
		return cooklib.Response{StatusCode: 404}
	}

	if err := msgpack.NewDecoder(body).Decode(data); err != nil {
		return cooklib.Response{StatusCode: 404}
	}

	return c.HandleRequest(cooklib.Request{c.Nodes()[0], path, data, 0})
}

func processGet(c *cooklib.CookFS, path string) cooklib.Response {
	return c.HandleRequest(cooklib.Request{c.Nodes()[0], path, nil, 0})
}

func newMux(ctx context.Context, c *cooklib.CookFS) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var response cooklib.Response

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

func (h *HTTPHandler) Listen(ctx context.Context, node *cooklib.Node, c *cooklib.CookFS) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", node.Port()),
		Handler: newMux(ctx, c),
	}

	go srv.ListenAndServe()

	<-ctx.Done()
	srv.Shutdown(ctx)
}

func (h *HTTPHandler) Send(ctx context.Context, req cooklib.Request) cooklib.Response {
	u := *req.Node
	u.Path = u.Path + req.Path

	var request *http.Request
	var err error
	if req.Data == nil {
		request, err = http.NewRequest("GET", (&u).String(), nil)
	} else {
		data, err := msgpack.Marshal(req.Data)
		if err != nil {
			return cooklib.Response{StatusCode: 400}
		}
		request, err = http.NewRequest("POST", (&u).String(), bytes.NewReader(data))
	}
	if err != nil {
		return cooklib.Response{StatusCode: 400}
	}

	response, err := (*http.Client)(h).Do(request.WithContext(ctx))
	if err != nil {
		return cooklib.Response{StatusCode: 400}
	}

	data, err := msgpack.NewDecoder(response.Body).DecodeInterface()
	if err != nil {
		return cooklib.Response{StatusCode: 400}
	}
	return cooklib.Response{response.StatusCode, data}
}
