package simpleq

import (
	"context"
	"fmt"
)

// KeyGen generates a key.
type KeyGen func(ctx context.Context) Either[Key, error]

// AddKey inserts a key to keys list.
type AddKey func(ctx context.Context, key Key) error

// DelKey removes a key from keys list.
type DelKey func(ctx context.Context, key Key) error

// LstKey gets keys(first in first out order).
type LstKey func(ctx context.Context) Either[Iter[Key], error]

// GetOldestKey gets oldest key if exists.
type GetOldestKey func(ctx context.Context) Either[Option[Key], error]

type KeyManager struct {
	add AddKey
	del DelKey
	lst LstKey
}

type NonAtomicKeyManagerBuilder struct {
	KeyUnpack
	KeySerialize
	KeyBond
	Set
	Get
	Key
}

func (b NonAtomicKeyManagerBuilder) Build() Either[KeyManager, error] {
	var ing Iter[bool] = IterFromArray([]bool{
		nil == b.KeyUnpack,
		nil == b.KeySerialize,
		nil == b.Set,
		nil == b.Get,
	})
	var iok = ing.Map(func(ng bool) (ok bool) { return !ng })
	var ok bool = iok.Reduce(true, func(state bool, b bool) bool {
		return state && b
	})
	var o Option[KeyManager] = OptionFromBool(ok, func() KeyManager {
		var lk LstKey = LstKeyBuilderNew(b.KeyUnpack)(b.Get)(b.Key)
		var nadkb NonAtomicDelKeyBuilder = NonAtomicDelKeyBuilderNewDefault(b.KeySerialize)
		var sk SetKey = SetKeyBuilderNew(b.KeyBond.pack)(b.Set)(b.Key)
		var dkb DelKeyBuilder = nadkb(lk)(sk)
		return KeyManager{
			add: NonAtomicAddKeyBuilderNew(b.KeyBond)(lk)(b.Set)(b.Key),
			del: dkb(b.Key),
			lst: lk,
		}
	})
	return o.OkOrElse(func() error {
		return fmt.Errorf("Invalid kvstore")
	})
}

type KeySerialize func(ctx context.Context, key Key) Either[[]byte, error]
type KeyDeserialize func(ctx context.Context, dat []byte) Either[Key, error]

type KeyAppend func(ctx context.Context, packed []byte, key Key) Either[[]byte, error]

type KeyPack func(ctx context.Context, keys Iter[Key]) Either[[]byte, error]
type KeyUnpack func(ctx context.Context, packed []byte) Either[Iter[Key], error]

func (u KeyUnpack) UnpackData(ctx context.Context, packed Data) Either[Iter[Key], error] {
	return u(ctx, packed.Val())
}

func (u KeyUnpack) UnpackItem(ctx context.Context, packed Item) Either[Iter[Key], error] {
	return u.UnpackData(ctx, packed.Val())
}

func (u KeyUnpack) UnpackOpt(ctx context.Context, packed Option[Item]) Either[Iter[Key], error] {
	var oeik Option[Either[Iter[Key], error]] = OptionMap(
		packed,
		func(i Item) Either[Iter[Key], error] {
			return u.UnpackItem(ctx, i)
		},
	)
	return oeik.UnwrapOrElse(func() Either[Iter[Key], error] {
		return EitherOk(IterEmpty[Key]())
	})
}

func KeyPackBuilderNew(ser KeySerialize) func(a KeyAppend) KeyPack {
	return func(a KeyAppend) KeyPack {
		return func(ctx context.Context, keys Iter[Key]) Either[[]byte, error] {
			reducer := func(state Either[[]byte, error], item Key) Either[[]byte, error] {
				return state.FlatMap(func(packed []byte) Either[[]byte, error] {
					return a(ctx, packed, item)
				})
			}
			return IterReduce(keys, EitherOk[[]byte](nil), reducer)
		}
	}
}

type KeyBond struct {
	pack KeyPack
	join KeyAppend
}

type KeyBondBuilder struct {
	KeySerialize
	KeyAppend
}

func (b KeyBondBuilder) Build() Either[KeyBond, error] {
	var valid bool = nil != b.KeySerialize && nil != b.KeyAppend
	var okb Option[KeyBond] = OptionFromBool(valid, func() KeyBond {
		return KeyBond{
			pack: KeyPackBuilderNew(b.KeySerialize)(b.KeyAppend),
			join: b.KeyAppend,
		}
	})
	return okb.OkOrElse(func() error { return fmt.Errorf("Invalid builder") })
}

