package main

import (
	"fmt"
	"strings"
	"sync"
)

var (
	ChunkNotFound   = fmt.Errorf("no such chunk")
	RecipieNotFound = fmt.Errorf("no such recipie")
)

type InMemoryChunkStore struct {
	sync.Mutex

	data map[Hash][CHUNK_SIZE]byte
}

func NewInMemoryChunkStore() *InMemoryChunkStore {
	return &InMemoryChunkStore{data: make(map[Hash][CHUNK_SIZE]byte)}
}

func (m *InMemoryChunkStore) Bind(c *CookFS) {
}

func (m *InMemoryChunkStore) Run(chan struct{}) error {
	return nil
}

func (m *InMemoryChunkStore) Save(chunk Chunk) error {
	m.Lock()
	m.data[chunk.Hash] = chunk.Data
	m.Unlock()
	return nil
}

func (m *InMemoryChunkStore) Load(h Hash) (Chunk, error) {
	if data, ok := m.data[h]; ok {
		return Chunk{h, data}, nil
	} else {
		return Chunk{}, ChunkNotFound
	}
}

func (m *InMemoryChunkStore) Delete(h Hash) error {
	if _, ok := m.data[h]; !ok {
		return ChunkNotFound
	}

	m.Lock()
	delete(m.data, h)
	m.Unlock()
	return nil
}

type InMemoryRecipieStore struct {
	sync.Mutex

	data map[string]Recipie
}

func NewInMemoryRecipieStore() *InMemoryRecipieStore {
	return &InMemoryRecipieStore{data: make(map[string]Recipie)}
}

func (m *InMemoryRecipieStore) Bind(c *CookFS) {
}

func (m *InMemoryRecipieStore) Run(chan struct{}) error {
	return nil
}

func (m *InMemoryRecipieStore) Save(tag string, recipie Recipie) error {
	m.Lock()
	m.data[tag] = recipie
	m.Unlock()
	return nil
}

func (m *InMemoryRecipieStore) Load(tag string) (Recipie, error) {
	if data, ok := m.data[tag]; ok {
		return data, nil
	} else {
		return Recipie{}, RecipieNotFound
	}
}

func (m *InMemoryRecipieStore) Delete(tag string) error {
	if _, ok := m.data[tag]; !ok {
		return RecipieNotFound
	}

	m.Lock()
	delete(m.data, tag)
	m.Unlock()
	return nil
}

func (m *InMemoryRecipieStore) Find(prefix string) ([]string, error) {
	var result []string

	for tag := range m.data {
		if strings.HasPrefix(tag, prefix) {
			result = append(result, tag)
		}
	}

	return result, nil
}
