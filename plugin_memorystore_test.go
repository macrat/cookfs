package main_test

import (
	"testing"

	cookfs "."
)

func Test_InMemoryChunkStore(t *testing.T) {
	m := cookfs.NewInMemoryChunkStore()

	chunk := cookfs.NewChunk(cookfs.Hash{}, []byte("hello"))

	if err := m.Save(chunk); err != nil {
		t.Errorf("failed to save chunk because; %s", err.Error())
	}

	if result, err := m.Load(chunk.Hash); err != nil {
		t.Errorf("failed to load chunk because; %s", err.Error())
	} else if result.Data != chunk.Data {
		t.Errorf("mismatched saved data and loaded data; except %s but got %v", "hello", result.Data)
	}

	if err := m.Delete(chunk.Hash); err != nil {
		t.Errorf("failed to delete chunk because; %s", err.Error())
	} else if _, err := m.Load(chunk.Hash); err != cookfs.ChunkNotFound {
		t.Errorf("InMemoryChunkStore.Delete was succeed but data is not deleted")
	}
}
