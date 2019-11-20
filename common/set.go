package common

type StringSet struct {
	m map[string]int
}

func NewStringSet(values ...string) *StringSet {
	s := new(StringSet)
	s.m = make(map[string]int)
	for i, value := range values {
		s.m[value] = i + 1
	}
	return s
}

func (s *StringSet) Add(value string) {
	s.m[value] = len(s.m) + 1
}

func (s *StringSet) Remove(value string) {
	delete(s.m, value)
}

func (s *StringSet) Contains(value string) bool {
	return s.Get(value) > 0
}

func (s *StringSet) Get(value string) int {
	return s.m[value]
}
