package set

import (
	"sync"
)

type Set[T comparable] struct {
	elements map[T]struct{}
	mu       sync.RWMutex
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		elements: make(map[T]struct{}),
	}
}

func (s *Set[T]) Add(value T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.elements[value] = struct{}{}
}

func (s *Set[T]) Remove(value T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.elements, value)
}

func (s *Set[T]) Contains(value T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.elements[value]
	return exists
}

func (s *Set[T]) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.elements)
}

func (s *Set[T]) Values() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]T, 0, len(s.elements))
	for key := range s.elements {
		values = append(values, key)
	}
	return values
}
