package main

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	CHUNK_SIZE         = 64
	LEADER_DEATH_TIMER = 1 * time.Second
	POLL_TIMEOUT       = 100 * time.Millisecond
)

func checkHash(hash string, data []byte) bool {
	except, err := hex.DecodeString(hash)
	if err != nil || len(except) != sha512.Size {
		return false
	}

	sum := sha512.Sum512(data)

	for i := range except {
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

func (c *ChunkHandler) Read(hash string) (io.ReadCloser, *ChunkIOError) {
	if f, err := os.Open(path.Join(c.BasePath, hash)); err != nil {
		return nil, &ChunkIOError{http.StatusNotFound, err.Error()}
	} else {
		return f, nil
	}
}

func (c *ChunkHandler) Create(hash string) (io.WriteCloser, *ChunkIOError) {
	if _, err := os.Stat(path.Join(c.BasePath, hash)); err == nil {
		return nil, &ChunkIOError{http.StatusNoContent, "Already exists"}
	}

	if f, err := os.Create(path.Join(c.BasePath, hash)); err != nil {
		return nil, &ChunkIOError{http.StatusInternalServerError, err.Error()}
	} else {
		return f, nil
	}
}

func (c *ChunkHandler) Delete(hash string) *ChunkIOError {
	if err := os.Remove(path.Join(c.BasePath, hash)); err != nil {
		return &ChunkIOError{http.StatusInternalServerError, err.Error()}
	} else {
		return nil
	}
}

func (c *ChunkHandler) ServeList(w http.ResponseWriter, r *http.Request) {
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

func (c *ChunkHandler) ServeGET(hash string, w http.ResponseWriter, r *http.Request) {
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

func (c *ChunkHandler) ServePUT(hash string, w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Length", "0")

	data := make([]byte, CHUNK_SIZE+1)

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
		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *ChunkHandler) ServeDELETE(hash string, w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Length", "0")

	if err := c.Delete(hash); err != nil {
		w.WriteHeader(err.Code())
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *ChunkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/chunk" {
		if r.Method == "GET" {
			c.ServeList(w, r)
		} else {
			w.Header().Add("Content-Length", "0")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	} else if len(r.URL.Path) < len("/chunk/")+sha512.Size*2 {
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

type Leader struct {
	URL         string    `json:"url"`
	Term        int64     `json:"term"`
	LastContact time.Time `json:"last_contact"`
}

func (l Leader) IsAlive() bool {
	return time.Now().Sub(l.LastContact) <= LEADER_DEATH_TIMER
}

type PollInfo struct {
	Term      int64
	Timestamp time.Time
}

type LeaderHandler struct {
	sync.Mutex

	leader Leader
	poll   PollInfo
}

func (l *LeaderHandler) HandleInfo(w http.ResponseWriter, r *http.Request) {
	info, err := json.Marshal(l.leader)
	if err != nil {
		w.Header().Add("Content-Length", "0")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Length", strconv.Itoa(len(info)))
	w.WriteHeader(http.StatusOK)
	w.Write(info)
}

func (l *LeaderHandler) HandleAlive(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Length", "0")

	var req struct {
		URL  string `json:"url"`
		Term int64  `json:"term"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if l.leader.Term > req.Term && (l.leader.Term != req.Term || l.leader.URL != req.URL) {
		w.WriteHeader(http.StatusConflict)
		return
	}

	l.Lock()
	defer l.Unlock()

	l.leader = Leader{
		URL:         req.URL,
		Term:        req.Term,
		LastContact: time.Now(),
	}

	w.WriteHeader(http.StatusNoContent)
}

func (l *LeaderHandler) HandlePollRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Length", "0")

	var req struct {
		Term int64 `json:"term"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	l.Lock()
	defer l.Unlock()

	if l.leader.Term < req.Term && (l.poll.Term < req.Term || l.poll.Timestamp.Add(POLL_TIMEOUT).Before(time.Now())) {
		l.poll.Term = req.Term
		l.poll.Timestamp = time.Now()

		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

func (l *LeaderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/leader/poll_request" {
		if r.Method == "POST" {
			l.HandlePollRequest(w, r)
		} else {
			w.Header().Add("Content-Length", "0")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	switch r.Method {
	case "GET":
		l.HandleInfo(w, r)

	case "POST":
		l.HandleAlive(w, r)

	default:
		w.Header().Add("Content-Length", "0")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type Mux struct {
	httpServer *http.ServeMux
}

func NewMux(chunk *ChunkHandler, leader *LeaderHandler) *Mux {
	m := http.NewServeMux()

	m.Handle("/chunk", chunk)
	m.Handle("/chunk/", chunk)
	m.Handle("/leader", leader)
	m.Handle("/leader/poll_request", leader)

	return &Mux{m}
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("timestamp:%s method:%s path:\"%s\"\n", time.Now(), r.Method, r.URL.Path)

	if strings.HasSuffix(r.URL.Path, "/") {
		http.Redirect(w, r, r.URL.Path[:len(r.URL.Path)-1], http.StatusMovedPermanently)
		return
	}

	m.httpServer.ServeHTTP(w, r)
}

func main() {
	/* TODO
	"/health" GET/POST
	"/leader" GET redirect
	"/leader/alive" POST
	"/leader/request" POST
	"/recipe" GET/POST
	"/recipe/<path/to/recipe>" GET
	"/recipe/journal" GET/POST
	"/recipe/journal/commit" PUT
	"/chunk/backorder" GET/POST
	"/metrics" GET
	*/

	http.ListenAndServe(":8080", NewMux(&ChunkHandler{"chunks/"}, &LeaderHandler{}))
}
