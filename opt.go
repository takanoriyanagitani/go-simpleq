package simpleq

type optValue[T any] func() T
type optEmpty func() bool
type optOkOrElse[T any] func(ng func() error) Either[T, error]
type optFilter[T any] func(flt func(T) bool) Option[T]
type optForEach[T any] func(f func(T))
type optUnwrapOr[T any] func(alt T) T
type optMap[T any] func(f func(T) T) Option[T]
type optUnwrapOrElse[T any] func(f func() T) T

type Option[T any] struct {
	value        optValue[T]
	empty        optEmpty
	okOrElse     optOkOrElse[T]
	filter       optFilter[T]
	unwrapOr     optUnwrapOr[T]
	unwrapOrElse optUnwrapOrElse[T]
	omap         optMap[T]
	forEach      optForEach[T]
}

func (o Option[T]) Value() T                                  { return o.value() }
func (o Option[T]) Empty() bool                               { return o.empty() }
func (o Option[T]) HasValue() bool                            { return !o.Empty() }
func (o Option[T]) OkOrElse(ng func() error) Either[T, error] { return o.okOrElse(ng) }
func (o Option[T]) Filter(flt func(T) bool) Option[T]         { return o.filter(flt) }
func (o Option[T]) UnwrapOr(t T) T                            { return o.unwrapOr(t) }
func (o Option[T]) UnwrapOrElse(f func() T) T                 { return o.unwrapOrElse(f) }
func (o Option[T]) Map(f func(T) T) Option[T]                 { return o.omap(f) }
func (o Option[T]) ForEach(f func(T))                         { o.forEach(f) }

func OptionNew[T any](t T) Option[T] {
	return Option[T]{
		value:        func() T { return t },
		empty:        func() bool { return false },
		okOrElse:     func(_ func() error) Either[T, error] { return EitherOk(t) },
		unwrapOr:     func(_ T) T { return t },
		unwrapOrElse: func(_ func() T) T { return t },
		omap:         func(f func(T) T) Option[T] { return OptionNew(f(t)) },
		forEach:      func(f func(T)) { f(t) },
		filter: func(flt func(T) bool) Option[T] {
			if flt(t) {
				return OptionNew(t)
			}
			return OptionEmpty[T]()
		},
	}
}

func OptionEmpty[T any]() Option[T] {
	return Option[T]{
		value:        func() (t T) { return },
		empty:        func() bool { return true },
		okOrElse:     func(ng func() error) Either[T, error] { return EitherNg[T](ng()) },
		filter:       func(_ func(T) bool) Option[T] { return OptionEmpty[T]() },
		unwrapOr:     func(alt T) T { return alt },
		unwrapOrElse: func(f func() T) T { return f() },
		omap:         func(_ func(T) T) Option[T] { return OptionEmpty[T]() },
		forEach:      func(_ func(T)) {},
	}
}

func OptionFromBool[T any](b bool, f func() T) Option[T] {
	if b {
		var t T = f()
		return OptionNew(t)
	}
	return OptionEmpty[T]()
}

func OptionMap[T, U any](o Option[T], f func(T) U) Option[U] {
	if o.HasValue() {
		var t T = o.Value()
		var u U = f(t)
		return OptionNew(u)
	}
	return OptionEmpty[U]()
}

func OptionFromArray[T any](a []T, ix int) Option[T] {
	if ix < len(a) {
		var t T = a[ix]
		return OptionNew(t)
	}
	return OptionEmpty[T]()
}

func OptionFlatMap[T, U any](o Option[T], f func(T) Option[U]) Option[U] {
	if o.HasValue() {
		var t T = o.Value()
		return f(t)
	}
	return OptionEmpty[U]()
}
