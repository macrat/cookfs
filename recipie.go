package main

import (
	"encoding/json"
)

type Recipie struct {
	Tag  string
	Data []Hash
}

func (r Recipie) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Data)
}

func (r *Recipie) UnmarshalJSON(raw []byte) error {
	r.Tag = ""
	return json.Unmarshal(raw, &r.Data)
}
