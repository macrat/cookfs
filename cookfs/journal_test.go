package cookfs

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test_calcEntryID(t *testing.T) {
	recipes := map[string]Recipe{
		"/foo": {CalcHash([]byte("hello"))},
		"/bar": {CalcHash([]byte("world"))},
	}

	except := CalcHash([]byte(
		fmt.Sprintf("{\"/bar\":[\"%s\"],\"/foo\":[\"%s\"]}", CalcHash([]byte("world")), CalcHash([]byte("hello"))),
	))
	got := calcEntryID(recipes)

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
	a1 := NewJournalEntry(nil, map[string]Recipe{})
	a2 := NewJournalEntry(a1, map[string]Recipe{})
	b1 := NewJournalEntry(nil, map[string]Recipe{})

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
	a1 := NewJournalEntry(nil, map[string]Recipe{})
	a2 := NewJournalEntry(a1, map[string]Recipe{})
	a2.Previous = nil

	if err := a2.Join(a1); err == nil {
		t.Errorf("a1 is not next of a2 but join succeed")
	}

	if err := a1.Join(a2); err != nil {
		t.Errorf("failed join a2 to after of a1; %s", err.Error())
	}
}

func Test_JournalEntry_Json(t *testing.T) {
	x := NewJournalEntry(nil, map[string]Recipe{})

	j, err := json.Marshal(x)
	if err != nil {
		t.Errorf("failed to marshal json: %s", err.Error())
	}

	y := &JournalEntry{}
	if err = json.Unmarshal(j, y); err != nil {
		t.Errorf("failed to unmarshal json: %s", err.Error())
	}

	if x.ChainID != y.ChainID {
		t.Errorf("x.ChainID(%v) != y.ChainID(%v)", x.ChainID, y.ChainID)
	}
	if x.EntryID != y.EntryID {
		t.Errorf("x.EntryID(%v) != y.EntryID(%v)", x.EntryID, y.EntryID)
	}
}

func Test_JournalChain(t *testing.T) {
	jc := JournalChain{}

	err := jc.AddRecipes(map[string]Recipe{
		"/tag/to/one": {CalcHash([]byte("hello"))},
	})
	if err != nil {
		t.Errorf("failed to add recipes into JournalManager; %s", err.Error())
	}
	if jc.Dirty == nil {
		t.Errorf("failed to add recipes into JournalManager; Dirty is nil")
	}
	if jc.Head != nil {
		t.Errorf("not committed yet but Head is already not nil")
	}

	err = jc.AddRecipes(map[string]Recipe{
		"/tag/to/two": {CalcHash([]byte("world"))},
	})
	if err != nil {
		t.Errorf("failed to add recipes into JournalManager; %s", err.Error())
	}
	if jc.Dirty == nil {
		t.Errorf("failed to add recipes into JournalManager; Dirty is nil")
	}
	if jc.Head != nil {
		t.Errorf("not committed yet but Head is already not nil")
	}

	err = jc.AddEntry(NewJournalEntry(nil, map[string]Recipe{
		"/tag/to/not-chained": {CalcHash([]byte("foobar"))},
	}))
	if err == nil {
		t.Errorf("succeed to add not chained entry")
	} else if err != JournalIsNotChainedError {
		t.Errorf("couses unexcepted error on adding not chained entry: %s", err.Error())
	}

	err = jc.AddEntry(NewJournalEntry(jc.Dirty.Previous, map[string]Recipe{
		"/tag/to/three": {CalcHash([]byte("world"))},
	}))
	if err != nil {
		t.Errorf("failed to add recipes into JournalManager; %s", err.Error())
	}
	if jc.Dirty == nil {
		t.Errorf("failed to add recipes into JournalManager; Dirty is nil")
	}
	if jc.Head != nil {
		t.Errorf("not committed yet but Head is already not nil")
	}

	if _, ok := jc.Dirty.Recipes["/tag/to/three"]; !ok {
		t.Errorf("added entry was not found; found recipes is %v", jc.Dirty.Recipes)
	}
	if _, ok := jc.Dirty.Previous.Recipes["/tag/to/one"]; !ok {
		t.Errorf("added entry was not found; found recipes is %v", jc.Dirty.Previous.Recipes)
	}

	if len(jc.GetCommitted(10)) != 0 {
		t.Errorf("committed entries length is not excepted value; it was %d", len(jc.GetCommitted(10)))
	}
	if len(jc.GetDirty()) != 2 {
		t.Errorf("dirty entries length is not excepted value; it was %d", len(jc.GetDirty()))
	}

	err = jc.Commit(Hash{})
	if err == nil {
		t.Errorf("succeed to commit with invalid hash")
	} else if err != NoSuchJournalError {
		t.Errorf("couses unexcepted error on commiting with invalid hash: %s", err.Error())
	}

	err = jc.Commit(jc.Dirty.Previous.ChainID)
	if err != nil {
		t.Errorf("failed to commit; %s", err.Error())
	}
	if jc.Head != jc.Dirty.Previous {
		t.Errorf("commit succeed but Head is not updated; got %v", jc.Head)
	}

	err = jc.Commit(jc.Head.ChainID)
	if err == nil {
		t.Errorf("succeed to commit the same journal twice")
	} else if err != JournalAlreadyCommittedError {
		t.Errorf("couses unexcepted error on commiting the same journal twice; %s", err.Error())
	}

	if len(jc.GetCommitted(10)) != 1 {
		t.Errorf("committed entries length is not excepted value; it was %d", len(jc.GetCommitted(10)))
	}
	if len(jc.GetDirty()) != 1 {
		t.Errorf("dirty entries length is not excepted value; it was %d", len(jc.GetDirty()))
	}
}
