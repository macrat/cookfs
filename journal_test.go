package main

import (
	"fmt"
	"testing"
)

func Test_calcEntryID(t *testing.T) {
	recipies := map[string]Recipie{
		"/foo": {CalcHash([]byte("hello"))},
		"/bar": {CalcHash([]byte("world"))},
	}

	except := CalcHash([]byte(
		fmt.Sprintf("{\"/bar\":[\"%s\"],\"/foo\":[\"%s\"]}", CalcHash([]byte("world")), CalcHash([]byte("hello"))),
	))
	got := calcEntryID(recipies)

	if got != except {
		t.Errorf("unexcepted entryID; excepted %s but got %s", except, got)
	}
}

func Test_calcChainID(t *testing.T) {
	hello := CalcHash([]byte("hello"))
	world := CalcHash([]byte("world"))

	except := CalcHash(hello[:], world[:])
	got := calcChainID(hello, world)

	if got != except {
		t.Errorf("unexcepted chainID; excepted %s but got %s", except, got)
	}
}

func Test_JournalEntry_IsPreviousOf(t *testing.T) {
	a1 := NewJournalEntry(nil, map[string]Recipie{})
	a2 := NewJournalEntry(a1, map[string]Recipie{})
	b1 := NewJournalEntry(nil, map[string]Recipie{})

	if !a1.IsPreviousOf(a2) {
		t.Errorf("a1 is previous of a2 but IsPreviousOf said that's not it")
	}

	if a2.IsPreviousOf(a1) {
		t.Errorf("a2 is not previous of a1 but IsPreviousOf said it is previous")
	}

	if a1.IsPreviousOf(b1) || b1.IsPreviousOf(a1) {
		t.Errorf("a1 and b1 is not chained but IsPreviousOf said it is chained")
	}

	if !(*JournalEntry)(nil).IsPreviousOf(a1) {
		t.Errorf("nil is previous of a1 but IsPreviousOf said that's not it")
	}

	if (*JournalEntry)(nil).IsPreviousOf(a2) {
		t.Errorf("nil is not previous of a2 but IsPreviousOf said it is previous")
	}
}

func Test_JournalEntry_Join(t *testing.T) {
	a1 := NewJournalEntry(nil, map[string]Recipie{})
	a2 := NewJournalEntry(a1, map[string]Recipie{})
	a2.Previous = nil

	if err := a2.Join(a1); err == nil {
		t.Errorf("a1 is not next of a2 but join succeed")
	}

	if err := a1.Join(a2); err != nil {
		t.Errorf("failed join a2 to after of a1; %s", err.Error())
	}
}

func Test_JournalManager(t *testing.T) {
	jm := JournalManager{}

	err := jm.AddRecipies(map[string]Recipie{
		"/tag/to/one": {CalcHash([]byte("hello"))},
	})
	if err != nil {
		t.Errorf("failed to add recipies into JournalManager; %s", err.Error())
	}
	if jm.Dirty == nil {
		t.Errorf("failed to add recipies into JournalManager; Dirty is nil")
	}
	if jm.Head != nil {
		t.Errorf("not committed yet but Head is already not nil")
	}

	err = jm.AddRecipies(map[string]Recipie{
		"/tag/to/two": {CalcHash([]byte("world"))},
	})
	if err != nil {
		t.Errorf("failed to add recipies into JournalManager; %s", err.Error())
	}
	if jm.Dirty == nil {
		t.Errorf("failed to add recipies into JournalManager; Dirty is nil")
	}
	if jm.Head != nil {
		t.Errorf("not committed yet but Head is already not nil")
	}

	err = jm.AddEntry(NewJournalEntry(nil, map[string]Recipie{
		"/tag/to/not-chained": {CalcHash([]byte("foobar"))},
	}))
	if err == nil {
		t.Errorf("succeed to add not chained entry")
	} else if err != JournalIsNotChainedError {
		t.Errorf("couses unexcepted error on adding not chained entry: %s", err.Error())
	}

	err = jm.AddEntry(NewJournalEntry(jm.Dirty.Previous, map[string]Recipie{
		"/tag/to/three": {CalcHash([]byte("world"))},
	}))
	if err != nil {
		t.Errorf("failed to add recipies into JournalManager; %s", err.Error())
	}
	if jm.Dirty == nil {
		t.Errorf("failed to add recipies into JournalManager; Dirty is nil")
	}
	if jm.Head != nil {
		t.Errorf("not committed yet but Head is already not nil")
	}

	if _, ok := jm.Dirty.Recipies["/tag/to/three"]; !ok {
		t.Errorf("added entry was not found; found recipies is %v", jm.Dirty.Recipies)
	}
	if _, ok := jm.Dirty.Previous.Recipies["/tag/to/one"]; !ok {
		t.Errorf("added entry was not found; found recipies is %v", jm.Dirty.Previous.Recipies)
	}

	err = jm.Commit(Hash{})
	if err == nil {
		t.Errorf("succeed to commit with invalid hash")
	} else if err != NoSuchJournalError {
		t.Errorf("couses unexcepted error on commiting with invalid hash: %s", err.Error())
	}

	err = jm.Commit(jm.Dirty.Previous.ChainID)
	if err != nil {
		t.Errorf("failed to commit; %s", err.Error())
	}
	if jm.Head != jm.Dirty.Previous {
		t.Errorf("commit succeed but Head is not updated; got %v", jm.Head)
	}

	err = jm.Commit(jm.Head.ChainID)
	if err == nil {
		t.Errorf("succeed to commit the same journal twice")
	} else if err != JournalAlreadyCommittedError {
		t.Errorf("couses unexcepted error on commiting the same journal twice; %s", err.Error())
	}
}
