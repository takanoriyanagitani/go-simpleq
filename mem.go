package simpleq

import (
	"context"
	"fmt"
)

func MapKvstoreNew() Either[KvStore, error] {
	m := make(map[GroupId]map[Id]Data)

	getOrNew := func(g GroupId) map[Id]Data {
		return MapGet(m, g).UnwrapOrElse(func() map[Id]Data {
			var sm map[Id]Data = make(map[Id]Data)
			m[g] = sm
			return sm
		})
	}

	newGetById := func(g GroupId) func(i Id) Option[Data] {
		return func(i Id) Option[Data] {
			var om Option[map[Id]Data] = MapGet(m, g)
			return OptionFlatMap(om, func(sm map[Id]Data) Option[Data] {
				return MapGet(sm, i)
			})
		}
	}

	k2group := func(k Key) Either[GroupId, error] {
		return k.Group().
			OkOrElse(func() error { return fmt.Errorf("Group id missing") })
	}

	k2id := func(k Key) Either[Id, error] {
		return k.Id().
			OkOrElse(func() error { return fmt.Errorf("Id missing") })
	}

	k2dat := func(k Key) func(g GroupId) Either[Option[Data], error] {
		return func(g GroupId) Either[Option[Data], error] {
			var ei Either[Id, error] = k2id(k)
			return EitherMap(ei, newGetById(g))
		}
	}

	k2data := func(k Key) Either[Option[Data], error] {
		return ComposeEither(
			k2group,
			k2dat(k),
		)(k)
	}

	k2item := func(k Key) Either[Option[Item], error] {
		var ed Either[Option[Data], error] = k2data(k)
		return EitherMap(ed, func(o Option[Data]) Option[Item] {
			return OptionMap(o, func(d Data) Item {
				return ItemNew(k, d)
			})
		})
	}

	get := func(_ context.Context, key Key) Either[Option[Item], error] {
		return k2item(key)
	}

	item2mem := func(dat Data) func(g GroupId) func(i Id) error {
		return func(g GroupId) func(Id) error {
			return func(i Id) error {
				var sm map[Id]Data = getOrNew(g)
				sm[i] = dat
				return nil
			}
		}
	}

	k2set := func(key Key) func(g GroupId) Either[func(Item) error, error] {
		return func(g GroupId) Either[func(Item) error, error] {
			var ei Either[Id, error] = k2id(key)
			return EitherMap(ei, func(i Id) func(Item) error {
				return func(itm Item) error {
					var d Data = itm.Val()
					return item2mem(d)(g)(i)
				}
			})
		}
	}

	k2setter := func(k Key) Either[func(Item) error, error] {
		return ComposeEither(
			k2group,
			k2set(k),
		)(k)
	}

	set := func(_ context.Context, item Item) error {
		var key Key = item.Key()
		var es Either[func(Item) error, error] = k2setter(key)
		var ee Either[error, error] = EitherMap(es, func(f func(Item) error) error {
			return f(item)
		})
		return ee.UnwrapOrElse(Identity[error])
	}

	key2del := func(g GroupId) func(i Id) error {
		return func(i Id) error {
			var om Option[map[Id]Data] = MapGet(m, g)
			om.ForEach(func(m map[Id]Data) {
				delete(m, i)
			})
			return nil
		}
	}

	k2del := func(key Key) func(g GroupId) Either[func() error, error] {
		return func(g GroupId) Either[func() error, error] {
			var ei Either[Id, error] = k2id(key)
			return EitherMap(ei, func(i Id) func() error {
				return func() error {
					return key2del(g)(i)
				}
			})
		}
	}

	k2remover := func(k Key) Either[func() error, error] {
		return ComposeEither(
			k2group,
			k2del(k),
		)(k)
	}

	del := func(_ context.Context, key Key) error {
		var ef Either[func() error, error] = k2remover(key)
		var ee Either[error, error] = EitherMap(ef, func(f func() error) error {
			return f()
		})
		return ee.UnwrapOrElse(Identity[error])
	}

	return KvStoreBuilderNew().
		WithGet(get).
		WithSet(set).
		WithDel(del).
		Build()
}
