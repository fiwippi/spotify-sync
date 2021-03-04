package sets

import (
	"fmt"
	"strings"
)

var exists = struct{}{}

type Set struct {
	// Structs take zero bytes unlike bools
	m map[string]struct{}
}

// Creates a new Set
func NewSet() *Set {
	s := &Set{}
	s.m = make(map[string]struct{})
	return s
}

// Adds an item to the Set
func (s *Set) Add(values ...string) {
	for _, v := range values {
		s.m[v] = exists
	}
}

// Removes all items from the set.
func (s *Set) Clear() {
	s.m = make(map[string]struct{})
}

// Creates and returns a copy of a set
func (s *Set) Copy() *Set {
	ns := NewSet()
	for x := range s.m {
		ns.Add(x)
	}

	return ns
}

// Checks if a value is in the Set, returns true if it exists
func (s *Set) Has(value string) bool {
	_, c := s.m[value]
	return c
}

// Returns whether the Set is empty
func (s *Set) IsEmpty() bool {
	return s.Size() == 0
}

// Returns the Set as a list
func (s *Set) List() []string {
	list := make([]string, 0, len(s.m))

	for item := range s.m {
		list = append(list, item)
	}

	return list
}

// Pop removes and returns item from Set. The underlying Set is modified.
// "" (empty string) is returned if the item does not exist.
func (s *Set) Pop() string {
	for item := range s.m {
		delete(s.m, item)
		return item
	}
	return ""
}

// Removes an item from the Set
func (s *Set) Remove(value string) {
	delete(s.m, value)
}

// Returns the length of the Set as int
func (s *Set) Size() int {
	return len(s.m)
}

// Returns a string representation of the Set
func (s *Set) String() string {
	t := make([]string, 0, len(s.List()))
	for _, item := range s.List() {
		t = append(t, fmt.Sprintf("%v", item))
	}

	return fmt.Sprintf("[%s]", strings.Join(t, ", "))
}
