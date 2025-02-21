package set

type Set[T comparable] struct {
	values map[T]bool
}

func New[T comparable]() *Set[T] {
	return &Set[T]{
		values: map[T]bool{},
	}
}

func (s *Set[T]) Add(val T) {
	s.values[val] = true
}

func (s *Set[T]) Contains(val T) bool {
	return s.values[val]
}
