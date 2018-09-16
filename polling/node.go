package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

type Node url.URL

func ParseNode(raw_url string) (*Node, error) {
	u, err := url.Parse(raw_url)
	if err != nil {
		return nil, err
	}
	return (*Node)(u), nil
}

func ForceParseNode(raw_url string) *Node {
	u, err := url.Parse(raw_url)
	if err != nil {
		panic(err.Error())
	}
	return (*Node)(u)
}

func (n *Node) String() string {
	return (*url.URL)(n).String()
}

func (n *Node) Port() int {
	port, err := strconv.Atoi((*url.URL)(n).Port())
	if err != nil {
		return 80
	}
	return port
}

func (n *Node) Equals(another *Node) bool {
	return *n == *another
}

func (n *Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

func (n *Node) UnmarshalJSON(raw []byte) error {
	var x string
	if err := json.Unmarshal(raw, &x); err != nil {
		return err
	}

	u, err := url.Parse(x)
	if err != nil {
		return err
	}

	*n = (Node)(*u)
	return nil
}

func (n *Node) Join(subpath string) *Node {
	u := *n
	u.Path = path.Join(u.Path, subpath)
	return &u
}

func (n *Node) Get(endpoint string) (*http.Response, error) {
	return (&http.Client{Timeout: 200 * time.Millisecond}).Get(n.Join(endpoint).String())
}

func (n *Node) Post(endpoint string, data interface{}) (*http.Response, error) {
	x, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return (&http.Client{Timeout: 200 * time.Millisecond}).Post(n.Join(endpoint).String(), "application/json", bytes.NewReader(x))
}
