package cookfs

import (
	"testing"
)

func Test_InMemoryChunkStore(t *testing.T) {
	m := NewInMemoryChunkStore()

	chunk := NewChunk(Hash{}, []byte("hello"))

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
	} else if _, err := m.Load(chunk.Hash); err != ChunkNotFound {
		t.Errorf("InMemoryChunkStore.Delete was succeed but data is not deleted")
	}
}

func Test_InMemoryRecipeStore(t *testing.T) {
	m := NewInMemoryRecipeStore()

	a := Recipe{CalcHash([]byte("hello")), CalcHash([]byte("world"))}
	b := Recipe{CalcHash([]byte("hello")), CalcHash([]byte("world")), CalcHash([]byte("foobar"))}
	recipes := []struct {
		tag    string
		recipe Recipe
	}{
		{"/tag/of/foobar", a},
		{"/tag/to/hogefuga", b},
	}

	for _, x := range recipes {
		if err := m.Save(x.tag, x.recipe); err != nil {
			t.Errorf("failed to save recipe because; %s", err.Error())
		}
	}

	for _, x := range recipes {
		if got, err := m.Load(x.tag); err != nil {
			t.Errorf("failed to load recipe because; %s", err.Error())
		} else if len(got) != len(x.recipe) {
			t.Errorf("failed to load recipe; excepted data is %#v but got %#v", x.recipe, got)
		} else {
			for i, y := range got {
				if y != x.recipe[i] {
					t.Errorf("failed to load recipe; excepted data is %#v but got %#v", x.recipe, got)
				}
			}
		}
	}

	found_tests := []struct {
		prefix string
		except []string
	}{
		{"/tag/", []string{"/tag/of/foobar", "/tag/to/hogefuga"}},
		{"/tag/of/", []string{"/tag/of/foobar"}},
	}
	for _, test := range found_tests {
		if founds, err := m.Find(test.prefix); err != nil {
			t.Errorf("failed to find tag because; %s", err.Error())
		} else if len(founds) != len(test.except) {
			t.Errorf("InMemoryRecipeStore.Find(%#v) was returns unexcepted result; %#v", test.prefix, founds)
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
					t.Errorf("InMemoryRecipeStore.Find(%#v) was returns unexcepted result; %#v", test.prefix, founds)
					break
				}
			}
		}
	}
}
