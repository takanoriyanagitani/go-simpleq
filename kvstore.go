package simpleq

import (
	"context"
	"fmt"
)

type KvStore struct {
	get Get
	set Set
	del Del
}

func KvStoreNew(get Get, set Set, del Del) Either[KvStore, error] {
	return EitherNg[KvStore](nil)
}

type KvStoreBuilder struct {
	Get
	Set
	Del
}

func KvStoreBuilderNew() (b KvStoreBuilder) { return }

func (b KvStoreBuilder) WithGet(g Get) KvStoreBuilder {
	b.Get = g
	return b
}

func (b KvStoreBuilder) WithSet(s Set) KvStoreBuilder {
	b.Set = s
	return b
}

func (b KvStoreBuilder) WithDel(d Del) KvStoreBuilder {
	b.Del = d
	return b
}

func (b KvStoreBuilder) Build() Either[KvStore, error] {
	var ing Iter[bool] = IterFromArray([]bool{
		nil == b.Get,
		nil == b.Set,
		nil == b.Del,
	})
	var iok = ing.Map(func(ng bool) (ok bool) { return !ng })
	var ok bool = iok.Reduce(true, func(state bool, b bool) bool {
		return state && b
	})
	var o Option[KvStore] = OptionFromBool(ok, func() KvStore {
		return KvStore{
			get: b.Get,
			set: b.Set,
			del: b.Del,
		}
	})
	return o.OkOrElse(func() error {
		return fmt.Errorf("Invalid kvstore")
	})
}

// Set sets an item(queue).
type Set func(ctx context.Context, item Item) error

// Get gets an item(queue) by key.
type Get func(ctx context.Context, key Key) Either[Option[Item], error]

// Del deletes a queue by the key.
// This should not make an error if the key does not exists.
type Del func(ctx context.Context, key Key) error
