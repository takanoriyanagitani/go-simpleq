package simpleq

import (
	"context"
)

// Data contains queue data.
type Data struct {
	raw []byte
}

// Item is a queue with its identifier(key) and its contents(data).
type Item struct {
	key Key
	val Data
}

// ItemNew creates an item(queue).
func ItemNew(key Key, val Data) Item {
	return Item{
		key,
		val,
	}
}

func (i Item) Key() Key { return i.key }

// Del deletes a queue by the key.
// This should not make an error if the key does not exists.
type Del func(ctx context.Context, key Key) error

// Get gets the oldest queue if it exists.
type Get func(ctx context.Context) Either[Option[Item], error]

// Set sets an item(queue).
type Set func(ctx context.Context, item Item) error

// AddKey inserts a key.
type AddKey func(ctx context.Context, key Key) error

// DelKey removes a key.
type DelKey func(ctx context.Context, key Key) error

// LstKey gets keys(first in first out order).
type LstKey func(ctx context.Context) Either[Iter[Key], error]

// Upsert inserts/overwrites an item(queue).
type Upsert func(ctx context.Context, item Item) error

// NonAtomicUpsertBuilderNew creates upsert.
// 1. Add a key.
// 2. Add an item(queue).
// Note: This may create a queue without data on power failure/crash.
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

// Push inserts a queue.
type Push func(ctx context.Context, data Data) error

// PushBuilderNew creates Push.
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
