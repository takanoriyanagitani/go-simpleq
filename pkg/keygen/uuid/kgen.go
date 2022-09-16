package kgen

import (
	"context"

	"github.com/google/uuid"

	sq "github.com/takanoriyanagitani/go-simpleq"
)

type GenStr func(ctx context.Context) sq.Either[string, error]

func GenBuilderNew(gen GenStr) sq.KeyGen {
	return func(ctx context.Context) sq.Either[sq.Key, error] {
		var og sq.Option[string] = gen(ctx).Ok()
		var oi sq.Option[string] = gen(ctx).Ok()
		var g sq.Option[sq.GroupId] = sq.OptionMap(og, sq.GroupIdNew)
		var i sq.Option[sq.Id] = sq.OptionMap(oi, sq.IdNew)
		return sq.KeyNew(g, i)
	}
}

type Uuid2str func(u uuid.UUID) string

type RandomGen func(ctx context.Context) sq.Either[uuid.UUID, error]

func GenStrBuilderNew(u2s Uuid2str) func(rg RandomGen) GenStr {
	return func(rg RandomGen) GenStr {
		return func(ctx context.Context) sq.Either[string, error] {
			var eu sq.Either[uuid.UUID, error] = rg(ctx)
			return sq.EitherMap(eu, u2s)
		}
	}
}

var RandomGenDefault RandomGen = func(_ctx context.Context) sq.Either[uuid.UUID, error] {
	return sq.EitherNew(uuid.NewRandom())
}

var Uuid2StrDefault Uuid2str = func(u uuid.UUID) string { return u.String() }

var GenStrDefault GenStr = GenStrBuilderNew(Uuid2StrDefault)(RandomGenDefault)

var KeyGenDefault sq.KeyGen = GenBuilderNew(GenStrDefault)
