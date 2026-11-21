package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	enters := WhenThisEnters("Proud Pack-Rhino — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = ChooseOne(
		ModeDoing("Put a shield counter on target permanent.",
			TargetPermanent("target permanent"),
			func(_ *game.StackItem, ctx *Context, _ int) error { return shieldCounterOnFirstTarget(ctx) }),
		ModeSteps("Proliferate.", nil, Proliferate{}),
	)
	Register(Spec{
		OracleID:     "292cd535-916e-486e-b0cc-ecdb9477287d",
		Name:         "Proud Pack-Rhino",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
