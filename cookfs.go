package main

type Runnable interface {
	Start() error
	Stop() error
}

type StorePlugin interface {
	Runnable

	Save(Hash, []byte) error
	Load(Hash) ([]byte, error)
	Delete(Hash) error
}

type EndpointPlugin interface {
	Runnable

	Self() *Node
	SendAlive(Term)
	PollRequest(Term) bool
}

type CookFS struct {
	Store    StorePlugin
	Endpoint EndpointPlugin

	Polling *Polling
}

func NewCookFS(store StorePlugin, endpoint EndpointPlugin) (*CookFS, error) {
	return &CookFS{store, endpoint, NewPolling(endpoint)}, nil
}

func (c CookFS) Start() error {
	for _, x := range []Runnable{c.Store, c.Endpoint, c.Polling} {
		if err := x.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (c CookFS) Stop() error {
	for _, x := range []Runnable{c.Store, c.Endpoint, c.Polling} {
		if err := x.Stop(); err != nil {
			return err
		}
	}
	return nil
}
