package main

import (
	"fmt"
	"strings"
	"sync"
)

var (
	ChunkNotFound  = fmt.Errorf("no such chunk")
	RecipeNotFound = fmt.Errorf("no such recipe")
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

type InMemoryRecipeStore struct {
	sync.Mutex

	data map[string]Recipe
}

func NewInMemoryRecipeStore() *InMemoryRecipeStore {
	return &InMemoryRecipeStore{data: make(map[string]Recipe)}
}

func (m *InMemoryRecipeStore) Bind(c *CookFS) {
}

func (m *InMemoryRecipeStore) Run(chan struct{}) error {
	return nil
}

func (m *InMemoryRecipeStore) Save(tag string, recipe Recipe) error {
	m.Lock()
	m.data[tag] = recipe
	m.Unlock()
	return nil
}

func (m *InMemoryRecipeStore) Load(tag string) (Recipe, error) {
	if data, ok := m.data[tag]; ok {
		return data, nil
	} else {
		return Recipe{}, RecipeNotFound
	}
}

func (m *InMemoryRecipeStore) Delete(tag string) error {
	if _, ok := m.data[tag]; !ok {
		return RecipeNotFound
	}

	m.Lock()
	delete(m.data, tag)
	m.Unlock()
	return nil
}

func (m *InMemoryRecipeStore) Find(prefix string) ([]string, error) {
	var result []string

	for tag := range m.data {
		if strings.HasPrefix(tag, prefix) {
			result = append(result, tag)
		}
	}

	return result, nil
}
