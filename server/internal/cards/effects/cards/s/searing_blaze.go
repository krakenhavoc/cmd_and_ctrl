package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Searing Blaze — Instant {R}{R}:
//
//	"Searing Blaze deals 1 damage to target player or planeswalker and 1
//	 damage to target creature that player or that planeswalker's controller
//	 controls.
//	 Landfall — If you had a land enter the battlefield under your control
//	 this turn, Searing Blaze deals 3 damage to that player or planeswalker and
//	 3 damage to that creature instead."
//
// Two clauses (CR 601.2c): slot 0 the player or planeswalker, slot 1 a creature
// tied to it (ControlledByTargetsController, Ravager of the Fells' relation), so
// the creature must belong to the player named, or to the planeswalker's
// controller. Both are required targets; each is judged again at resolution
// (CR 608.2b), so a half whose target is gone is dealt nothing while the other
// lands. The damage is one instruction (CR 120), 3 and 3 instead of 1 and 1
// when you had a land enter under your control this turn, read as the spell
// resolves from the turn tally that still counts a land that has since left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74b08a70-b0bb-4340-98a0-b1d5b7c9d2cc",
		Name:         "Searing Blaze",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetPlayerOrPlaneswalker(),
			effects.ControlledByTargetsController(
				effects.TargetCreature("target creature that player or that planeswalker's controller controls"), 0),
		),
		Purpose: effects.ForTargets(effects.DamageToTarget(0, 1), effects.DamageToTarget(1, 1)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 1
			if effects.WwkHadLandfallThisTurn(ctx) {
				n = 3
			}
			return effects.DamageEachClauseTarget(n, n)(item, ctx)
		},
	})
}
