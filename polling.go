package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Polling struct {
	sync.Mutex

	Discover DiscoverPlugin
	Transmit TransmitPlugin

	PollingInterval   time.Duration
	SendAliveInterval time.Duration
	LeaderDeathTimer  time.Duration
	CandidacyTimerMin time.Duration
	CandidacyTimerMax time.Duration

	currentTerm  Term
	aliveArrived chan struct{}
	lastPoll     time.Time
	isLeader     bool
}

func NewPolling(discover DiscoverPlugin, transmit TransmitPlugin) *Polling {
	return &Polling{
		Discover:          discover,
		Transmit:          transmit,
		PollingInterval:   500 * time.Millisecond,
		SendAliveInterval: 100 * time.Millisecond,
		LeaderDeathTimer:  400 * time.Millisecond,
		CandidacyTimerMin: 100 * time.Millisecond,
		CandidacyTimerMax: 300 * time.Millisecond,
		aliveArrived:      make(chan struct{}),
	}
}

func (p *Polling) CurrentTerm() Term {
	return p.currentTerm
}

func (p *Polling) StartPoll() {
	newTerm := Term{
		ID:     p.currentTerm.ID + 1,
		Leader: p.Discover.Self(),
	}

	if p.Transmit.PollRequest(newTerm) {
		p.Lock()
		defer p.Unlock()

		fmt.Printf("%s: I'm leader of %d\n", p.Discover.Self(), newTerm.ID)

		p.currentTerm = newTerm
		p.isLeader = true
	} else {
		fmt.Printf("%s: failed to promotion to leader of %d\n", p.Discover.Self(), newTerm.ID)
	}
}

func (p *Polling) AliveArrived(term Term) (accepted bool) {
	if !term.NewerThan(p.currentTerm) && !term.Equals(p.currentTerm) {
		fmt.Printf("denied: %s -> %s\n", p.currentTerm, term)
		return false
	}

	if !term.Equals(p.currentTerm) {
		fmt.Printf("%s: term changed to %s\n", p.Discover.Self(), term)
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

		fmt.Printf("%s: vote to %s\n", p.Discover.Self(), term)

		return true
	} else {
		return false
	}
}

func (p *Polling) Loop(stop chan struct{}) {
	rand.Seed(time.Now().UnixNano())

	candidacyCount := 1

	for {
		if p.isLeader {
			select {
			case <-p.aliveArrived:
				fmt.Printf("%s: relegation to follower because arrived alive of %s\n", p.Discover.Self(), p.currentTerm)
				p.isLeader = false
				return

			case <-time.After(p.SendAliveInterval):
				p.Transmit.SendAlive(p.currentTerm)

			case <-stop:
				return
			}
		} else {
			deathTimer := time.Duration(rand.Int63n(int64(p.CandidacyTimerMax-p.CandidacyTimerMin))) + p.CandidacyTimerMin
			if candidacyCount == 0 {
				deathTimer += p.LeaderDeathTimer
			} else {
				deathTimer *= time.Duration(candidacyCount)
			}

			select {
			case <-p.aliveArrived:
				candidacyCount = 0
				continue

			case <-time.After(deathTimer):
				fmt.Printf("%s: %s is dead\n", p.Discover.Self(), p.currentTerm)
				p.StartPoll()
				if p.isLeader {
					candidacyCount = 0
				} else {
					candidacyCount++
				}

			case <-stop:
				return
			}
		}
	}
}

func (p *Polling) IsLeader() bool {
	return p.isLeader
}

func (p *Polling) Bind(c *CookFS) {
}

func (p *Polling) Run(stop chan struct{}) error {
	go p.Loop(stop)

	return nil
}
