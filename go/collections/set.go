package collections

import (
	"cmp"
	"slices"
)

type Set[T cmp.Ordered] map[T]struct{}

func (s Set[T]) Add(e T) {
	s[e] = struct{}{}
}

func (s Set[T]) Contains(e T) bool {
	_, ok := s[e]
	return ok
}

func (s Set[T]) Slice() []T {
	slice := make([]T, 0, len(s))
	for e := range s {
		slice = append(slice, e)
	}
	slices.Sort(slice)

	return slice
}

func NewSet[T cmp.Ordered](elements ...T) Set[T] {
	s := make(Set[T], len(elements))
	for _, element := range elements {
		s.Add(element)
	}
	return s
}
