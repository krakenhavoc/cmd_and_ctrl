package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bandage — Instant {W} (Stronghold, #5627):
//
//	"Prevent the next 1 damage that would be dealt to any target this
//	 turn.
//	 Draw a card."
//
// Caregiver's and Benevolent Ancestor's shared
// preventNextOneDamageToTheTarget body, a one-point shield on the target
// that lasts until it is used or the turn ends (CR 615.7), then the
// cantrip. The instructions run in order (CR 608.2c). The spell's only
// target is the shield's: if that target is gone in response, the spell
// doesn't resolve (CR 608.2b) and no card is drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c664df64-0103-462b-a0f8-c483b152974a",
		Name:         "Bandage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPrevent, Draws: 1},
		Targets:      effects.TargetAny(),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.PreventNextOneDamageToTheTarget("Bandage")(ctx.Game, ctx.Item)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
