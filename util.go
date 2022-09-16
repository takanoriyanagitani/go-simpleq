package simpleq

func ComposeEither[T, U, V, E any](f func(T) Either[U, E], g func(U) Either[V, E]) func(T) Either[V, E] {
	return func(t T) Either[V, E] {
		var eu Either[U, E] = f(t)
		return EitherFlatMap(eu, g)
	}
}
