package main

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/vmihailenco/msgpack"
)

type Node url.URL

func ParseNode(raw string) (*Node, error) {
	u, err := url.ParseRequestURI(raw)
	return (*Node)(u), err
}

func MustParseNode(raw string) *Node {
	u, err := ParseNode(raw)
	if err != nil {
		panic(err.Error())
	}
	return u
}

func (n *Node) String() string {
	if n == nil {
		return ""
	}
	return (*url.URL)(n).String()
}

func (n *Node) Port() string {
	return (*url.URL)(n).Port()
}

func (n *Node) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(n.String())
}

func (n *Node) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("\"%s\"", n.String())
	return []byte(s), nil
}

func (n *Node) UnmarshalMsgpack(raw []byte) error {
	var s string

	if err := msgpack.Unmarshal(raw, &s); err != nil {
		return err
	}

	parsed, err := ParseNode(s)
	if err != nil {
		return err
	}

	*n = *parsed

	return nil
}

func (n *Node) UnmarshalJSON(raw []byte) error {
	if !json.Valid(raw) || raw[0] != '"' && raw[len(raw)-1] != '"' {
		return fmt.Errorf("invalid Node")
	}

	parsed, err := ParseNode(string(raw[1 : len(raw)-1]))
	if err != nil {
		return err
	}

	*n = *parsed

	return nil
}

type NodesFunc func() []*Node
