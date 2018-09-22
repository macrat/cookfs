package main

import (
	"encoding/json"
	"fmt"
)

var (
	RecipiesIsEmptyError         = fmt.Errorf("can't add empty recipies")
	JournalAlreadyCommittedError = fmt.Errorf("journal entry was already committed")
	NoSuchJournalError           = fmt.Errorf("no such journal entry entry")
	JournalIsNotChainedError     = fmt.Errorf("jurnal entry is not chained")
)

type JournalEntry struct {
	Previous *JournalEntry
	EntryID  Hash
	ChainID  Hash

	Recipies map[string]Recipie
}

func calcEntryID(recipies map[string]Recipie) Hash {
	j, _ := json.Marshal(recipies)
	return CalcHash(j)
}

func calcChainID(previousChainID, nextEntryID Hash) Hash {
	return CalcHash(previousChainID[:], nextEntryID[:])
}

func NewJournalEntry(previous *JournalEntry, recipies map[string]Recipie) *JournalEntry {
	entryID := calcEntryID(recipies)

	prevID := Hash{}
	if previous != nil {
		prevID = previous.ChainID
	}
	chainID := calcChainID(prevID, entryID)

	return &JournalEntry{
		Previous: previous,
		EntryID:  entryID,
		ChainID:  chainID,
		Recipies: recipies,
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
	PreviousID Hash               `json:"previous_id"`
	EntryID    Hash               `json:"entry_id"`
	ChainID    Hash               `json:"chain_id"`
	Recipies   map[string]Recipie `json:"recipies"`
}

func (j *JournalEntry) MarshalJSON() ([]byte, error) {
	x := jsonJournalEntry{
		EntryID:  j.EntryID,
		ChainID:  j.ChainID,
		Recipies: j.Recipies,
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
	j.Recipies = x.Recipies

	return nil
}

type JournalManager struct {
	Head  *JournalEntry
	Dirty *JournalEntry
}

func (j *JournalManager) AddEntry(entry *JournalEntry) error {
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

func (j *JournalManager) AddRecipies(recipies map[string]Recipie) error {
	if recipies == nil || len(recipies) == 0 {
		return fmt.Errorf("can't add empty recipies")
	}

	return j.AddEntry(NewJournalEntry(j.Dirty, recipies))
}

func (j *JournalManager) Commit(chainID Hash) error {
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

func (j *JournalManager) GetDirty() []*JournalEntry {
	result := []*JournalEntry{}

	x := j.Dirty
	for x != j.Head {
		result = append(result, x)

		x = x.Previous
	}

	return result
}

func (j *JournalManager) GetCommitted(num int) []*JournalEntry {
	result := make([]*JournalEntry, 0, num)

	x := j.Head
	for i := 0; i < num && x != nil; i++ {
		result = append(result, x)

		x = x.Previous
	}

	return result
}
