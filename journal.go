package main

import (
	"fmt"
)

var (
	InvalidEntryIDError = fmt.Errorf("invalid entry ID")
	IDNotFoundError     = fmt.Errorf("ID not found error")
)

type JournalEntry struct {
	ID   int64
	Data interface{}
	Next *JournalEntry
}

type JournalManager struct {
	journal *JournalEntry
	head    *JournalEntry
}

func (jm *JournalManager) Journal() *JournalEntry {
	return jm.journal
}

func (jm *JournalManager) Head() *JournalEntry {
	return jm.head
}

func (jm *JournalManager) Dirty() *JournalEntry {
	j := jm.head
	if j == nil {
		j = jm.journal
		if j == nil {
			return nil
		}
	}

	for j.Next != nil {
		j = j.Next
	}

	return j
}

func (jm *JournalManager) GetDirty(id int64) *JournalEntry {
	j := jm.head
	if j == nil {
		j = jm.journal
	}

	for j != nil && j.ID != id {
		j = j.Next
	}

	return j
}

func (jm *JournalManager) AddDirtyEntries(entry *JournalEntry) error {
	j := jm.GetDirty(entry.ID - 1)
	if j != nil {
		j.Next = entry
		return nil
	}

	last := jm.Dirty()
	if last == nil {
		jm.journal = entry
		return nil
	}

	if last.ID+1 != entry.ID {
		return InvalidEntryIDError
	}

	last.Next = entry

	return nil
}

func (jm *JournalManager) Commit(id int64) error {
	d := jm.GetDirty(id)
	if d == nil {
		return IDNotFoundError
	}

	jm.head = d
	return nil
}

func (jm *JournalManager) CommittedID() int64 {
	if head := jm.Head(); head == nil {
		return 0
	} else {
		return jm.Head().ID
	}
}
