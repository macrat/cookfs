package cooklib

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type CookFS struct {
	sync.RWMutex

	leader *Node
	term   int64
	state  *State

	Nodes   func() []*Node
	Handler CommunicationHandler
	Config  Config

	alive chan *Node
}

func NewCookFS(handler CommunicationHandler, nodes func() []*Node, config Config) *CookFS {
	return &CookFS{
		state:   NewState(),
		Nodes:   nodes,
		Handler: handler,
		Config:  config,
		alive:   make(chan *Node),
	}
}

func (c *CookFS) AliveMessage(alive AliveMessage) Response {
	c.RLock()
	if (c.leader.String() == alive.Leader.String() && c.term == alive.Term) || c.term < alive.Term {
		c.RUnlock()
		c.Lock()
		c.alive <- alive.Leader
		c.leader = alive.Leader
		c.term = alive.Term
		c.Unlock()

		return Response{StatusCode: 200}
	} else {
		c.RUnlock()
		return Response{StatusCode: 409}
	}
}

func (c *CookFS) PollRequest(request PollRequest) Response {
	return Response{StatusCode: 200}
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
			c.RLock()
			msg := AliveMessage{c.leader, c.term, c.state.PatchID}
			c.RUnlock()
			return Response{200, msg}

		default:
			return Response{StatusCode: 404}
		}
	}
}

func (c *CookFS) RunFollower(ctx context.Context) {
	var cancelCandidacy context.CancelFunc
	candidacyCount := 0

	for {
		deathTimer := time.Duration(rand.Int63n(int64(c.Config.CandidacyWaitMax-c.Config.CandidacyWaitMin))) + c.Config.CandidacyWaitMin
		if candidacyCount == 0 {
			deathTimer += c.Config.LeaderDeathTimer
		} else {
			deathTimer *= time.Duration(candidacyCount + 1)
		}

		select {
		case leader := <-c.alive:
			candidacyCount = 0
			if leader.String() != c.Nodes()[0].String() && cancelCandidacy != nil {
				cancelCandidacy()
				cancelCandidacy = nil
			}
			continue

		case <-time.After(deathTimer):
			candidacyCount++
			var ctx2 context.Context
			ctx2, cancelCandidacy = context.WithCancel(ctx)
			go c.RunCandidacy(ctx2)

		case <-ctx.Done():
			return
		}
	}
}

func (c *CookFS) RunCandidacy(ctx context.Context) {
	println("been candidacy")

	withTimeout, _ := context.WithTimeout(ctx, 1*time.Second)

	worker := NewWorkerPool(ctx, c.Handler, c.Config.SendWorkersNum)

	c.RLock()
	msg := PollRequest{c.Nodes()[0], c.term + 1, c.state.PatchID}
	c.RUnlock()

	if worker.OverHalf(withTimeout, c.Nodes(), "/term/poll", msg) {
		c.Lock()
		c.term++
		c.leader = c.Nodes()[0]
		c.Unlock()
		c.RunLeader(ctx, worker)
	}
}

func (c *CookFS) RunLeader(ctx context.Context, worker WorkerPool) {
	println("been leader")

	interval := time.Tick(c.Config.AliveInterval)

	for {
		select {
		case <-interval:
			c.RLock()
			msg := AliveMessage{c.Nodes()[0], c.term, c.state.PatchID}
			c.RUnlock()
			worker.SendOnly(ctx, c.Nodes(), "/term", msg)

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
			result := handler.Send(ctx, t.request)
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
	w := WorkerPool{make(chan WorkerTask), ctx}

	for i := 0; i < workersNum; i++ {
		go CommunicateWorker(ctx, handler, w.task)
	}

	return w
}

func (w WorkerPool) SendOnly(ctx context.Context, nodes []*Node, path string, data interface{}) {
	for _, node := range nodes {
		w.task <- WorkerTask{Request{node, path, data}, nil}

		select {
		case <-ctx.Done():
			return
		case <-w.ctx.Done():
			return

		default:
		}
	}
}

func (w WorkerPool) OverHalf(ctx context.Context, nodes []*Node, path string, data interface{}) bool {
	response := make(chan Response)

	for _, node := range nodes {
		w.task <- WorkerTask{Request{node, path, data}, response}
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