func (k KeyBond) GetPackedKeys(ctx context.Context, lk LstKey) Either[[]byte, error] {
	var eik Either[Iter[Key], error] = lk(ctx)
	return EitherFlatMap(eik, func(ik Iter[Key]) Either[[]byte, error] {
		return k.pack(ctx, ik)
	})
}

func (k KeyBond) AppendKey(ctx context.Context, packed []byte, key Key) Either[[]byte, error] {
	return k.join(ctx, packed, key)
}

type AddKeyBuilder func(Key) AddKey

type LstKeyBuilder func(Key) LstKey

type DelKeyBuilder func(Key) DelKey

type GetKeyRemoved func(keys Iter[Key]) func(removeMe Key) Iter[Key]

func GetKeyRemovedBuilderNew(iflt IterFilter[Key]) func(ser KeySerialize) GetKeyRemoved {
	return func(ser KeySerialize) GetKeyRemoved {
		return func(keys Iter[Key]) func(Key) Iter[Key] {
			return func(removeMe Key) Iter[Key] {
				return iflt(keys, func(key Key) bool {
					var ksame bool = removeMe.Equal(key)
					return !ksame
				})
			}
		}
	}
}

var GetKeyRemovedBuilderDefault func(ser KeySerialize) GetKeyRemoved = GetKeyRemovedBuilderNew(
	IterFilterDefaultNew[Key](),
)

type SetKey func(ctx context.Context, keys Iter[Key]) error

func SetKeyBuilderNew(pack KeyPack) func(Set) func(Key) SetKey {
	return func(set Set) func(Key) SetKey {
		return func(mng Key) SetKey {
			return func(ctx context.Context, keys Iter[Key]) error {
				var packed Either[[]byte, error] = pack(ctx, keys)
				var edata Either[Data, error] = EitherMap(packed, DataNew)
				var eitem Either[Item, error] = EitherMap(edata, func(d Data) Item {
					return ItemNew(mng, d)
				})
				var ee Either[error, error] = EitherMap(eitem, func(i Item) error {
					return set(ctx, i)
				})
				return ee.UnwrapOrElse(Identity[error])
			}
		}
	}
}

type NonAtomicDelKeyBuilder func(LstKey) func(SetKey) DelKeyBuilder

func NonAtomicDelKeyBuilderNew(gkr GetKeyRemoved) NonAtomicDelKeyBuilder {
	return func(lk LstKey) func(SetKey) DelKeyBuilder {
		return func(sk SetKey) DelKeyBuilder {
			return func(mng Key) DelKey {
				return func(ctx context.Context, key Key) error {
					var ekeys Either[Iter[Key], error] = lk(ctx)
					var removed Either[Iter[Key], error] = ekeys.Map(func(i Iter[Key]) Iter[Key] {
						return gkr(i)(key)
					})
					var ee Either[error, error] = EitherMap(removed, func(i Iter[Key]) error {
						return sk(ctx, i)
					})
					return ee.UnwrapOrElse(Identity[error])
				}
			}
		}
	}
}

var NonAtomicDelKeyBuilderNewDefault func(ser KeySerialize) NonAtomicDelKeyBuilder = Compose(
	GetKeyRemovedBuilderDefault,
	NonAtomicDelKeyBuilderNew,
)

func LstKeyBuilderNew(unpack KeyUnpack) func(Get) LstKeyBuilder {
	return func(get Get) LstKeyBuilder {
		return func(mng Key) LstKey {
			return func(ctx context.Context) Either[Iter[Key], error] {
				var eoi Either[Option[Item], error] = get(ctx, mng)
				return EitherFlatMap(eoi, func(oi Option[Item]) Either[Iter[Key], error] {
					return unpack.UnpackOpt(ctx, oi)
				})
			}
		}
	}
}

func NonAtomicAddKeyBuilderNew(kb KeyBond) func(LstKey) func(Set) AddKeyBuilder {
	return func(lk LstKey) func(Set) AddKeyBuilder {
		return func(set Set) AddKeyBuilder {
			return func(mng Key) AddKey {
				return func(ctx context.Context, key Key) error {
					var epacked Either[[]byte, error] = kb.GetPackedKeys(ctx, lk)
					var appended Either[[]byte, error] = epacked.FlatMap(
						func(packed []byte) Either[[]byte, error] {
							return kb.AppendKey(ctx, packed, key)
						},
					)
					var edata Either[Data, error] = EitherMap(appended, DataNew)
					var eitem Either[Item, error] = EitherMap(edata, func(d Data) Item {
						return ItemNew(mng, d)
					})
					var ee Either[error, error] = EitherMap(eitem, func(i Item) error {
						return set(ctx, i)
					})
					return ee.UnwrapOrElse(Identity[error])
				}
			}
		}
	}
}
