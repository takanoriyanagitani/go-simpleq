package simpleq

import (
	"testing"
)

func TestKvstore(t *testing.T) {
	t.Parallel()

	t.Run("KvStoreBuilder", func(t *testing.T) {
		t.Parallel()

		t.Run("Build", func(t *testing.T) {
			t.Parallel()

			t.Run("invalid", func(t *testing.T) {
				t.Parallel()

				var b KvStoreBuilder = KvStoreBuilderNew()
				var ek Either[KvStore, error] = b.Build()
				t.Run("Must invalid", checker(ek.IsNg(), true))
			})
		})
	})
}
