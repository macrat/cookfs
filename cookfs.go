package main

type Runnable interface {
	Run(chan struct{}) error
}

type Plugin interface {
	Runnable

	Bind(*CookFS)
}

type StorePlugin interface {
	Plugin

	Save(Hash, []byte) error
	Load(Hash) ([]byte, error)
	Delete(Hash) error
}

type DiscoverPlugin interface {
	Plugin

	Self() *Node
	Nodes() []*Node
}

type TransmitPlugin interface {
	Plugin

	SendAlive(Term)
	PollRequest(Term) bool
}

type ReceivePlugin interface {
	Plugin
}

type CookFS struct {
	Store    StorePlugin
	Discover DiscoverPlugin
	Transmit TransmitPlugin
	Receive  ReceivePlugin

	Polling *Polling
}

func NewCookFS(store StorePlugin, discover DiscoverPlugin, transmit TransmitPlugin, receive ReceivePlugin) *CookFS {
	c := &CookFS{
		Store:    store,
		Discover: discover,
		Transmit: transmit,
		Receive:  receive,
		Polling:  NewPolling(discover, transmit),
	}

	for _, p := range c.plugins() {
		p.Bind(c)
	}

	return c
}

func (c CookFS) plugins() []Plugin {
	return []Plugin{c.Store, c.Discover, c.Transmit, c.Receive}
}

func (c CookFS) runnables() []Runnable {
	return []Runnable{c.Store, c.Discover, c.Transmit, c.Receive, c.Polling}
}

func (c CookFS) Run(stop chan struct{}) error {
	for _, x := range c.runnables() {
		if err := x.Run(stop); err != nil {
			return err
		}
	}
	return nil
}
