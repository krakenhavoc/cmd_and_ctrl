package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blunt the Assault — Instant {3}{G} (Scars of Mirrodin, #5577):
//
//	"You gain 1 life for each creature on the battlefield. Prevent all
//	 combat damage that would be dealt this turn."
//
// Fog of War's life count, over every player's creatures as the spell
// resolves, then Darkness's prevention shield for the rest of the turn
// (CR 615.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dccd0533-fda0-41ff-b540-c4f1bb32fc95",
		Name:         "Blunt the Assault",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPrevent},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				n := len(ctx.CreatureIDs())
				if n == 0 {
					return nil
				}
				return effects.GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.PreventAllCombatDamageThisTurn{Label: "Blunt the Assault: prevent combat damage"}.Apply(ctx)
			}),
		),
	})
}
