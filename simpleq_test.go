package simpleq

import (
	"testing"
)

func checkBuilderNew[T any](comp func(a, b T) (same bool)) func(a, b T) func(t *testing.T) {
	return func(got, expected T) func(t *testing.T) {
		return func(t *testing.T) {
			var same bool = comp(got, expected)
			if !same {
				t.Errorf("Unexpected value got.\n")
				t.Errorf("Expected: %v\n", expected)
				t.Fatalf("Got:      %v\n", got)
			}
		}
	}
}

func checker[T comparable](a, b T) func(t *testing.T) {
	comp := func(x, y T) (same bool) { return x == y }
	return checkBuilderNew(comp)(a, b)
}

func TestAll(t *testing.T) {
}
