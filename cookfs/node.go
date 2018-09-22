package cookfs

import (
	"encoding/json"
	"net"
	"net/url"
	"path"
	"strconv"
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

func (n *Node) Hostname() string {
	h, _, err := net.SplitHostPort(n.Host)
	if err != nil {
		return n.Host
	}
	return h
}

func (n *Node) Port() int {
	_, p, err := net.SplitHostPort(n.Host)
	if err != nil {
		return 80
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return 80
	}
	return port
}

func (n *Node) Equals(another *Node) bool {
	return *n == *another
}

func (n *Node) Join(subpath string) *Node {
	u := *n
	u.Path = path.Join(u.Path, subpath)
	return &u
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

func (n *Node) MarshalYAML() (interface{}, error) {
	return n.String(), nil
}

func (n *Node) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	u, err := url.Parse(s)
	if err != nil {
		return err
	}

	*n = (Node)(*u)

	return nil
}
