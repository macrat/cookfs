package main

type Runnable interface {
	Run(chan struct{}) error
}

type Plugin interface {
	Runnable

	Bind(*CookFS)
}

type RecipiePlugin interface {
	Plugin

	Save(tag string, recipie Recipie) error
	Load(tag string) (Recipie, error)
	Delete(tag string) error
	Find(prefix string) ([]string, error)
}

type ChunkPlugin interface {
	Plugin

	Save(Chunk) error
	Load(Hash) (Chunk, error)
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
	Recipie  RecipiePlugin
	Chunk    ChunkPlugin
	Discover DiscoverPlugin
	Transmit TransmitPlugin
	Receive  ReceivePlugin

	Polling *Polling
}

func NewCookFS(recepie RecipiePlugin, chunk ChunkPlugin, discover DiscoverPlugin, transmit TransmitPlugin, receive ReceivePlugin) *CookFS {
	c := &CookFS{
		Recipie:  recepie,
		Chunk:    chunk,
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
	return []Plugin{c.Recipie, c.Chunk, c.Discover, c.Transmit, c.Receive}
}

func (c CookFS) runnables() []Runnable {
	return []Runnable{c.Recipie, c.Chunk, c.Discover, c.Transmit, c.Receive, c.Polling}
}

func (c CookFS) Run(stop chan struct{}) error {
	for _, x := range c.runnables() {
		if err := x.Run(stop); err != nil {
			return err
		}
	}
	return nil
}
