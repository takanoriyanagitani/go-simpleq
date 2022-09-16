package simpleq

func ComposeEither[T, U, V, E any](f func(T) Either[U, E], g func(U) Either[V, E]) func(T) Either[V, E] {
	return func(t T) Either[V, E] {
		var eu Either[U, E] = f(t)
		return EitherFlatMap(eu, g)
	}
}

func CoalesceError(e1 error, e2 error) error {
	if nil == e1 {
		return e2
	}
	return e1
}

func Identity[T any](t T) T { return t }

func MapGet[T comparable, U any](m map[T]U, t T) Option[U] {
	u, found := m[t]
	if found {
		return OptionNew(u)
	}
	return OptionEmpty[U]()
}
