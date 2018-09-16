package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

type Node url.URL

func ForceNewNode(raw_url string) *Node {
	u, err := url.Parse(raw_url)
	if err != nil {
		panic(err.Error())
	}
	return (*Node)(u)
}

func (n *Node) String() string {
	return (*url.URL)(n).String()
}

func (n *Node) Port() int {
	port, err := strconv.Atoi((*url.URL)(n).Port())
	if err != nil {
		return 80
	}
	return port
}

func (n *Node) Equals(another *Node) bool {
	return *n == *another
}

func (n *Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

func (n *Node) UnmarshalJSON(raw []byte) error {
	var x string
	if err := json.Unmarshal(raw, &x); err != nil {
		return err
	}

	u, err := url.Parse(x)
	if err != nil {
		return err
	}

	*n = (Node)(*u)
	return nil
}

func (n *Node) Join(subpath string) *Node {
	u := *n
	u.Path = path.Join(u.Path, subpath)
	return &u
}

func (n *Node) Get(endpoint string) (*http.Response, error) {
	return http.Get(n.Join(endpoint).String())
}

func (n *Node) Post(endpoint string, data interface{}) (*http.Response, error) {
	x, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return (&http.Client{Timeout: 100 * time.Millisecond}).Post(n.Join(endpoint).String(), "application/json", bytes.NewReader(x))
}

type Term struct {
	ID     int64 `json:"id"`
	Leader *Node `json:"leader"`
}

func (t Term) Equals(another Term) bool {
	return t.ID == another.ID && t.Leader.String() == another.Leader.String()
}

func (t Term) NewerThan(another Term) bool {
	return t.ID > another.ID
}

func (t Term) OlderThan(another Term) bool {
	return t.ID < another.ID
}

type Polling struct {
	sync.Mutex

	Self *Node

	SendAlive   func(Term)
	PollRequest func(Term) bool

	PollingInterval     time.Duration
	SendAliveInterval   time.Duration
	LeaderDeathTimerMin time.Duration
	LeaderDeathTimerMax time.Duration

	currentTerm  Term
	aliveArrived chan struct{}
	lastPoll     time.Time

	currentLoop chan struct{}
}

func NewPolling(url *Node, sendAlive func(Term), pollRequest func(Term) bool) *Polling {
	return &Polling{
		Self:                url,
		SendAlive:           sendAlive,
		PollRequest:         pollRequest,
		PollingInterval:     1 * time.Second,
		SendAliveInterval:   100 * time.Millisecond,
		LeaderDeathTimerMin: 500 * time.Millisecond,
		LeaderDeathTimerMax: 700 * time.Millisecond,
		aliveArrived:        make(chan struct{}),
	}
}

func (p *Polling) StartPoll() {
	newTerm := Term{
		ID:     p.currentTerm.ID + 1,
		Leader: p.Self,
	}

	if p.PollRequest(newTerm) {
		p.Lock()
		defer p.Unlock()

		fmt.Printf("%s: I'm leader of %d\n", p.Self, newTerm.ID)

		p.currentTerm = newTerm
		p.StartLeaderLoop()
	} else {
		fmt.Printf("%s: failed to promotion to leader of %d\n", p.Self, newTerm.ID)
	}
}

func (p *Polling) AliveArrived(term Term) (accepted bool) {
	if !term.NewerThan(p.currentTerm) && !term.Equals(p.currentTerm) {
		fmt.Printf("denied: %s(%d) -> %s(%d)\n", p.currentTerm.Leader, p.currentTerm.ID, term.Leader, term.ID)
		return false
	}

	if !term.Equals(p.currentTerm) {
		fmt.Printf("%s: leader changed to %s at term %d\n", p.Self, term.Leader, term.ID)
	}

	p.Lock()
	defer p.Unlock()

	p.currentTerm = term
	p.aliveArrived <- struct{}{}

	return true
}

func (p *Polling) CanPoll(term Term) bool {
	if term.NewerThan(p.currentTerm) && p.lastPoll.Add(p.PollingInterval).Before(time.Now()) {
		p.Lock()
		defer p.Unlock()

		p.lastPoll = time.Now()

		fmt.Printf("%s: vote to %s\n", p.Self, term.Leader)

		return true
	} else {
		return false
	}
}

func (p *Polling) StartLeaderLoop() {
	stop := make(chan struct{})

	if p.currentLoop != nil {
		close(p.currentLoop)
	}
	p.currentLoop = stop

	go func() {
		for {
			select {
			case <-p.aliveArrived:
				fmt.Printf("%s: relegation to follower because alive arrived from %s\n", p.Self, p.currentTerm.Leader)
				p.StartFollowerLoop()
				return

			case <-time.After(p.SendAliveInterval):
				p.SendAlive(p.currentTerm)

			case <-stop:
				return
			}
		}
	}()
}

func (p *Polling) LeaderDeathTimer() time.Duration {
	return time.Duration(rand.Int63n(int64(p.LeaderDeathTimerMax-p.LeaderDeathTimerMin))) + p.LeaderDeathTimerMin
}

func (p *Polling) StartFollowerLoop() {
	stop := make(chan struct{})

	rand.Seed(time.Now().UnixNano())

	if p.currentLoop != nil {
		close(p.currentLoop)
	}
	p.currentLoop = stop

	go func() {
		for {
			select {
			case <-p.aliveArrived:
				continue

			case <-time.After(p.LeaderDeathTimer()):
				fmt.Printf("%s: leader(%s) of %d is dead\n", p.Self, p.currentTerm.Leader, p.currentTerm.ID)
				p.StartPoll()

			case <-stop:
				return
			}
		}
	}()
}

func (p *Polling) StopLoop() {
	if p.currentLoop != nil {
		close(p.currentLoop)
	}
	p.currentLoop = nil
}

type HTTPPolling struct {
	Polling *Polling

	Self  *Node
	Nodes []*Node
}

func NewHTTPPolling(self *Node, nodes []*Node) HTTPPolling {
	p := HTTPPolling{Self: self, Nodes: nodes}

	p.Polling = NewPolling(self, p.SendAlive, p.PollRequest)

	return p
}

func (p HTTPPolling) SendAlive(term Term) {
	wg := &sync.WaitGroup{}

	for _, n := range p.Nodes {
		if n.Equals(p.Self) {
			continue
		}

		wg.Add(1)
		go func(node *Node) {
			defer wg.Done()

			resp, err := node.Post("/alive", term)
			if err != nil {
				fmt.Printf("%s: failed to send alive to %s\n", p.Self, node)
			} else if resp.StatusCode != http.StatusNoContent {
				fmt.Printf("%s: denied alive by %s: %s\n", p.Self, node, resp.Status)
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
		x, err := json.Marshal(p.Polling.currentTerm)
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

func (p HTTPPolling) StartLoop() {
	p.Polling.StartFollowerLoop()
}

func main() {
	self := ForceNewNode(os.Args[1])

	nodes := []*Node{self}
	for _, u := range os.Args[2:] {
		nodes = append(nodes, ForceNewNode(u))
	}

	p := NewHTTPPolling(self, nodes)

	p.StartLoop()

	http.ListenAndServe(fmt.Sprintf(":%d", self.Port()), p)
}
