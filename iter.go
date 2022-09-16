package simpleq

type Iter[T any] func() Option[T]

func (i Iter[T]) TryForEach(f func(T) error) error {
	for o := i(); o.HasValue(); o = i() {
		var t T = o.Value()
		var e error = f(t)
		if nil != e {
			return e
		}
	}
	return nil
}

func IterReduce[T, U any](i Iter[T], init U, reducer func(U, T) U) U {
	state := init
	for o := i(); o.HasValue(); o = i() {
		var val T = o.Value()
		state = reducer(state, val)
	}
	return state
}

func IterTryCollect[T any](i Iter[Either[Option[T], error]]) Either[[]T, error] {
	reducer := func(state Either[[]T, error], item Either[Option[T], error]) Either[[]T, error] {
		return state.Map(func(collected []T) []T {
			return collected
		})
	}
	return IterReduce(i, EitherOk[[]T](nil), reducer)
}
