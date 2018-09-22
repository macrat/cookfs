package cookfs

import (
	"fmt"
)

type Term struct {
	ID     int64 `json:"id"`
	Leader *Node `json:"leader"`
}

func (t Term) String() string {
	return fmt.Sprintf("[Term %d](%s)", t.ID, t.Leader)
}

func (t Term) Equals(another Term) bool {
	return t.ID == another.ID && t.Leader.Equals(another.Leader)
}

func (t Term) NewerThan(another Term) bool {
	return t.ID > another.ID
}

func (t Term) OlderThan(another Term) bool {
	return t.ID < another.ID
}

type TermStatus struct {
	Term

	JournalID Hash `json:"journal"`
}
