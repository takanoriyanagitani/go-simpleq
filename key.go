package simpleq

import (
	"context"
	"fmt"
)

// GroupId is group id for a queue.
type GroupId struct {
	raw string
}

func GroupIdNew(raw string) GroupId { return GroupId{raw} }

// Id is sub id for a queue
type Id struct {
	raw string
}

func IdNew(raw string) Id { return Id{raw} }

// Key is a unique identifier for a queue.
type Key struct {
	groupId Option[GroupId]
	id      Option[Id]
}

func (k Key) Group() Option[GroupId] { return k.groupId }
func (k Key) Id() Option[Id]         { return k.id }

// KeyNew creates new key.
// GroupId or Id must have a value.
func KeyNew(groupId Option[GroupId], id Option[Id]) Either[Key, error] {
	var valid bool = groupId.HasValue() || id.HasValue()
	var okey Option[Key] = OptionFromBool(valid, func() Key {
		return Key{
			groupId,
			id,
		}
	})
	return okey.OkOrElse(func() error { return fmt.Errorf("Id missing") })
}

func (k Key) WithId(id Id) Key {
	k.id = OptionNew(id)
	return k
}

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
