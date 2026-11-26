package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Needle Drop — Instant {R} (Lorwyn, S82 OLD b546, #5593):
//
//	"Needle Drop deals 1 damage to any target that was dealt damage this
//	 turn.
//	 Draw a card."
//
// The target must already have been dealt damage this turn when it is
// chosen (a creature by the turn's damage events, a player by the turn's
// damage-taken tally), and again as the spell resolves (CR 608.2b); it need
// not be damaged still. The draw is the second sentence, so it happens
// whenever the spell resolves at all.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8622a2b2-fe07-41bd-bef3-e60731e0d701",
		Name:         "Needle Drop",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 1)),
		Targets:      effects.S82OldB546AnyTargetDealtDamageThisTurn(),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DamageToFirstTargetStep(1)(ctx.Item, ctx)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
