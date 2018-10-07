package cooklib

import (
	"time"
)

type Config struct {
	AliveInterval    time.Duration
	LeaderDeathTimer time.Duration
	CandidacyWaitMin time.Duration
	CandidacyWaitMax time.Duration

	SendWorkersNum int
}

var (
	DefaultConfig = Config {
		AliveInterval: 100 * time.Millisecond,
		LeaderDeathTimer: 300 * time.Millisecond,
		CandidacyWaitMin: 100 * time.Millisecond,
		CandidacyWaitMax: 700 * time.Millisecond,

		SendWorkersNum: 10,
	}
)
