package util

import (
	"reflect"
)

type SliceSearch[T any] struct {
	slice []T
	ind   int
}

func NewSliceSearch[T any](slice []T) *SliceSearch[T] {
	return &SliceSearch[T]{
		slice: slice,
		ind:   0,
	}
}

func (s *SliceSearch[T]) Match(val T) bool {
	if s.ind == len(s.slice) {
		s.ind = 0
		return false
	}
	if reflect.DeepEqual(s.slice[s.ind], val) {
		s.ind++
	}
	return s.ind == len(s.slice)
}

func (s *SliceSearch[T]) Reset() { s.ind = 0 }
