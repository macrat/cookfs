package main

import (
	"testing"
)

func Test_Hash(t *testing.T) {
	data := "0102030405060708090a0b0c0d0e0f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	h, err := ParseHash(data)
	if err != nil {
		t.Errorf("failed to parse: %s", err.Error())
	}
	if h.String() != data {
		t.Errorf("excepted: %s but got %s", data, h.String())
	}
}
