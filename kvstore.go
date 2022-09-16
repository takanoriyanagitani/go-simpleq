package simpleq

import (
	"context"
)

type KvStore struct {
	get Get
	set Set
}

// Set sets an item(queue).
type Set func(ctx context.Context, item Item) error

// Get gets an item(queue) by key.
type Get func(ctx context.Context, key Key) Either[Option[Item], error]
