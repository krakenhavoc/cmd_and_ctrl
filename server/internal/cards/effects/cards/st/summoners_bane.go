package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summoner's Bane — Instant {2}{U}{U} (Zendikar, #5583):
//
//	"Counter target creature spell. Create a 2/2 blue Illusion creature
//	 token."
//
// The token is made whether or not the spell was countered, so long as
// the spell is still a legal target as this resolves; with no legal target
// the whole spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9b0c209-e346-45ee-a272-0b53589c21df",
		Name:         "Summoner's Bane",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target creature spell", effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			if err := (effects.CounterTarget{StackID: targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Template: effects.TokenCard("2/2 blue Illusion"), N: 1}.Apply(ctx)
		},
	})
}
