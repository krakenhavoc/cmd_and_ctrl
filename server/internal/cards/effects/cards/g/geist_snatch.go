package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geist Snatch — Instant {2}{U}{U}:
//
//	"Counter target creature spell. Create a 1/1 blue Spirit creature token
//	 with flying."
//
// The token is made even when the spell is no longer there? No: with its
// only target gone the spell does not resolve (CR 608.2b), so no token.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5fcfcf5-8af5-4395-b989-427fa92e0dd2",
		Name:         "Geist Snatch",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target creature spell", effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			if err := (effects.CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 blue Spirit with flying"), N: 1}.Apply(ctx)
		},
	})
}
