package main_test

import (
	"testing"

	cookfs "."
)

func Test_InMemoryChunkStore(t *testing.T) {
	m := cookfs.NewInMemoryChunkStore()

	chunk := cookfs.NewChunk(cookfs.Hash{}, []byte("hello"))

	if err := m.Save(chunk); err != nil {
		t.Errorf("failed to save chunk because; %s", err.Error())
	}

	if result, err := m.Load(chunk.Hash); err != nil {
		t.Errorf("failed to load chunk because; %s", err.Error())
	} else if result.Data != chunk.Data {
		t.Errorf("mismatched saved data and loaded data; except %s but got %v", "hello", result.Data)
	}

	if err := m.Delete(chunk.Hash); err != nil {
		t.Errorf("failed to delete chunk because; %s", err.Error())
	} else if _, err := m.Load(chunk.Hash); err != cookfs.ChunkNotFound {
		t.Errorf("InMemoryChunkStore.Delete was succeed but data is not deleted")
	}
}

func Test_InMemoryRecipieStore(t *testing.T) {
	m := cookfs.NewInMemoryRecipieStore()

	a := []cookfs.Hash{cookfs.CalcHash([]byte("hello")), cookfs.CalcHash([]byte("world"))}
	b := []cookfs.Hash{cookfs.CalcHash([]byte("hello")), cookfs.CalcHash([]byte("world")), cookfs.CalcHash([]byte("foobar"))}
	recipies := map[string][]cookfs.Hash{
		"/tag/of/foobar": a,
		"/tag/to/hogefuga": b,
	}

	for tag, recipie := range recipies {
		if err := m.Save(tag, recipie); err != nil {
			t.Errorf("failed to save recipie because; %s", err.Error())
		}
	}

	for tag, recipie := range recipies {
		if got, err := m.Load(tag); err != nil {
			t.Errorf("failed to load recipie because; %s", err.Error())
		} else if len(got) != len(recipie) {
			t.Errorf("failed to load recipie; excepted %#v but got %#v", recipie, got)
		} else {
			for i, x := range got {
				if x != recipie[i] {
					t.Errorf("failed to load recipie; excepted %#v but got %#v", recipie, got)
				}
			}
		}
	}

	found_tests := []struct {
		prefix string
		except []string
	} {
		{"/tag/", []string{"/tag/of/foobar", "/tag/to/hogefuga"}},
		{"/tag/of/", []string{"/tag/of/foobar"}},
	}
	for _, test := range found_tests {
		if founds, err := m.Find(test.prefix); err != nil {
			t.Errorf("failed to find tag because; %s", err.Error())
		} else if len(founds) != len(test.except) {
			t.Errorf("InMemoryRecipieStore.Find was returns unexcepted result; %#v", founds)
		} else {
			for _, x := range test.except {
				ok := false
				for _, y := range founds {
					if x == y {
						ok = true
						break
					}
				}
				if !ok {
					t.Errorf("InMemoryRecipieStore.Find was returns unexcepted result; %#v", founds)
					break
				}
			}
		}
	}
}
