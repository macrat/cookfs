package main

type DummyStore struct{}

func (ds DummyStore) Bind(c *CookFS) {
}

func (ds DummyStore) Run(chan struct{}) error {
	return nil
}

func (ds DummyStore) Save(h Hash, b []byte) error {
	return nil
}

func (ds DummyStore) Load(h Hash) ([]byte, error) {
	return nil, nil
}

func (ds DummyStore) Delete(h Hash) error {
	return nil
}
