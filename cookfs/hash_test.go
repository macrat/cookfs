package cookfs

import (
	"encoding/json"
	"testing"

	"github.com/go-yaml/yaml"
)

func Test_Hash(t *testing.T) {
	data := "0102030405060708090a0b0c0d0e0f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	h, err := ParseHash(data)
	if err != nil {
		t.Errorf("failed to parse: %s", err.Error())
	}
	if h.String() != data {
		t.Errorf("excepted %s but got %s", data, h.String())
	}
	if h.ShortHash() != data[:8] {
		t.Errorf("excepted %s but got %s", data, h.String())
	}
}

func Test_CalcHash(t *testing.T) {
	except, err := ParseHash("2f3831bccc94cf061bcfa5f8c23c1429d26e3bc6b76edad93d9025cb91c903af6cf9c935dc37193c04c2c66e7d9de17c358284418218afea2160147aaa912f4c")
	if err != nil {
		t.Errorf("failed to parse hash; %s", err.Error())
	}

	if h := CalcHash([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}); h != except {
		t.Errorf("failed to calcurate hash: got %x", h)
	}
}

func Test_CalcHash_MultiData(t *testing.T) {
	hello := CalcHash([]byte("hello"))
	world := CalcHash([]byte("world"))

	except := CalcHash(append(hello[:], world[:]...))
	got := CalcHash(hello[:], world[:])

	if got != except {
		t.Errorf("excepted %s but got %s", except, got)
	}
}

func Test_Hash_JSON(t *testing.T) {
	j := []byte("\"0102030405060708090a0b0c0d0e0f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\"")
	var h Hash

	if err := json.Unmarshal(j, &h); err != nil {
		t.Errorf("failed to unmarshal json: %s", err.Error())
	}

	j2, err := json.Marshal(h)
	if err != nil {
		t.Errorf("failed to marshal json: %s", err.Error())
	}

	if string(j) != string(j2) {
		t.Errorf("failed to convert to/from json: excepted %s but got %s", j, j2)
	}
}

func Test_Hash_YAML(t *testing.T) {
	y := []byte("0102030405060708090a0b0c0d0e0f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\n")
	var h Hash

	if err := yaml.Unmarshal(y, &h); err != nil {
		t.Errorf("failed to unmarshal yaml: %s", err.Error())
	}

	y2, err := yaml.Marshal(h)
	if err != nil {
		t.Errorf("failed to marshal yaml: %s", err.Error())
	}

	if string(y) != string(y2) {
		t.Errorf("failed to convert to/from yaml: excepted %s but got %s", y, y2)
	}
}
