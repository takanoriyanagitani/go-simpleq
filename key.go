package simpleq

import (
	"context"
	"fmt"
)

type GroupId struct {
	raw string
}

func GroupIdNew(raw string) GroupId { return GroupId{raw} }

type Id struct {
	raw string
}

func IdNew(raw string) Id { return Id{raw} }

type Key struct {
	groupId Option[GroupId]
	id      Option[Id]
}

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

type KeyGen func(ctx context.Context) Either[Key, error]
