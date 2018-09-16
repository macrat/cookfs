package main_test

import (
	"testing"

	cookfs "."
)

var (
	node_a = cookfs.ForceParseNode("http://example.com/path/to")
	node_b = cookfs.ForceParseNode("http://localhost:8080/foo/bar")
)

func Test_Node_String(t *testing.T) {
	if node_a.String() != "http://example.com/path/to" {
		t.Error("failed convert to string")
	}
	if node_b.String() != "http://localhost:8080/foo/bar" {
		t.Error("failed convert to string")
	}
}

func Test_Node_Port(t *testing.T) {
	if node_a.Port() != 80 {
		t.Errorf("port must be 80 but got %d", node_a.Port())
	}
	if node_b.Port() != 8080 {
		t.Errorf("port must be 8080 but got %d", node_b.Port())
	}
}

func Test_Node_Equals(t *testing.T) {
	if node_a.Equals(node_b) {
		t.Errorf("saied %s and %s is equals", node_a, node_b)
	}
	if !node_a.Equals(node_a) {
		t.Errorf("saied %s and %s is not equals", node_a, node_a)
	}
	if !node_b.Equals(node_b) {
		t.Errorf("saied %s and %s is not equals", node_b, node_b)
	}
}

func Test_Node_Join(t *testing.T) {
	joined := node_a.Join("/foo/bar")

	if !joined.Equals(cookfs.ForceParseNode("http://example.com/path/to/foo/bar")) {
		t.Errorf("failed to join; got %s", joined)
	}
}
