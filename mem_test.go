package simpleq

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
)

func TestMem(t *testing.T) {
	t.Parallel()

	var dummyKeyGenBuilder func() KeyGen = func() KeyGen {
		var i2g func(int) GroupId = Compose(strconv.Itoa, GroupIdNew)
		var i2i func(int) Id = Compose(strconv.Itoa, IdNew)

		var i2og func(int) Option[GroupId] = Compose(i2g, OptionNew[GroupId])
		var i2oi func(int) Option[Id] = Compose(i2i, OptionNew[Id])

		return func(_ context.Context) Either[Key, error] {
			return KeyNew(
				i2og(rand.Int()),
				i2oi(rand.Int()),
			)
		}
	}

	t.Run("MapKvstoreNew", func(t *testing.T) {
		t.Parallel()

		var dummyKeyGen KeyGen = dummyKeyGenBuilder()

		var ek Either[KvStore, error] = MapKvstoreNew()
		e := ek.TryForEach(func(k KvStore) error {
			t.Run("kvstore got", func(t *testing.T) {
				t.Run("Get", func(t *testing.T) {
					var ek Either[Key, error] = dummyKeyGen(context.Background())
					var e error = ek.TryForEach(func(key Key) error {
						var eoi Either[Option[Item], error] = k.Get(context.Background(), key)
						t.Run("optional item got", checker(eoi.IsOk(), true))
						return eoi.TryForEach(func(o Option[Item]) error {
							t.Run("item empty", checker(o.HasValue(), false))
							return nil
						})
					})
					t.Run("key got", checker(nil == e, true))

					t.Run("invalid key", func(t *testing.T) {
						t.Run("missing group", func(t *testing.T) {
							var ek Either[Key, error] = KeyNew(
								OptionEmpty[GroupId](),
								OptionNew(IdNew("")),
							)
							var e error = ek.TryForEach(func(key Key) error {
								var eoi Either[Option[Item], error] = k.Get(context.Background(), key)
								t.Run("invalid key", checker(eoi.IsNg(), true))
								return nil
							})
							t.Run("key got", checker(nil == e, true))
						})

						t.Run("missing id", func(t *testing.T) {
							var ek Either[Key, error] = KeyNew(
								OptionNew(GroupIdNew("")),
								OptionEmpty[Id](),
							)
							var e error = ek.TryForEach(func(key Key) error {
								var eoi Either[Option[Item], error] = k.Get(context.Background(), key)
								t.Run("invalid key", checker(eoi.IsNg(), true))
								return nil
							})
							t.Run("key got", checker(nil == e, true))
						})
					})
				})

				t.Run("Set", func(t *testing.T) {
					var ek Either[Key, error] = dummyKeyGen(context.Background())
					var e error = ek.TryForEach(func(key Key) error {
						t.Run("empty", func(t *testing.T) {
							var dat Data = DataNew(nil)
							var itm Item = ItemNew(key, dat)
							var e error = k.Set(context.Background(), itm)
							t.Run("item set", checker(nil == e, true))
						})
						return nil
					})
					t.Run("key got", checker(nil == e, true))
				})

				t.Run("Del", func(t *testing.T) {
					var ek Either[Key, error] = dummyKeyGen(context.Background())
					var e error = ek.TryForEach(func(key Key) error {
						t.Run("missing key", func(t *testing.T) {
							var e error = k.Del(context.Background(), key)
							t.Run("key absent", checker(nil == e, true))
						})

						t.Run("key exists", func(t *testing.T) {
							var dat Data = DataNew([]byte("hw"))
							var itm Item = ItemNew(key, dat)
							var e error = k.Set(context.Background(), itm)
							t.Run("item tried to set", checker(nil == e, true))

							var eoi Either[Option[Item], error] = k.Get(context.Background(), key)
							e = eoi.TryForEach(func(oi Option[Item]) error {
								t.Run("item got", checker(oi.HasValue(), true))
								return nil
							})
							t.Run("tried to get item", checker(nil == e, true))

							e = k.Del(context.Background(), key)
							t.Run("tried to remove item", checker(nil == e, true))

							eoi = k.Get(context.Background(), key)
							e = eoi.TryForEach(func(oi Option[Item]) error {
								t.Run("item missing", checker(oi.HasValue(), false))
								return nil
							})
							t.Run("tried to get removed item", checker(nil == e, true))
						})
						return nil
					})
					t.Run("key got", checker(nil == e, true))
				})
			})
			return nil
		})
		t.Run("kvstore got", checker(nil == e, true))
	})
}
