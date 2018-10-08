package cooklib

import (
	"time"
)

type Config struct {
	AliveInterval    time.Duration
	AliveTimeout     time.Duration
	LeaderDeathTimer time.Duration
	CandidacyWaitMin time.Duration
	CandidacyWaitMax time.Duration
	CandidacyTimeout time.Duration
	PollingInterval  time.Duration

	SendWorkersNum int
}

var (
	DefaultConfig = Config{
		AliveInterval:    100 * time.Millisecond,
		AliveTimeout:     500 * time.Millisecond,
		LeaderDeathTimer: 500 * time.Millisecond,
		CandidacyWaitMin: 100 * time.Millisecond,
		CandidacyWaitMax: 1 * time.Second,
		CandidacyTimeout: 500 * time.Millisecond,
		PollingInterval:  1 * time.Second,

		SendWorkersNum: 10,
	}
)
