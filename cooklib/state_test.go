package cooklib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/vmihailenco/msgpack"
)

func Test_UUID(t *testing.T) {
	u := NewUUID([]byte("hello world"))
	excepted := "2a4a2ab2-f6b3-58b3-a885-704769b0a49c"
	if u.String() != excepted {
		t.Errorf("failed to calculate UUID: excepted %s but got %s", excepted, u)
	}

	bin := []byte{0x2a, 0x4a, 0x2a, 0xb2, 0xf6, 0xb3, 0x58, 0xb3, 0xa8, 0x85, 0x70, 0x47, 0x69, 0xb0, 0xa4, 0x9c}
	if bytes.Compare(u.Binary(), bin) != 0 {
		t.Errorf("failed to calculate UUID.Binary: excepted %x but got %x", bin, u.Binary())
	}

	j, err := json.Marshal(u)
	if err != nil {
		t.Errorf("failed to marshal to json: %s", err.Error())
	}

	if string(j) != fmt.Sprintf(`"%s"`, excepted) {
		t.Errorf(`failed to marshal to json: excepted "%s" but got %s`, excepted, string(j))
	}

	var u2 UUID
	err = json.Unmarshal(j, &u2)
	if err != nil {
		t.Errorf("failed to unmarshal from json: %s", err.Error())
	}

	if u2.String() != u.String() {
		t.Errorf("failed to unmarshal from json: excepted %s but got %s", u, u2)
	}
}

func Test_ChunkHolders(t *testing.T) {
	ch := ChunkHolders{
		NewChunkID([]byte("hello")): []*Node{MustParseNode("http://example.com")},
		NewChunkID([]byte("world")): []*Node{MustParseNode("http://foobar.com")},
	}

	b, err := msgpack.Marshal(ch)
	if err != nil {
		t.Errorf("failed to marshal to messagepack: %s", err.Error())
	}

	var ch2 ChunkHolders
	err = msgpack.Unmarshal(b, &ch2)
	if err != nil {
		t.Errorf("failed to unmarshal from messagepack: %s", err.Error())
	}

	if len(ch2) != len(ch) {
		t.Errorf("failed to unmarshal from messagepack: must have %d elements but got %d elements", len(ch), len(ch2))
	}

	for k, v := range ch2 {
		if fmt.Sprint(ch[k]) != fmt.Sprint(v) {
			t.Errorf("failed to unmarshal from messagepack: %s must have %s but got %s", k, ch[k], v)
		}
	}
}
