package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/vmihailenco/msgpack"
)

func Test_Node(t *testing.T) {
	raw := "http://localhost:5890/foobar"
	node, err := ParseNode(raw)
	if err != nil {
		t.Fatalf("failed to parse node: %s", err.Error())
	}

	if node.String() != raw {
		t.Errorf("failed to parse node: excepted %s but got %s", raw, node)
	}

	if node.Port() != "5890" {
		t.Errorf("excepted port is 5890 but got %s", node.Port())
	}

	j, err := json.Marshal(node)
	if err != nil {
		t.Errorf("failed to marshal to json: %s", err.Error())
	}
	if string(j) != fmt.Sprintf(`"%s"`, raw) {
		t.Errorf(`failed to marshal to json: excepted "%s" but got %s`, raw, string(j))
	}

	var node_j *Node
	err = json.Unmarshal(j, &node_j)
	if err != nil {
		t.Errorf("failed to unmarshal from json: %s", err.Error())
	}
	if node_j.String() != node.String() {
		t.Errorf(`failed to unmarshal from json: excepted %s but got %s`, node, node_j)
	}

	m, err := msgpack.Marshal(node)
	if err != nil {
		t.Errorf("failed to marshal to messagepack: %s", err.Error())
	}

	var node_m *Node
	err = msgpack.Unmarshal(m, &node_m)
	if err != nil {
		t.Errorf("failed to unmarshal from messagepack: %s", err.Error())
	}
	if node_m.String() != node.String() {
		t.Errorf(`failed to unmarshal from messagepack: excepted %s but got %s`, node, node_m)
	}
}
