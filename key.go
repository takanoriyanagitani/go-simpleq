package simpleq

import (
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

func (k Key) Equal(other Key) bool {
	var gsame bool = OptionEqOpt(k.Group(), other.Group())
	var isame bool = OptionEqOpt(k.Id(), other.Id())
	return gsame && isame
}

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
