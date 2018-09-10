package main

import (
	"io"
	"net/http"
	"os"
	"path"
)

const (
	CHUNK_SIZE = 64
)

type ChunkIOError struct {
	code    int
	message string
}

func (c *ChunkIOError) Error() string {
	return c.message
}

func (c *ChunkIOError) Code() int {
	return c.code
}

type ChunkHandler struct {
	BasePath string
}

func (c ChunkHandler) Read(hash string) (io.Reader, *ChunkIOError) {
	if f, err := os.Open(path.Join(c.BasePath, hash)); err != nil {
		return nil, &ChunkIOError{http.StatusNotFound, err.Error()}
	} else {
		return f, nil
	}
}

func (c ChunkHandler) Create(hash string) (io.Writer, *ChunkIOError) {
	if f, err := os.Create(path.Join(c.BasePath, hash)); err != nil {
		return nil, &ChunkIOError{http.StatusInternalServerError, err.Error()}
	} else {
		return f, nil
	}
}

func (c ChunkHandler) Delete(hash string) *ChunkIOError {
	if err := os.Remove(path.Join(c.BasePath, hash)); err != nil {
		return &ChunkIOError{http.StatusInternalServerError, err.Error()}
	} else {
		return nil
	}
}

func (c ChunkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	println(r.URL.Path)
	hash := r.URL.Path[len("/chunk/"):]

	switch r.Method {
	case "GET":
		f, err := c.Read(hash)
		if err != nil {
			w.WriteHeader(err.Code())
		} else {
			w.Header().Add("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			io.Copy(w, f)
		}

	case "PUT":
		data := make([]byte, CHUNK_SIZE + 1)

		if size, err := r.Body.Read(data); err != nil || size != CHUNK_SIZE {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		f, err := c.Create(hash)
		if err != nil {
			w.WriteHeader(err.Code())
		} else if s, e := f.Write(data); e != nil || s != CHUNK_SIZE {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusCreated)
			return
		}

	case "DELETE":
		if err := c.Delete(hash); err != nil {
			w.WriteHeader(err.Code())
		} else {
			w.WriteHeader(http.StatusNoContent)
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func main() {
	/*
	"/master/staus" GET/POST
	"/master/poll_request" POST
	"/status" GET/POST
	"/recipie" POST
	"/chunk/<[0-9a-f]+>" GET/PUT
	*/

	http.Handle("/chunk/", ChunkHandler{"chunks/"})
	http.ListenAndServe(":8080", nil)
}
