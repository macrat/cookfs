package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type HTTPPolling struct {
	Polling *Polling

	self  *Node
	Nodes []*Node
}

func NewHTTPPolling(self *Node, nodes []*Node) HTTPPolling {
	p := HTTPPolling{self: self, Nodes: nodes}

	p.Polling = NewPolling(p)

	return p
}

func (p HTTPPolling) Self() *Node {
	return p.self
}

func (p HTTPPolling) SendAlive(term Term) {
	wg := &sync.WaitGroup{}

	for _, n := range p.Nodes {
		if n.Equals(p.self) {
			continue
		}

		wg.Add(1)
		go func(node *Node) {
			defer wg.Done()

			resp, err := node.Post("/alive", term)
			if err != nil {
				fmt.Printf("%s: failed to send alive to %s\n", p.self, node)
			} else if resp.StatusCode != http.StatusNoContent {
				fmt.Printf("%s: denied alive by %s: %s\n", p.self, node, resp.Status)
			}
		}(n)
	}

	wg.Wait()
}

func (p HTTPPolling) PollRequest(term Term) bool {
	response := make(chan bool)
	defer close(response)

	for _, n := range p.Nodes {
		go func(node *Node) {
			resp, err := node.Post("/poll", term)
			defer func() {
				recover()
			}()
			response <- err == nil && resp.StatusCode == http.StatusNoContent
		}(n)
	}

	score := 0
	for range p.Nodes {
		if <-response {
			score++
		}

		if float64(score) >= float64(len(p.Nodes))/2.0 {
			return true
		}
	}

	return false
}

func (p HTTPPolling) ServeAlive(w http.ResponseWriter, r *http.Request) {
	var term Term
	if err := json.NewDecoder(r.Body).Decode(&term); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if p.Polling.AliveArrived(term) {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusConflict)
	}
}

func (p HTTPPolling) ServePoll(w http.ResponseWriter, r *http.Request) {
	var term Term
	if err := json.NewDecoder(r.Body).Decode(&term); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if p.Polling.CanPoll(term) {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusConflict)
	}
}

func (p HTTPPolling) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" && r.URL.Path == "/" {
		x, err := json.Marshal(p.Polling.CurrentTerm())
		if err != nil {
			fmt.Fprintln(w, err.Error())
		} else {
			w.Write(x)
		}
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	switch r.URL.Path {
	case "/alive":
		p.ServeAlive(w, r)

	case "/poll":
		p.ServePoll(w, r)

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (p HTTPPolling) Start() error {
	return p.Polling.Start()
}

func (p HTTPPolling) Stop() error {
	return p.Polling.Stop()
}
