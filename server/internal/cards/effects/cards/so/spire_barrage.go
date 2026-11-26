package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spire Barrage — Sorcery {4}{R} (Zendikar, S82 OLD batch 492, #5583):
//
//	"Spire Barrage deals damage to any target equal to the number of Mountains
//	 you control."
//
// The Mountains are counted as it resolves, through the layers, with the land
// type read through the spell's text (CR 612.2), as Spitting Earth counts them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef84d5e2-5db7-40e7-af49-a2f6f494a8a4",
		Name:         "Spire Barrage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.B08LandsWithSubtypeControlled(ctx.Game, ctx.Controller(), ctx.TextSubtype("Mountain"))
			for _, t := range ctx.LegalTargets() {
				return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
			}
			return nil
		},
	})
}
