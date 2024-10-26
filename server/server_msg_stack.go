package server

import "sync"

type stack struct {
	vals []int
	mu   sync.Mutex
}

func newStack() *stack {
	return &stack{
		vals: []int{},
	}
}

// push value onto the stack
func (s *stack) push(val int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vals = append(s.vals, val)
}

// pop value from the stack
func (s *stack) pop() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	l := len(s.vals)
	top := s.vals[l-1]
	s.vals = s.vals[:l-1]
	return top

}
