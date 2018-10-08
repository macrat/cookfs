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
	PollingWindow    time.Duration

	SendWorkersNum int
}

var (
	DefaultConfig = Config{
		AliveInterval:    100 * time.Millisecond,
		AliveTimeout:     500 * time.Millisecond,
		LeaderDeathTimer: 1500 * time.Millisecond,
		CandidacyTimeout: 1000 * time.Millisecond,
		PollingWindow:    500 * time.Millisecond,

		SendWorkersNum: 10,
	}
)
