package cookfs

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	becameCandidacyTotal prometheus.Counter
	becameLeaderTotal    prometheus.Counter
	doPollingTotal       prometheus.Counter
	denyPollingTotal     prometheus.Counter
	currentTermID        prometheus.Counter
	journalAddTotal      prometheus.Counter
	journalCommitTotal   prometheus.Counter
}

func NewMetrics() Metrics {
	var m Metrics

	m.becameCandidacyTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "cookfs",
		Name:      "became_candidacy_total",
		Help:      "Count of became candidacy",
	})
	m.becameLeaderTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "cookfs",
		Name:      "became_leader_total",
		Help:      "Count of became leader",
	})
	m.doPollingTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "cookfs",
		Name:      "do_polling_total",
		Help:      "Count of did polling",
	})
	m.denyPollingTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "cookfs",
		Name:      "deny_polling_total",
		Help:      "Count of denied polling",
	})
	m.currentTermID = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "cookfs",
		Name:      "current_term_id",
		Help:      "Current term ID (that is total count of term changed in the cluster)",
	})
	m.journalAddTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "cookfs",
		Name:      "journal_add_total",
		Help:      "Count of journal entry added",
	})
	m.journalCommitTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "cookfs",
		Name:      "journal_commit_total",
		Help:      "Count of journal entry committed",
	})

	prometheus.MustRegister(m.becameCandidacyTotal)
	prometheus.MustRegister(m.becameLeaderTotal)
	prometheus.MustRegister(m.doPollingTotal)
	prometheus.MustRegister(m.denyPollingTotal)
	prometheus.MustRegister(m.currentTermID)
	prometheus.MustRegister(m.journalAddTotal)
	prometheus.MustRegister(m.journalCommitTotal)

	return m
}

func (m Metrics) Run(chan struct{}) error {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":3000", nil)
	}()
	return nil
}

func (m Metrics) Bind(c *CookFS) {
}

func (m Metrics) BecameCandidacy() {
	m.becameCandidacyTotal.Inc()
}

func (m Metrics) BecameLeader() {
	m.becameLeaderTotal.Inc()
}

func (m Metrics) DoPolling() {
	m.doPollingTotal.Inc()
}

func (m Metrics) DenyPolling() {
	m.denyPollingTotal.Inc()
}

func (m Metrics) TermChanged(oldTerm, newTerm Term) {
	m.currentTermID.Add(float64(newTerm.ID - oldTerm.ID))
}

func (m Metrics) JournalAdded() {
	m.journalAddTotal.Inc()
}

func (m Metrics) JournalCommitted() {
	m.journalCommitTotal.Inc()
}
