package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack"
)

var (
	Namespace = uuid.NewSHA1(uuid.UUID{}, []byte("cookfs"))
)

type UUID uuid.UUID

func NewUUID(data []byte) UUID {
	return UUID(uuid.NewSHA1(Namespace, data))
}

func (u UUID) String() string {
	return uuid.UUID(u).String()
}

func (u UUID) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("\"%s\"", u.String())
	return []byte(s), nil
}

func (u *UUID) UnmarshalJSON(raw []byte) error {
	if raw[0] != '"' && raw[len(raw)-1] != '"' {
		return fmt.Errorf("invalid UUID")
	}

	parsed, err := uuid.ParseBytes(raw[1 : len(raw)-1])
	if err != nil {
		return err
	}

	*u = UUID(parsed)

	return nil
}

type ChunkID struct {
	UUID
}

func NewChunkID(data []byte) ChunkID {
	return ChunkID{NewUUID(data)}
}

type Recipe []ChunkID

type RecipeList map[string]Recipe

func (r RecipeList) MarshalMsgpack() ([]byte, error) {
	data := make(map[string]interface{})
	for k, v := range r {
		data[k] = v
	}

	buf := bytes.NewBuffer(make([]byte, 0))
	err := msgpack.NewEncoder(buf).SortMapKeys(true).Encode(data)
	return buf.Bytes(), err
}

type StateID struct {
	UUID
}

func calcStateID(state *State) StateID {
	encoded, _ := msgpack.Marshal(struct {
		PatchID PatchID
		Recipes RecipeList
	}{
		state.PatchID,
		state.Recipes,
	})

	return StateID{NewUUID(encoded)}
}

type State struct {
	ID      StateID    `json:"id"`
	PatchID PatchID    `json:"patch_id"`
	Recipes RecipeList `json:"recipes"`
}

func NewState() *State {
	s := State{}
	s.Recipes = make(map[string]Recipe)
	s.ID = calcStateID(&s)
	return &s
}

func (s *State) String() string {
	return fmt.Sprintf("State[ID=%s Recipes=%d]", s.ID, len(s.Recipes))
}

func (s *State) UnmarshalMsgpack(raw []byte) error {
	if err := msgpack.Unmarshal(raw, s); err != nil {
		return err
	}

	id := calcStateID(s)
	if id != s.ID {
		return fmt.Errorf("invalid ID")
	}

	return nil
}

func (s *State) UnmarshalJSON(raw []byte) error {
	if err := json.Unmarshal(raw, s); err != nil {
		return err
	}

	id := calcStateID(s)
	if id != s.ID {
		return fmt.Errorf("invalid ID")
	}

	return nil
}

func (s *State) Apply(patch Patch) {
	for k, v := range patch.Recipes {
		if v == nil {
			delete(s.Recipes, k)
		} else {
			s.Recipes[k] = v
		}
	}

	s.PatchID = patch.ID
	s.ID = calcStateID(s)
}

type PatchID struct {
	UUID
}

func calcPatchID(patch Patch) PatchID {
	encoded, _ := msgpack.Marshal(struct {
		Previous PatchID
		Recipes  RecipeList
	}{
		patch.Previous,
		patch.Recipes,
	})

	return PatchID{NewUUID(encoded)}
}

type Patch struct {
	Previous PatchID    `json:"previous"`
	ID       PatchID    `json:"id"`
	Recipes  RecipeList `json:"recipes"`
}

func NewPatch(previous PatchID, recipes RecipeList) (Patch, error) {
	p := Patch{
		Previous: previous,
		Recipes:  recipes,
	}
	p.ID = calcPatchID(p)
	return p, nil
}

func (p Patch) AddedRecipes() int {
	r := 0
	for _, x := range p.Recipes {
		if x != nil {
			r++
		}
	}
	return r
}

func (p Patch) DeletedRecipes() int {
	r := 0
	for _, x := range p.Recipes {
		if x == nil {
			r++
		}
	}
	return r
}

func (p Patch) String() string {
	return fmt.Sprintf("Patch[ID=%s Previous=%s AddedRecipes=%d DeletedRecipes=%d]", p.ID, p.Previous, p.AddedRecipes(), p.DeletedRecipes())
}

func (p *Patch) UnmarshalMsgpack(raw []byte) error {
	if err := msgpack.Unmarshal(raw, p); err != nil {
		return err
	}

	id := calcPatchID(*p)
	if id != p.ID {
		return fmt.Errorf("invalid ID")
	}

	return nil
}

func (p *Patch) UnmarshalJSON(raw []byte) error {
	if err := json.Unmarshal(raw, p); err != nil {
		return err
	}

	id := calcPatchID(*p)
	if id != p.ID {
		return fmt.Errorf("invalid ID")
	}

	return nil
}

type PatchChain struct {
	chain []Patch
}

func (c *PatchChain) String() string {
	return fmt.Sprint(c.chain)
}

func (c *PatchChain) Has(id PatchID) bool {
	for _, p := range c.chain {
		if p.ID == id {
			return true
		}
	}
	return false
}

func (c *PatchChain) ApplyTo(state *State, id PatchID) error {
	if !c.Has(id) {
		return fmt.Errorf("unknown entry")
	}

	for i, p := range c.chain {
		state.Apply(p)

		if p.ID == id {
			c.chain = c.chain[:i]
			break
		}
	}

	return nil
}

func (c *PatchChain) Add(patch Patch) error {
	if len(c.chain) == 0 {
		c.chain = []Patch{patch}
		return nil
	}

	for i, p := range c.chain {
		if p.ID == patch.Previous {
			c.chain = append(c.chain[:i+1], patch)
			return nil
		}
	}

	return fmt.Errorf("not chained")
}

func (c *PatchChain) New(recipes RecipeList) (Patch, error) {
	prev := PatchID{}
	if len(c.chain) > 0 {
		prev = c.chain[len(c.chain)-1].ID
	}

	patch, err := NewPatch(prev, recipes)
	if err != nil {
		return Patch{}, err
	}
	c.chain = append(c.chain, patch)

	return patch, nil
}
