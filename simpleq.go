package simpleq

import (
	"context"
)

type Data struct {
	raw []byte
}

type Item struct {
	key Key
	val Data
}

func ItemNew(key Key, val Data) Item {
	return Item{
		key,
		val,
	}
}

func (i Item) Key() Key { return i.key }

// Del deletes queue by key.
// If key does not exists, this must return nil.
type Del func(ctx context.Context, key Key) error

// Get gets oldest queue if it exists.
type Get func(ctx context.Context) Either[Option[Item], error]

// Set sets queue.
type Set func(ctx context.Context, item Item) error

// AddKey inserts key.
type AddKey func(ctx context.Context, key Key) error

// DelKey removes key.
type DelKey func(ctx context.Context, key Key) error

// LstKey gets keys.
type LstKey func(ctx context.Context) Either[Iter[Key], error]

// Upsert inserts/overwrites queue.
type Upsert func(ctx context.Context, item Item) error

// NonAtomicUpsertBuilderNew creates upsert.
// 1. add key.
// 2. add item(data).
// Note: This may create a queue without data on power failure.
func NonAtomicUpsertBuilderNew(ak AddKey, s Set) Upsert {
	return func(ctx context.Context, item Item) error {
		var k Key = item.Key()
		var e error = ak(ctx, k)
		if nil != e {
			return e
		}
		return s(ctx, item)
	}
}

// Push inserts queue.
// Key will be auto generated.
type Push func(ctx context.Context, data Data) error

func PushBuilderNew(keygen KeyGen, upsert Upsert) Push {
	return func(ctx context.Context, data Data) error {
		var ek Either[Key, error] = keygen(ctx)
		var ei Either[Item, error] = EitherMap(ek, func(k Key) Item {
			return ItemNew(k, data)
		})
		return ei.TryForEach(func(item Item) error {
			return upsert(ctx, item)
		})
	}
}
