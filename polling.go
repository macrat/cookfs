package main

import (
	"sync"
	"time"
	"math/rand"
	"fmt"
)

type Polling struct {
	sync.Mutex

	Endpoint EndpointPlugin

	PollingInterval     time.Duration
	SendAliveInterval   time.Duration
	LeaderDeathTimerMin time.Duration
	LeaderDeathTimerMax time.Duration

	currentTerm  Term
	aliveArrived chan struct{}
	lastPoll     time.Time

	currentLoop chan struct{}
}

func NewPolling(endpoint EndpointPlugin) *Polling {
	return &Polling{
		Endpoint:            endpoint,
		PollingInterval:     1 * time.Second,
		SendAliveInterval:   100 * time.Millisecond,
		LeaderDeathTimerMin: 500 * time.Millisecond,
		LeaderDeathTimerMax: 600 * time.Millisecond,
		aliveArrived:        make(chan struct{}),
	}
}

func (p *Polling) CurrentTerm() Term {
	return p.currentTerm
}

func (p *Polling) StartPoll() {
	newTerm := Term{
		ID:     p.currentTerm.ID + 1,
		Leader: p.Endpoint.Self(),
	}

	if p.Endpoint.PollRequest(newTerm) {
		p.Lock()
		defer p.Unlock()

		fmt.Printf("%s: I'm leader of %d\n", p.Endpoint.Self(), newTerm.ID)

		p.currentTerm = newTerm
		p.startLeaderLoop()
	} else {
		fmt.Printf("%s: failed to promotion to leader of %d\n", p.Endpoint.Self(), newTerm.ID)
	}
}

func (p *Polling) AliveArrived(term Term) (accepted bool) {
	if !term.NewerThan(p.currentTerm) && !term.Equals(p.currentTerm) {
		fmt.Printf("denied: %s -> %s\n", p.currentTerm, term)
		return false
	}

	if !term.Equals(p.currentTerm) {
		fmt.Printf("%s: term changed to %s\n", p.Endpoint.Self(), term)
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

		fmt.Printf("%s: vote to %s\n", p.Endpoint.Self(), term)

		return true
	} else {
		return false
	}
}

func (p *Polling) startLeaderLoop() {
	stop := make(chan struct{})

	if p.currentLoop != nil {
		close(p.currentLoop)
	}
	p.currentLoop = stop

	go func() {
		for {
			select {
			case <-p.aliveArrived:
				fmt.Printf("%s: relegation to follower because arrived alive of %s\n", p.Endpoint.Self(), p.currentTerm)
				p.startFollowerLoop()
				return

			case <-time.After(p.SendAliveInterval):
				p.Endpoint.SendAlive(p.currentTerm)

			case <-stop:
				return
			}
		}
	}()
}

func (p *Polling) leaderDeathTimer() time.Duration {
	return time.Duration(rand.Int63n(int64(p.LeaderDeathTimerMax-p.LeaderDeathTimerMin))) + p.LeaderDeathTimerMin
}

func (p *Polling) startFollowerLoop() {
	stop := make(chan struct{})

	rand.Seed(time.Now().UnixNano())

	if p.currentLoop != nil {
		close(p.currentLoop)
	}
	p.currentLoop = stop

	go func() {
		candidate_count := 1

		for {
			select {
			case <-p.aliveArrived:
				candidate_count = 1
				continue

			case <-time.After(p.leaderDeathTimer() * time.Duration(candidate_count)):
				fmt.Printf("%s: %s is dead\n", p.Endpoint.Self(), p.currentTerm)
				p.StartPoll()
				candidate_count++

			case <-stop:
				return
			}
		}
	}()
}

func (p *Polling) Start() error {
	p.Stop()
	p.startFollowerLoop()
	return nil
}

func (p *Polling) Stop() error {
	if p.currentLoop != nil {
		close(p.currentLoop)
	}
	p.currentLoop = nil
	return nil
}
