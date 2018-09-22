package main

type Runnable interface {
	Run(chan struct{}) error
}

type Plugin interface {
	Runnable

	Bind(*CookFS)
}

type RecipePlugin interface {
	Plugin

	Save(tag string, recipe Recipe) error
	Load(tag string) (Recipe, error)
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
	Recipe   RecipePlugin
	Chunk    ChunkPlugin
	Discover DiscoverPlugin
	Transmit TransmitPlugin
	Receive  ReceivePlugin

	Polling *Polling
	Journal *JournalManager
}

func NewCookFS(recepie RecipePlugin, chunk ChunkPlugin, discover DiscoverPlugin, transmit TransmitPlugin, receive ReceivePlugin) *CookFS {
	c := &CookFS{
		Recipe:   recepie,
		Chunk:    chunk,
		Discover: discover,
		Transmit: transmit,
		Receive:  receive,
		Polling:  NewPolling(discover, transmit),
		Journal:  &JournalManager{},
	}

	for _, p := range c.plugins() {
		p.Bind(c)
	}

	return c
}

func (c CookFS) plugins() []Plugin {
	return []Plugin{c.Recipe, c.Chunk, c.Discover, c.Transmit, c.Receive}
}

func (c CookFS) runnables() []Runnable {
	return []Runnable{c.Recipe, c.Chunk, c.Discover, c.Transmit, c.Receive, c.Polling}
}

func (c CookFS) Run(stop chan struct{}) error {
	for _, x := range c.runnables() {
		if err := x.Run(stop); err != nil {
			return err
		}
	}
	return nil
}
