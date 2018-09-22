package main

import (
	"encoding/json"
	"testing"
)

func Test_Recipie(t *testing.T) {
	r := Recipie{CalcHash([]byte("hello")), CalcHash([]byte("world"))}
	except := "[\"9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043\",\"11853df40f4b2b919d3815f64792e58d08663767a494bcbb38c0b2389d9140bbb170281b4a847be7757bde12c9cd0054ce3652d0ad3a1a0c92babb69798246ee\"]"

	if j, err := json.Marshal(r); err != nil {
		t.Errorf("failed to marshal recipie; %s", err.Error())
	} else if string(j) != except {
		t.Errorf("failed to marshal recipie; except %v but got %v", except, string(j))
	}

	var r2 Recipie
	if err := json.Unmarshal([]byte(except), &r2); err != nil {
		t.Errorf("failed to unmarshal recipie; %s", err.Error())
	} else if len(r2) != len(r) {
		t.Errorf("unmarshaled recipie has unexcepted data; excepted %v but got %v", r, r2)
	} else {
		for i, x := range r2 {
			if r[i] != x {
				t.Errorf("unmarshaled recipie has unexcepted data; excepted %v but got %v", r, r2)
			}
		}
	}
}
