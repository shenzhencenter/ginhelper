package ginhelper

import "strings"

type StringSet struct {
	stringSet map[string]struct{}
}

func NewStringSet(opts ...string) *StringSet {
	set := &StringSet{
		stringSet: make(map[string]struct{}),
	}
	for _, opt := range opts {
		set.Add(opt)
	}
	return set
}

func (s *StringSet) Add(str string) {
	s.stringSet[str] = struct{}{}
}

func (s *StringSet) Contains(str string) bool {
	_, ok := s.stringSet[str]
	return ok
}

func (s *StringSet) Remove(str string) {
	delete(s.stringSet, str)
}

func (s *StringSet) Clear() {
	s.stringSet = make(map[string]struct{})
}

func (s *StringSet) Size() int {
	return len(s.stringSet)
}

func (s *StringSet) IsEmpty() bool {
	return s.Size() == 0
}

func (s *StringSet) ToSlice() []string {
	slice := make([]string, 0, s.Size())
	for str := range s.stringSet {
		slice = append(slice, str)
	}
	return slice
}

func (s *StringSet) MatchesPrefix(str string) bool {
	for k := range s.stringSet {
		if strings.HasPrefix(str, k) {
			return true
		}
	}
	return false
}
