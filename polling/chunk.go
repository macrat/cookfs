package main

import (
	"crypto/sha512"
	"encoding/hex"
)

type Hash [sha512.Size]byte

func ParseHash(raw string) (Hash, error) {
	h, err := hex.DecodeString(raw)
	if err != nil {
		return Hash{}, err
	}
	return Hash{}, err // TODO
}

func (h Hash) String() string {
	return hex.EncodeToString([]byte(h[:]))
}

const (
	CHUNK_SIZE = 64
)

type Chunk struct {
	Hash string
	Data [CHUNK_SIZE]byte
}
