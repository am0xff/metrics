package storage

type Storage[T any] struct {
	data map[string]T
}

func NewStorage[T any]() *Storage[T] {
	return &Storage[T]{
		data: map[string]T{},
	}
}

func (s *Storage[T]) Get(key string) (T, bool) {
	v, ok := s.data[key]
	return v, ok
}

func (s *Storage[T]) Set(key string, value T) {
	s.data[key] = value
}

func (s *Storage[T]) Keys() []string {
	keys := make([]string, 0, len(s.data))

	for k := range s.data {
		keys = append(keys, k)
	}

	return keys
}
