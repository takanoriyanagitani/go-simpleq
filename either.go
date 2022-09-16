package simpleq

type eitherUnwrapOr[T any] func(alt T) T

type eitherTryForEach[T, E any] func(f func(T) E) E
type eitherMap[T, E any] func(f func(T) T) Either[T, E]
type eitherFlatMap[T, E any] func(f func(T) Either[T, E]) Either[T, E]
type eitherUnwrapOrElse[T, E any] func(f func(E) T) T

type Either[T, E any] struct {
	left         Option[E]
	right        Option[T]
	unwrapOr     eitherUnwrapOr[T]
	tryForEach   eitherTryForEach[T, E]
	emap         eitherMap[T, E]
	flatMap      eitherFlatMap[T, E]
	unwrapOrElse eitherUnwrapOrElse[T, E]
}

func (e Either[T, E]) IsOk() bool                                  { return e.right.HasValue() }
func (e Either[T, E]) IsNg() bool                                  { return !e.IsOk() }
func (e Either[T, E]) TryForEach(f func(T) E) E                    { return e.tryForEach(f) }
func (e Either[T, E]) Map(f func(T) T) Either[T, E]                { return e.emap(f) }
func (e Either[T, E]) FlatMap(f func(T) Either[T, E]) Either[T, E] { return e.flatMap(f) }
func (e Either[T, E]) Ok() Option[T]                               { return e.right }
func (e Either[T, E]) UnwrapOrElse(f func(E) T) T                  { return e.unwrapOrElse(f) }

func EitherRight[T, E any](t T) Either[T, E] {
	return Either[T, E]{
		left:         OptionEmpty[E](),
		right:        OptionNew(t),
		unwrapOr:     func(_ T) T { return t },
		unwrapOrElse: func(_ func(E) T) T { return t },
		tryForEach:   func(f func(T) E) E { return f(t) },
		emap:         func(f func(T) T) Either[T, E] { return EitherRight[T, E](f(t)) },
		flatMap:      func(f func(T) Either[T, E]) Either[T, E] { return f(t) },
	}
}

func EitherLeft[T, E any](e E) Either[T, E] {
	return Either[T, E]{
		left:         OptionNew(e),
		right:        OptionEmpty[T](),
		unwrapOr:     func(alt T) T { return alt },
		unwrapOrElse: func(f func(E) T) T { return f(e) },
		tryForEach:   func(_ func(T) E) E { return e },
		emap:         func(_ func(T) T) Either[T, E] { return EitherLeft[T, E](e) },
		flatMap:      func(_ func(T) Either[T, E]) Either[T, E] { return EitherLeft[T, E](e) },
	}
}

func EitherOk[T any](t T) Either[T, error]     { return EitherRight[T, error](t) }
func EitherNg[T any](e error) Either[T, error] { return EitherLeft[T, error](e) }

func EitherNew[T any](t T, e error) Either[T, error] {
	if nil == e {
		return EitherOk(t)
	}
	return EitherNg[T](e)
}

func EitherMap[T, U, E any](e Either[T, E], f func(T) U) Either[U, E] {
	if e.IsOk() {
		var t T = e.right.Value()
		var u U = f(t)
		return EitherRight[U, E](u)
	}
	return EitherLeft[U, E](e.left.Value())
}

func EitherFlatMap[T, U, E any](e Either[T, E], f func(T) Either[U, E]) Either[U, E] {
	if e.IsOk() {
		var t T = e.right.Value()
		return f(t)
	}
	return EitherLeft[U, E](e.left.Value())
}
