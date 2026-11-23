package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Proud Pack-Rhino — Creature — Rhino {2}{W}, 3/3 (Modern Horizons 3):
//
//	"When this creature enters, choose one —
//	 • Put a shield counter on target permanent. (If it would be dealt
//	   damage or destroyed, remove a shield counter from it instead.)
//	 • Proliferate. (Choose any number of permanents and/or players, then
//	   give each another counter of each kind already there.)"
//
// Adaptive Sporesinger's modal enters trigger (CR 603.3c): the mode is
// chosen as the trigger goes on the stack and only the first bullet
// targets, any permanent. The counter is the engine's shield counter
// (shieldCounterOnFirstTarget, Kros's); a target gone by resolution gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Proud Pack-Rhino — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeDoing("Put a shield counter on target permanent.",
			effects.TargetPermanent("target permanent"),
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.ShieldCounterOnFirstTarget(ctx)
			}),
		effects.ModeSteps("Proliferate.", nil, effects.Proliferate{}),
	)
	effects.Register(effects.Spec{
		OracleID:     "292cd535-916e-486e-b0cc-ecdb9477287d",
		Name:         "Proud Pack-Rhino",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
