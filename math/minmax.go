package armath

import "golang.org/x/exp/constraints"

func MinMax[T constraints.Ordered](a, b T) (min, max T) {
	if a < b {
		min = a
		max = b
		return
	}
	min = b
	max = a
	return
}

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T constraints.Ordered](a, b T) T {
	if a < b {
		return b
	}
	return a
}
