package cookfs

import (
	"encoding/json"
	"testing"

	"github.com/go-yaml/yaml"
)

var (
	node_a = ForceParseNode("http://example.com/path/to")
	node_b = ForceParseNode("http://localhost:8080/foo/bar")
)

func Test_Node_String(t *testing.T) {
	if node_a.String() != "http://example.com/path/to" {
		t.Error("failed convert to string")
	}
	if node_b.String() != "http://localhost:8080/foo/bar" {
		t.Error("failed convert to string")
	}
}

func Test_Node_Hostname(t *testing.T) {
	if node_a.Port() != 80 {
		t.Errorf("port must be 80 but got %d", node_a.Port())
	}
	if node_b.Port() != 8080 {
		t.Errorf("port must be 8080 but got %d", node_b.Port())
	}
}

func Test_Node_Port(t *testing.T) {
	if node_a.Hostname() != "example.com" {
		t.Errorf("port must be example.com but got %s", node_a.Hostname())
	}
	if node_b.Hostname() != "localhost" {
		t.Errorf("port must be localhost but got %s", node_b.Hostname())
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

	if !joined.Equals(ForceParseNode("http://example.com/path/to/foo/bar")) {
		t.Errorf("failed to join; got %s", joined)
	}
}

func Test_Node_JSON(t *testing.T) {
	j := []byte("\"http://localhost:8080/test\"")
	n := &Node{}

	if err := json.Unmarshal(j, n); err != nil {
		t.Errorf("failed to unmarshal json: %s", err.Error())
	}

	j2, err := json.Marshal(n)
	if err != nil {
		t.Errorf("failed to marshal json: %s", err.Error())
	}

	if string(j) != string(j2) {
		t.Errorf("failed to convert to/from json: excepted %s but got %s", j, j2)
	}
}

func Test_Node_YAML(t *testing.T) {
	y := []byte("http://localhost:8080/test\n")
	n := &Node{}

	if err := yaml.Unmarshal(y, n); err != nil {
		t.Errorf("failed to unmarshal yaml: %s", err.Error())
	}

	y2, err := yaml.Marshal(n)
	if err != nil {
		t.Errorf("failed to marshal yaml: %s", err.Error())
	}

	if string(y) != string(y2) {
		t.Errorf("failed to convert to/from yaml: excepted %s but got %s", y, y2)
	}
}
