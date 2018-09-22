package main

import (
	"encoding/json"
	"fmt"
	"sync"
)

var (
	RecipesIsEmptyError          = fmt.Errorf("can't add empty recipes")
	JournalAlreadyCommittedError = fmt.Errorf("journal entry was already committed")
	NoSuchJournalError           = fmt.Errorf("no such journal entry entry")
	JournalIsNotChainedError     = fmt.Errorf("jurnal entry is not chained")
)

type JournalEntry struct {
	Previous *JournalEntry
	EntryID  Hash
	ChainID  Hash

	Recipes map[string]Recipe
}

func calcEntryID(recipes map[string]Recipe) Hash {
	j, _ := json.Marshal(recipes)
	return CalcHash(j)
}

func calcChainID(previousChainID, nextEntryID Hash) Hash {
	return CalcHash(previousChainID[:], nextEntryID[:])
}

func NewJournalEntry(previous *JournalEntry, recipes map[string]Recipe) *JournalEntry {
	entryID := calcEntryID(recipes)

	prevID := Hash{}
	if previous != nil {
		prevID = previous.ChainID
	}
	chainID := calcChainID(prevID, entryID)

	return &JournalEntry{
		Previous: previous,
		EntryID:  entryID,
		ChainID:  chainID,
		Recipes:  recipes,
	}
}

func (j *JournalEntry) IsPreviousOf(next *JournalEntry) bool {
	var prevID Hash
	if j != nil {
		prevID = j.ChainID
	}
	return next.ChainID == calcChainID(prevID, next.EntryID)
}

func (j *JournalEntry) Join(next *JournalEntry) error {
	if !j.IsPreviousOf(next) {
		return fmt.Errorf("entry:%s is not next entry of entry:%s", next.ChainID.ShortHash(), j.ChainID.ShortHash())
	}

	next.Previous = j

	return nil
}

type jsonJournalEntry struct {
	PreviousID Hash              `json:"previous_id"`
	EntryID    Hash              `json:"entry_id"`
	ChainID    Hash              `json:"chain_id"`
	Recipes    map[string]Recipe `json:"recipes"`
}

func (j *JournalEntry) MarshalJSON() ([]byte, error) {
	x := jsonJournalEntry{
		EntryID: j.EntryID,
		ChainID: j.ChainID,
		Recipes: j.Recipes,
	}

	if j.Previous != nil {
		x.PreviousID = j.Previous.ChainID
	}

	return json.Marshal(x)
}

func (j *JournalEntry) UnmarshalJSON(raw []byte) error {
	var x jsonJournalEntry

	if err := json.Unmarshal(raw, &x); err != nil {
		return err
	}

	if CalcHash(x.PreviousID[:], x.EntryID[:]) != x.ChainID {
		return fmt.Errorf("broken ID")
	}

	j.EntryID = x.EntryID
	j.ChainID = x.ChainID
	j.Recipes = x.Recipes

	return nil
}

type JournalChain struct {
	sync.Mutex

	Head  *JournalEntry
	Dirty *JournalEntry

	transmit TransmitPlugin
}

func (j *JournalChain) AddEntry(entry *JournalEntry) error {
	j.Lock()
	defer j.Unlock()

	if j.Dirty == nil {
		if err := j.Dirty.Join(entry); err != nil {
			return err
		}
		j.Dirty = entry
		return nil
	}

	dirty := j.Dirty

	stop := j.Head
	if stop != nil {
		stop = stop.Previous
	}

	for dirty != stop {
		if err := dirty.Join(entry); err == nil {
			j.Dirty = entry
			return nil
		}

		dirty = dirty.Previous
	}

	return JournalIsNotChainedError
}

func (j *JournalChain) AddRecipes(recipes map[string]Recipe) error {
	if recipes == nil || len(recipes) == 0 {
		return fmt.Errorf("can't add empty recipes")
	}

	return j.AddEntry(NewJournalEntry(j.Dirty, recipes))
}

