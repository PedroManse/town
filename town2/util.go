package town2

import (
	"sync"
)

type syncMap[K comparable, V any] struct {
	mx sync.Mutex
	mp map[K]V
}

func (s *syncMap[K, V]) Store(k K, v V) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.mp[k] = v
}

func (s *syncMap[K, V]) Load(k K) (V, bool) {
	s.mx.Lock()
	defer s.mx.Unlock()
	val, ok := s.mp[k]
	return val, ok
}
