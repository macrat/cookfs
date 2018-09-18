package main

import (
	"testing"
)

func Test_Hash(t *testing.T) {
	h := Hash{}
	except := ""
	if h.String() == except {
		t.Error("excepted: %s but got %s", except, h.String())
	}
}
