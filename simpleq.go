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

// Del deletes queue by key.
// If key does not exists, this must return nil.
type Del func(ctx context.Context, key Key) error

// Get gets oldest queue if it exists.
type Get func(ctx context.Context) Either[Option[Item], error]

// Upsert inserts/overwrites queue.
type Upsert func(ctx context.Context, item Item) error

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
