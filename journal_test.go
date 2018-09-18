package main_test

import (
	"testing"

	cookfs "."
)

func Test_JournalManager(t *testing.T) {
	jm := &cookfs.JournalManager{}

	if err := jm.AddDirtyEntries(&cookfs.JournalEntry{0, "hello world", nil}); err != nil {
		t.Errorf("failed add entry: %s", err.Error())
	}

	if err := jm.AddDirtyEntries(&cookfs.JournalEntry{1, "this is test", nil}); err != nil {
		t.Errorf("failed add entry: %s", err.Error())
	}

	if err := jm.AddDirtyEntries(&cookfs.JournalEntry{5, "invalid", nil}); err != cookfs.InvalidEntryIDError {
		t.Errorf("succeed push invalid entry")
	}

	if jm.Dirty() == nil {
		t.Errorf("failed to get dirty head")
	} else if jm.Dirty().ID != 1 {
		t.Errorf("dirty head ID excepted %d but got %d", 1, jm.Dirty().ID)
	} else if jm.Dirty().Data != "this is test" {
		t.Errorf("dirty head data excepted %#v but got %#v", "this is test", jm.Dirty().Data)
	}

	if err := jm.AddDirtyEntries(&cookfs.JournalEntry{1, "foo bar", nil}); err != nil {
		t.Errorf("failed add entry: %s", err.Error())
	}

	if jm.Dirty() == nil {
		t.Errorf("failed to get dirty head")
	} else if jm.Dirty().ID != 1 {
		t.Errorf("dirty head ID excepted %d but got %d", 1, jm.Dirty().ID)
	} else if jm.Dirty().Data != "foo bar" {
		t.Errorf("dirty head data excepted %#v but got %#v", "foo bar", jm.Dirty().Data)
	}

	if jm.CommittedID() != 0 {
		t.Errorf("commited ID excepted %d but got %d", 0, jm.CommittedID())
	}

	jm.Commit(0)

	if jm.CommittedID() != 0 {
		t.Errorf("commited ID excepted %d but got %d", 0, jm.CommittedID())
	}

	if jm.Head() == nil {
		t.Errorf("failed to get head")
	} else if jm.Head().ID != 0 {
		t.Errorf("head ID excepted %d but got %d", 0, jm.Head().ID)
	} else if jm.Head().Data != "hello world" {
		t.Errorf("dirty head data excepted %#v but got %#v", "hello world", jm.Head().Data)
	}

	jm.Commit(1)

	if jm.CommittedID() != 1 {
		t.Errorf("commited ID excepted %d but got %d", 1, jm.CommittedID())
	}

	if jm.Head() == nil {
		t.Errorf("failed to get head")
	} else if jm.Head().ID != 1 {
		t.Errorf("head ID excepted %d but got %d", 1, jm.Head().ID)
	} else if jm.Head().Data != "foo bar" {
		t.Errorf("dirty head data excepted %#v but got %#v", "foo bar", jm.Head().Data)
	}
}
