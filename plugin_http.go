package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func post(n *Node, endpoint string, timeout time.Duration, data []byte) (*http.Response, error) {
	return (&http.Client{Timeout: timeout}).Post(n.Join(endpoint).String(), "application/json", bytes.NewReader(data))
}

type HTTPTransmitPlugin struct {
	discover DiscoverPlugin
}

func (ht *HTTPTransmitPlugin) Bind(cook *CookFS) {
	ht.discover = cook.Discover
}

func (ht *HTTPTransmitPlugin) SendAlive(term TermStatus) {
	nodes := ht.discover.Nodes()

	data, _ := json.Marshal(term)

	for _, n := range nodes {
		if n.Equals(term.Leader) {
			continue
		}

		go func(node *Node) {
			resp, err := post(node, "/leader/alive", 500*time.Millisecond, data)
			if err != nil {
				fmt.Printf("%s: failed to send alive to %s: %s\n", term.Leader, node, err.Error())
			} else if resp.StatusCode != http.StatusNoContent {
				fmt.Printf("%s: denied alive by %s: %s\n", term.Leader, node, resp.Status)
			}
		}(n)
	}
}

func (ht *HTTPTransmitPlugin) polling(endpoint string, data interface{}) bool {
	nodes := ht.discover.Nodes()

	json_data, _ := json.Marshal(data)

	response := make(chan bool)
	defer close(response)

	for _, n := range nodes {
		go func(node *Node) {
			resp, err := post(node, endpoint, 200*time.Millisecond, json_data)
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

func (ht *HTTPTransmitPlugin) PollRequest(term TermStatus) bool {
	return ht.polling("/leader/poll", term)
}

func (ht *HTTPTransmitPlugin) AddJournalEntry(entry *JournalEntry) bool {
	return ht.polling("/journal", entry)
}

func (ht *HTTPTransmitPlugin) CommitJournal(id Hash) bool {
	return ht.polling("/journal/commit", id)
}

func (ht *HTTPTransmitPlugin) Run(stop chan struct{}) error {
	return nil
}

type HTTPReceivePlugin struct {
	self    *Node
	polling *Polling
	journal *Journal
	recipe  RecipePlugin

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
	hr.mux.HandleFunc("/recipe/{tag:.+}", hr.serveRecipePut).Methods("PUT")
	hr.mux.HandleFunc("/recipe/{tag:.+}", hr.serveRecipeGet).Methods("GET")
	hr.mux.HandleFunc("/recipe/", hr.serveRecipeList).Methods("GET")
	hr.mux.HandleFunc("/recipe/{prefix:.+}/", hr.serveRecipeList).Methods("GET")

	return hr
}

func (hr *HTTPReceivePlugin) Bind(cook *CookFS) {
	hr.self = cook.Discover.Self()
	hr.polling = cook.Polling
	hr.journal = cook.Journal
	hr.recipe = cook.Recipe
}

func (hr *HTTPReceivePlugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hr.mux.ServeHTTP(w, r)
}

func (hr *HTTPReceivePlugin) serveAlive(w http.ResponseWriter, r *http.Request) {
	var term TermStatus
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
	var term TermStatus
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
	x, err := json.Marshal(hr.journal.GetLog())
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

func (hr *HTTPReceivePlugin) serveRecipePut(w http.ResponseWriter, r *http.Request) {
	tag := "/" + mux.Vars(r)["tag"]

	if !hr.polling.IsLeader() {
		http.Redirect(w, r, hr.polling.CurrentTerm().Leader.Join("/recipe"+tag).String(), http.StatusSeeOther)
		return
	}

	var recipe Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := hr.journal.AddRecipe(tag, recipe); err != nil {
		fmt.Printf("%s: failed to add new recipe: %s: %s\n", hr.self, tag, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		fmt.Printf("%s: committed new recipe: %s\n", hr.self, tag)
		w.WriteHeader(http.StatusNoContent)
	}
}

func (hr *HTTPReceivePlugin) serveRecipeGet(w http.ResponseWriter, r *http.Request) {
	tag := "/" + mux.Vars(r)["tag"]

	recipe, err := hr.recipe.Load(tag)
	if err == RecipeNotFound {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(recipe)
}

func (hr *HTTPReceivePlugin) serveRecipeList(w http.ResponseWriter, r *http.Request) {
	prefix := mux.Vars(r)["prefix"]
	if prefix != "" {
		prefix = "/" + prefix + "/"
	}

	recipes, err := hr.recipe.Find(prefix)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if recipes == nil {
		w.Write([]byte("[]"))
	} else {
		json.NewEncoder(w).Encode(recipes)
	}
}

func (hr *HTTPReceivePlugin) Run(stop chan struct{}) error {
	go http.ListenAndServe(fmt.Sprintf(":%d", hr.self.Port()), hr)
	return nil
}
