package main

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
)

const (
	CHUNK_SIZE = 64
)

func checkHash(hash string, data []byte) bool {
	except, err := hex.DecodeString(hash)
	if err != nil || len(except) != sha512.Size {
		return false
	}

	sum := sha512.Sum512(data)

	for i, _ := range except {
		if except[i] != sum[i] {
			return false
		}
	}

	return true
}

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

func (c ChunkHandler) Read(hash string) (io.ReadCloser, *ChunkIOError) {
	if f, err := os.Open(path.Join(c.BasePath, hash)); err != nil {
		return nil, &ChunkIOError{http.StatusNotFound, err.Error()}
	} else {
		return f, nil
	}
}

func (c ChunkHandler) Create(hash string) (io.WriteCloser, *ChunkIOError) {
	if _, err := os.Stat(path.Join(c.BasePath, hash)); err == nil {
		return nil, &ChunkIOError{http.StatusOK, "Already exists"}
	}

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

func (c ChunkHandler) ServeList(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir(c.BasePath)
	if err != nil {
		w.Header().Add("Content-Length", "0")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	chunks := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			chunks = append(chunks, file.Name())
		}
	}

	bytes, err := json.Marshal(chunks)
	if err != nil {
		w.Header().Add("Content-Length", "0")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Length", strconv.Itoa(len(bytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func (c ChunkHandler) ServeGET(hash string, w http.ResponseWriter, r *http.Request) {
	f, err := c.Read(hash)
	if err != nil {
		w.Header().Add("Content-Length", "0")
		w.WriteHeader(err.Code())
		return
	}
	defer f.Close()

	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Length", strconv.Itoa(CHUNK_SIZE))
	w.WriteHeader(http.StatusOK)
	_, e := io.Copy(w, f)
	if e != nil {
		println(e.Error())
	}
}

func (c ChunkHandler) ServePUT(hash string, w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Length", "0")

	data := make([]byte, CHUNK_SIZE + 1)

	if size, err := r.Body.Read(data); err != io.EOF || size != CHUNK_SIZE {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !checkHash(hash, data[:CHUNK_SIZE]) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	f, err := c.Create(hash)
	if err != nil {
		w.WriteHeader(err.Code())
		return
	}
	defer f.Close()

	if s, e := f.Write(data[:CHUNK_SIZE]); e != nil || s != CHUNK_SIZE {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (c ChunkHandler) ServeDELETE(hash string, w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Length", "0")

	if err := c.Delete(hash); err != nil {
		w.WriteHeader(err.Code())
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (c ChunkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/chunk/" {
		if r.Method == "GET" {
			c.ServeList(w, r)
		} else {
			w.Header().Add("Content-Length", "0")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	} else if len(r.URL.Path) < len("/chunk/") + sha512.Size*2 {
		w.Header().Add("Content-Length", "0")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	hash := r.URL.Path[len("/chunk/"):]

	switch r.Method {
	case "GET":
		c.ServeGET(hash, w, r)

	case "PUT":
		c.ServePUT(hash, w, r)

	case "DELETE":
		c.ServeDELETE(hash, w, r)

	default:
		w.Header().Add("Content-Length", "0")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	/* TODO
	"/health" GET/POST
	"/leader" GET redirect
	"/leader/request" POST
	"/recipe" GET/POST
	"/recipe/<path/to/recipe>" GET
	"/recipe/journal" GET/POST
	"/recipe/journal/commit" PUT
	"/chunk/backorder" GET/POST
	"/metrics" GET
	*/

	http.Handle("/chunk/", ChunkHandler{"chunks/"})
	http.ListenAndServe(":8080", nil)
}
