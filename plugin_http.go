package plugin.bundled

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type HTTPTransmitPlugin struct {
	discover DiscoverPlugin
}

func (ht *HTTPTransmitPlugin) Bind(cook *CookFS) {
	ht.discover = cook.Discover
}

func (ht *HTTPTransmitPlugin) SendAlive(term Term) {
	wg := &sync.WaitGroup{}

	nodes := ht.discover.Nodes()

	for _, n := range nodes {
		if n.Equals(term.Leader) {
			continue
		}

		wg.Add(1)
		go func(node *Node) {
			defer wg.Done()

			resp, err := node.Post("/alive", term)
			if err != nil {
				fmt.Printf("%s: failed to send alive to %s\n", term.Leader, node)
			} else if resp.StatusCode != http.StatusNoContent {
				fmt.Printf("%s: denied alive by %s: %s\n", term.Leader, node, resp.Status)
			}
		}(n)
	}

	wg.Wait()
}

func (ht *HTTPTransmitPlugin) PollRequest(term Term) bool {
	nodes := ht.discover.Nodes()

	response := make(chan bool)
	defer close(response)

	for _, n := range nodes {
		go func(node *Node) {
			resp, err := node.Post("/poll", term)
			defer func() {
				recover()
			}()
			response <- err == nil && resp.StatusCode == http.StatusNoContent
		}(n)
	}

	score := 0
	for range nodes {
		if <-response {
			score++
		}

		if float64(score) >= float64(len(nodes))/2.0 {
			return true
		}
	}

	return false
}

func (ht *HTTPTransmitPlugin) Run(stop chan struct{}) error {
	return nil
}

type HTTPReceivePlugin struct {
	self    *Node
	polling *Polling
}

func (hr *HTTPReceivePlugin) Bind(cook *CookFS) {
	hr.self = cook.Discover.Self()
	hr.polling = cook.Polling
}

func (hr *HTTPReceivePlugin) serveAlive(w http.ResponseWriter, r *http.Request) {
	var term Term
	if err := json.NewDecoder(r.Body).Decode(&term); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if hr.polling.AliveArrived(term) {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusConflict)
	}
}

func (hr *HTTPReceivePlugin) servePoll(w http.ResponseWriter, r *http.Request) {
	var term Term
	if err := json.NewDecoder(r.Body).Decode(&term); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if hr.polling.CanPoll(term) {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusConflict)
	}
}

func (hr *HTTPReceivePlugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" && r.URL.Path == "/" {
		x, err := json.Marshal(hr.polling.CurrentTerm())
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
		hr.serveAlive(w, r)

	case "/poll":
		hr.servePoll(w, r)

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (hr *HTTPReceivePlugin) Run(stop chan struct{}) error {
	go http.ListenAndServe(fmt.Sprintf(":%d", hr.self.Port()), hr)
	return nil
}
