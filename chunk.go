package main

import (
	"crypto/sha512"
	"encoding/hex"
)

type Hash [sha512.Size]byte

func CalcHash(data []byte) Hash {
	return Hash(sha512.Sum512(data))
}

func ParseHash(raw string) (Hash, error) {
	x, err := hex.DecodeString(raw)
	if err != nil {
		return Hash{}, err
	}

	var h Hash
	copy(h[:], x)

	return h, nil
}

func (h Hash) String() string {
	return hex.EncodeToString([]byte(h[:]))
}

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
