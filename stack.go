// Copyright (c) HashiCorp, Inc.

package mql

type stack[T any] struct {
	data []T
}

func (s *stack[T]) push(v T) {
	s.data = append(s.data, v)
}

func (s *stack[T]) pop() (T, bool) {
	var x T
	if len(s.data) > 0 {
		x, s.data = s.data[len(s.data)-1], s.data[:len(s.data)-1]
		return x, true
	}
	return x, false
}

func (s *stack[T]) clear() {
	s.data = nil
}

func runesToString(s stack[rune]) string {
	var result string
	for {
		r, ok := s.pop()
		if !ok {
			break
		}
		result = string(r) + result
	}
	return result
}
