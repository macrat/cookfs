package main

import (
	"os"
)

type DummyStore struct {}

func (ds DummyStore) Bind(c *CookFS) {
}

func (ds DummyStore) Run(chan struct{}) error {
	return nil
}

func (ds DummyStore) Save(h Hash, b []byte) error {
	return nil
}

func (ds DummyStore) Load(h Hash) ([]byte, error) {
	return nil, nil
}

func (ds DummyStore) Delete(h Hash) error {
	return nil
}

func main() {
	self := ForceParseNode(os.Args[1])

	nodes := []*Node{self}
	for _, u := range os.Args[2:] {
		nodes = append(nodes, ForceParseNode(u))
	}

	store := DummyStore{}
	discover := SimpleDiscoverPlugin{self, nodes}
	transmit := &HTTPTransmitPlugin{}
	receive := &HTTPReceivePlugin{}

	cookfs := NewCookFS(store, discover, transmit, receive)

	stop := make(chan struct{})
	cookfs.Run(stop)
	for {
		select {
		case <-stop:
			panic("stopped")
			return
		}
	}
}
