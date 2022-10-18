package sets

import "sync"

type SyncSet[K comparable] struct {
	m    map[K]struct{}
	lock sync.Mutex
}

func NewSyncSet[K comparable]() *SyncSet[K] {
	return &SyncSet[K]{
		m: make(map[K]struct{}),
	}
}

func FromSlice[K comparable](slice []K) *SyncSet[K] {
	set := NewSyncSet[K]()

	for _, k := range slice {
		set.Add(k)
	}

	return set
}

func (s *SyncSet[K]) Add(k K) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.m[k] = struct{}{}
}

func (s *SyncSet[K]) Delete(k K) int {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.m, k)

	return len(s.m)
}

func (s *SyncSet[K]) Size(k K) int {
	s.lock.Lock()
	defer s.lock.Unlock()

	return len(s.m)
}

func (s *SyncSet[K]) All() []K {
	s.lock.Lock()
	defer s.lock.Unlock()

	r := make([]K, 0, len(s.m))
	for key := range s.m {
		r = append(r, key)
	}

	return r
}
