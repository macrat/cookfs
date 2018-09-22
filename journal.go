package main

import (
	"encoding/json"
	"fmt"
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
	return next.ChainID == calcChainID(j.ChainID, next.EntryID)
}

func (j *JournalEntry) Join(next *JournalEntry) error {
	if !j.IsPreviousOf(next) {
		return fmt.Errorf("entry:%s is not next entry of entry:%s", next.ChainID.ShortHash(), j.ChainID.ShortHash())
	}

	next.Previous = j

	return nil
}