func (j *JournalChain) Drop(chainID Hash) error {
	j.Lock()
	defer j.Unlock()

	x := j.Dirty

	for x != j.Head {
		if x.ChainID == chainID {
			j.Dirty = x.Previous
			return nil
		}

		x = x.Previous
	}

	return NoSuchJournalError
}

func (j *JournalChain) Commit(chainID Hash) error {
	j.Lock()
	defer j.Unlock()

	x := j.Dirty

	for x != j.Head {
		if x.ChainID == chainID {
			j.Head = x
			return nil
		}

		x = x.Previous
	}

	for x != nil {
		if x.ChainID == chainID {
			return JournalAlreadyCommittedError
		}

		x = x.Previous
	}

	return NoSuchJournalError
}

func (j *JournalChain) GetDirty() []*JournalEntry {
	result := []*JournalEntry{}

	x := j.Dirty
	for x != j.Head {
		result = append([]*JournalEntry{x}, result...)

		x = x.Previous
	}

	return result
}

func (j *JournalChain) GetCommitted(num int) []*JournalEntry {
	result := []*JournalEntry{}

	x := j.Head
	for i := 0; i < num && x != nil; i++ {
		result = append([]*JournalEntry{x}, result...)

		x = x.Previous
	}

	return result
}

type Journal struct {
	chain *JournalChain

	polling  *Polling
	recipe   RecipePlugin
	transmit TransmitPlugin
}

func NewJournal() *Journal {
	return &Journal{
		chain: &JournalChain{},
	}
}

func (j *Journal) Run(chan struct{}) error {
	return nil
}

func (j *Journal) Bind(c *CookFS) {
	j.polling = c.Polling
	j.recipe = c.Recipe
	j.transmit = c.Transmit
}

type JournalLog struct {
	Committed []*JournalEntry `json:"committed"`
	Dirty     []*JournalEntry `json:"dirty"`
}

func (j *Journal) GetLog() JournalLog {
	return JournalLog{
		Committed: j.chain.GetCommitted(20),
		Dirty:     j.chain.GetDirty(),
	}
}

func (j *Journal) AddEntry(entry *JournalEntry) error {
	return j.chain.AddEntry(entry)
}

func (j *Journal) Commit(chainID Hash) error {
	old_head := j.chain.Head

	if err := j.chain.Commit(chainID); err != nil {
		return err
	}

	var entries []*JournalEntry
	for x := j.chain.Head; x != old_head; x = x.Previous {
		entries = append([]*JournalEntry{x}, entries...)
	}

	for _, x := range entries {
		for tag, recipe := range x.Recipes {
			if err := j.recipe.Save(tag, recipe); err != nil {
				return err
			}
		}
	}

	return nil
}

func (j *Journal) AddRecipe(tag string, recipe Recipe) error {
	if !j.polling.IsLeader() {
		return fmt.Errorf("I'm not the leader")
	}

	entry := NewJournalEntry(j.chain.Dirty, map[string]Recipe{tag: recipe})

	if !j.transmit.AddJournalEntry(entry) {
		return fmt.Errorf("denied")
	}

	if !j.transmit.CommitJournal(entry.ChainID) {
		return fmt.Errorf("denied")
	}

	return nil
}

func (j *Journal) IsCommitted(h Hash) bool {
	for x := j.chain.Head; x != nil; x = x.Previous {
		if x.ChainID == h {
			return true
		}
	}
	return false
}

func (j *Journal) IsDirty(h Hash) bool {
	if j.chain.Dirty == nil {
		return false
	}

	for x := j.chain.Dirty; x != j.chain.Head; x = x.Previous {
		if x.ChainID == h {
			return true
		}
	}

	return false
}

func (j *Journal) HeadID() Hash {
	if j.chain.Head == nil {
		return Hash{}
	} else {
		return j.chain.Head.ChainID
	}
}
