package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
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

			resp, err := node.Post("/leader/alive", term)
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
			resp, err := node.Post("/leader/poll", term)
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
	journal *JournalManager

	mux *mux.Router
}

func NewHTTPReceivePlugin() *HTTPReceivePlugin {
	hr := &HTTPReceivePlugin{mux: mux.NewRouter()}

	hr.mux.HandleFunc("/leader", hr.serveLeader).Methods("GET")
	hr.mux.HandleFunc("/leader/alive", hr.serveAlive).Methods("POST")
	hr.mux.HandleFunc("/leader/poll", hr.servePoll).Methods("POST")
	hr.mux.HandleFunc("/journal", hr.serveJournalList).Methods("GET")
	hr.mux.HandleFunc("/journal", hr.serveJournalAdd).Methods("POST")
	hr.mux.HandleFunc("/journal/commit", hr.serveJournalCommit).Methods("POST")

	return hr
}

func (hr *HTTPReceivePlugin) Bind(cook *CookFS) {
	hr.self = cook.Discover.Self()
	hr.polling = cook.Polling
	hr.journal = cook.Journal
}

func (hr *HTTPReceivePlugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hr.mux.ServeHTTP(w, r)
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

func (hr *HTTPReceivePlugin) serveLeader(w http.ResponseWriter, r *http.Request) {
	x, err := json.Marshal(hr.polling.CurrentTerm())
	if err != nil {
		fmt.Fprintln(w, err.Error())
	} else {
		w.Write(x)
	}
}

func (hr *HTTPReceivePlugin) isLeader(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}

	ips, err := net.LookupHost(hr.polling.CurrentTerm().Leader.Hostname())
	if err != nil {
		return false
	}

	for _, ip := range ips {
		if host == ip {
			return true
		}
	}

	return false
}

func (hr *HTTPReceivePlugin) serveJournalList(w http.ResponseWriter, r *http.Request) {
	list := struct {
		Committed []*JournalEntry `json:"committed"`
		Dirty     []*JournalEntry `json:"dirty"`
	}{
		Committed: hr.journal.GetCommitted(20),
		Dirty: hr.journal.GetDirty(),
	}

	x, err := json.Marshal(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.Write(x)
	}
}

func (hr *HTTPReceivePlugin) serveJournalAdd(w http.ResponseWriter, r *http.Request) {
	if !hr.isLeader(r) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	entry := &JournalEntry{}
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := hr.journal.AddEntry(entry); err != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (hr *HTTPReceivePlugin) serveJournalCommit(w http.ResponseWriter, r *http.Request) {
	if !hr.isLeader(r) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var id Hash
	if err := json.NewDecoder(r.Body).Decode(&id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := hr.journal.Commit(id); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (hr *HTTPReceivePlugin) Run(stop chan struct{}) error {
	go http.ListenAndServe(fmt.Sprintf(":%d", hr.self.Port()), hr)
	return nil
}
