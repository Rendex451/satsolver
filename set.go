package main

type Set struct {
	elements map[Literal]struct{}
}

func NewSet() *Set {
	return &Set{
		elements: make(map[Literal]struct{}),
	}
}

func (s *Set) Add(value Literal) {
	s.elements[value] = struct{}{}
}

func (s *Set) Remove(value Literal) {
	delete(s.elements, value)
}

func (s *Set) Contains(value Literal) bool {
	_, exists := s.elements[value]
	return exists
}

func (s *Set) Size() int {
	return len(s.elements)
}

func (s *Set) Values() []Literal {
	values := make([]Literal, 0, len(s.elements))
	for key := range s.elements {
		values = append(values, key)
	}
	return values
}
