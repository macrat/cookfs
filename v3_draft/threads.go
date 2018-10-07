package main

import (
	"context"
	"net/url"
	"time"
)

type Follower struct {
	Leader *url.URL
	Term   int64
	State  *State

	alive chan struct{}
}

func NewFollower() Follower {
	return Follower{alive: make(chan struct{})}
}

func (f Follower) AliveMessage(alive AliveMessage) Response {
	if f.Leader == alive.Leader && f.Term == alive.Term || f.Term < alive.Term {
		f.Leader = alive.Leader
		f.Term = alive.Term
		f.alive <- struct{}{}

		return Response{StatusCode: 200}
	} else {
		return Response{StatusCode: 409}
	}
}

func (f Follower) Run(ctx context.Context, nodes func() []*url.URL) {
	var cancelCandidacy context.CancelFunc

	for {
		select {
		case <-f.alive:
			if cancelCandidacy != nil {
				cancelCandidacy()
				cancelCandidacy = nil
			}
			continue

		case <-time.After(500 * time.Millisecond):
			var c context.Context
			c, cancelCandidacy = context.WithTimeout(ctx, 1*time.Second)
			go Candidacy(c, f.Term+1, nodes)

		case <-ctx.Done():
			return
		}
	}
}

func Candidacy(ctx context.Context, term int64, nodes func() []*url.URL) {
	println("been candidacy")

	worker := NewWorkerPool(ctx, 10)

	if worker.OverHalf(nodes(), "/term/poll", term) {
		Leader(ctx, term, nodes, worker)
	}
}

func Leader(ctx context.Context, term int64, nodes func() []*url.URL, worker WorkerPool) {
	// TODO: do something
	interval := timer.Interval(100*time.Millisecond)

	for {
		select {
			case <-interval:
				worker.SendOnly(nodes(), "/term", AliveMessage{nodes()[0], term, /*TODO patch id*/})

			case <-ctx.Done():
				return
	}
}

type WorkerTask struct {
	request  Request
	response chan Response
}

func CommunicateWorker(ctx context.Context, task chan WorkerTask) {
	for {
		select {
		case t := <-task:
			result := handler.Send(ctx, task.request)
			if task.response != nil {
				task.response <- result
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

func NewWorkerPool(ctx context.Context, workersNum int) WorkerPool {
	w := WorkerPool{make(chan WorkerTask), ctx}

	for i := 0; i < workersNum; i++ {
		go CommunicateWorker(ctx, w.task)
	}

	return w
}

func (w WorkerPool) SendOnly(nodes []*url.URL, path string, data interface{}) {
	for _, node := range nodes {
		w.task <- WorkerTask{{node, path, data}, nil}
	}
}

func (w WorkerPool) OverHalf(nodes []*url.URL, path string, data interface{}) bool {
	response := chan Response

	for _, node ;= range nodes {
		w.task <- WorkerTask{{node, path, data}, response}
	}

	allow := 0
	deny := 0
	for range nodes {
		select {
		case resp := <-response:
			if resp.StautsCode == 200 || resp.StatusCode == 204 {
				allow++
			} else {
				deny++
			}

			if allow > len(nodes)/2 {
				return true
			} else if deny > len(nodes)/2 {
				return false
			}

		case <-w.ctx.Done():
			return false
		}
	}

	return false
}
