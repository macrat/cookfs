package main

const (
	CHUNK_SIZE = 64
)

type Chunk struct {
	Hash Hash
	Data [CHUNK_SIZE]byte
}

func NewChunk(hash Hash, data []byte) Chunk {
	chunk := Chunk{Hash: hash}
	copy(chunk.Data[:], data)
	return chunk
}
