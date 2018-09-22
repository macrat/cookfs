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
