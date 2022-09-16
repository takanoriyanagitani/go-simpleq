package simpleq

import (
	"context"
	"fmt"
)

// KeyGen generates a key.
type KeyGen func(ctx context.Context) Either[Key, error]

// AddKey inserts a key.
type AddKey func(ctx context.Context, key Key) error

// DelKey removes a key.
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

type KeyManagerBuilder struct {
	AddKey
	DelKey
	LstKey
}

func (b KeyManagerBuilder) Build() Either[KeyManager, error] {
	var ing Iter[bool] = IterFromArray([]bool{
		nil == b.AddKey,
		nil == b.DelKey,
		nil == b.LstKey,
	})
	var iok = ing.Map(func(ng bool) (ok bool) { return !ng })
	var ok bool = iok.Reduce(true, func(state bool, b bool) bool {
		return state && b
	})
	var o Option[KeyManager] = OptionFromBool(ok, func() KeyManager {
		return KeyManager{
			add: b.AddKey,
			del: b.DelKey,
			lst: b.LstKey,
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

func NonAtomicAddKeyBuilderNew(pack KeyPack) func(KeyAppend) func(LstKey) func(Set) func(Key) AddKey {
	return func(ap KeyAppend) func(LstKey) func(Set) func(Key) AddKey {
		return func(lst LstKey) func(Set) func(Key) AddKey {
			return func(set Set) func(Key) AddKey {
				return func(keymng Key) AddKey {
					return func(ctx context.Context, key Key) error {
						var eik Either[Iter[Key], error] = lst(ctx)
						var epacked Either[[]byte, error] = EitherFlatMap(eik, func(ik Iter[Key]) Either[[]byte, error] {
							return pack(ctx, ik)
						})
						var appended Either[[]byte, error] = EitherFlatMap(epacked, func(packed []byte) Either[[]byte, error] {
							return ap(ctx, packed, key)
						})
						var edata Either[Data, error] = EitherMap(appended, DataNew)
						var eitem Either[Item, error] = EitherMap(edata, func(d Data) Item {
							return ItemNew(keymng, d)
						})
						var ee Either[error, error] = EitherMap(eitem, func(item Item) error {
							return set(ctx, item)
						})
						return ee.UnwrapOrElse(Identity[error])
					}
				}
			}
		}
	}
}
