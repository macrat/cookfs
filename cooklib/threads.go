package cooklib

import (
	"context"
	"strings"
	"time"
	"fmt"
)

type CookFS struct {
	leader *Node
	term   int64
	state  *State

	Nodes   func() []*Node
	Handler CommunicationHandler
	Config  Config

	alive   chan *Node
	polling chan PollingTask
}

func NewCookFS(handler CommunicationHandler, nodes func() []*Node, config Config) *CookFS {
	return &CookFS{
		state:   NewState(),
		Nodes:   nodes,
		Handler: handler,
		Config:  config,
		alive:   make(chan *Node),
		polling: make(chan PollingTask, len(nodes())*2),
	}
}

func (c *CookFS) AliveMessage(alive AliveMessage) Response {
	if (c.leader.String() == alive.Leader.String() && c.term == alive.Term) || c.term < alive.Term {
		c.alive <- alive.Leader
		c.leader = alive.Leader
		c.term = alive.Term

		return Response{StatusCode: 200}
	} else {
		return Response{StatusCode: 409}
	}
}

func (c *CookFS) PollRequest(request PollRequest) Response {
	if c.term <= request.Term && c.state.PatchID == request.PatchID {
		accept := make(chan bool)
		c.polling <- PollingTask{request, accept}

		select {
		case acc := <-accept:
			if acc {
				return Response{StatusCode: 204}
			} else {
				return Response{StatusCode: 409}
			}

		case <-time.After(c.Config.CandidacyTimeout):
			return Response{StatusCode: 409}
		}
	} else {
		return Response{StatusCode: 409}
	}
}

func (c *CookFS) HandleRequest(request Request) Response {
	if request.Data != nil {
		switch request.Path {
		case "/term":
			return c.AliveMessage(*request.Data.(*AliveMessage))

		case "/term/poll":
			return c.PollRequest(*request.Data.(*PollRequest))

		default:
			return Response{StatusCode: 404}
		}
	} else {
		switch request.Path {
		case "/term":
			return Response{200, AliveMessage{c.leader, c.term, c.state.PatchID}}

		default:
			return Response{StatusCode: 404}
		}
	}
}

func (c *CookFS) RunFollower(ctx context.Context) {
	var cancelCandidacy context.CancelFunc

	go PollingConsiliator(ctx, c.polling, c.Config.PollingWindow)

	for {
		select {
		case leader := <-c.alive:
			if leader.String() != c.Nodes()[0].String() && cancelCandidacy != nil {
				cancelCandidacy()
				cancelCandidacy = nil
			}

		case <-time.After(c.Config.LeaderDeathTimer):
			var ctx2 context.Context
			ctx2, cancelCandidacy = context.WithCancel(ctx)
			go c.RunCandidacy(ctx2)

		case <-ctx.Done():
			return
		}
	}
}

func (c *CookFS) RunCandidacy(ctx context.Context) {
	fmt.Println("been candidacy of term", c.term+1)

	withTimeout, _ := context.WithTimeout(ctx, c.Config.CandidacyTimeout)

	worker := NewWorkerPool(ctx, c.Handler, c.Config.SendWorkersNum)

	msg := PollRequest{c.Nodes()[0], c.term + 1, c.state.PatchID}

	if worker.OverHalf(withTimeout, c.Nodes(), "/term/poll", msg, c.Config.CandidacyTimeout) {
		c.term++
		c.leader = c.Nodes()[0]
		c.RunLeader(ctx, worker)
	}
}

func (c *CookFS) RunLeader(ctx context.Context, worker WorkerPool) {
	fmt.Println("been leader of term", c.term)

	sendAlive := func() {
		msg := AliveMessage{c.Nodes()[0], c.term, c.state.PatchID}
		worker.SendOnly(ctx, c.Nodes(), "/term", msg, c.Config.AliveTimeout)
	}

	go sendAlive()

	interval := time.Tick(c.Config.AliveInterval)

	for {
		select {
		case <-interval:
			go sendAlive()

		case <-ctx.Done():
			return
		}
	}
}

type WorkerTask struct {
	request  Request
	response chan Response
}

func CommunicateWorker(ctx context.Context, handler CommunicationHandler, task chan WorkerTask) {
	for {
		select {
		case t := <-task:
			c := ctx
			if t.request.Timeout != 0 {
				c, _ = context.WithTimeout(c, t.request.Timeout)
			}
			result := handler.Send(c, t.request)
			if t.response != nil {
				t.response <- result
			}

		case <-ctx.Done():
			return
		}
	}
}

type WorkerPool struct {
	task chan WorkerTask
	ctx  context.Context
}

func NewWorkerPool(ctx context.Context, handler CommunicationHandler, workersNum int) WorkerPool {
	w := WorkerPool{make(chan WorkerTask, workersNum*10), ctx}

	for i := 0; i < workersNum; i++ {
		go CommunicateWorker(ctx, handler, w.task)
	}

	return w
}

func (w WorkerPool) SendOnly(ctx context.Context, nodes []*Node, path string, data interface{}, timeout time.Duration) {
	for _, node := range nodes {
		select {
		case <-ctx.Done():
			return
		case <-w.ctx.Done():
			return

		default:
			w.task <- WorkerTask{Request{node, path, data, timeout}, nil}
		}
	}
}

func (w WorkerPool) OverHalf(ctx context.Context, nodes []*Node, path string, data interface{}, timeout time.Duration) bool {
	response := make(chan Response, len(nodes))

	for _, node := range nodes {
		w.task <- WorkerTask{Request{node, path, data, timeout}, response}
	}

	allow := 0
	deny := 0
	for range nodes {
		select {
		case resp := <-response:
			if resp.StatusCode == 200 || resp.StatusCode == 204 {
				allow++
			} else {
				deny++
			}

			if allow > len(nodes)/2 {
				return true
			} else if deny > len(nodes)/2 {
				return false
			}

		case <-ctx.Done():
			return false
		case <-w.ctx.Done():
			return false
		}
	}

	return false
}

type PollingTask struct {
	request PollRequest
	accept  chan bool
}

func PollingConsiliator(ctx context.Context, ch chan PollingTask, windowDuration time.Duration) {
	var window context.Context
	var candidate PollingTask
	var tasks []PollingTask

	waitWindow := func() <-chan struct{} {
		if window != nil {
			return window.Done()
		} else {
			return make(chan struct{})
		}
	}

	for {
		select {
		case task := <-ch:
			if window == nil {
				window, _ = context.WithTimeout(context.Background(), windowDuration)
				tasks = []PollingTask{task}
				candidate = task
			} else {
				tasks = append(tasks, task)
				if strings.Compare(task.request.Node.String(), candidate.request.Node.String()) < 0 {
					candidate = task
				}
			}

		case <-waitWindow():
			window = nil

			fmt.Println("poll to", candidate.request.Node.String(), "by", len(tasks), "nodes")

			candidate.accept <-true
			for _, t := range tasks {
				if t.request != candidate.request {
					t.accept <-false
				}
			}

		case <-ctx.Done():
			return
		}
	}
}
