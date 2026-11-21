package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flare of Denial — Instant {1}{U}{U} (Modern Horizons 3, #5948):
//
//	"You may sacrifice a nontoken blue creature rather than pay this
//	 spell's mana cost.
//	 Counter target spell."
//
// Counterspell's body with Flare of Fortitude's alternative cost: the
// creature is sacrificed as the spell goes on the stack (CR 118.9), so a
// countered Flare still costs it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "437cfd16-94bc-43ba-afd8-e9e9d028a4ca",
		Name:         "Flare of Denial",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a counterspell, as Counterspell
		Targets:      TargetSpell("target spell"),
		AlternativeCosts: []game.AlternativeCost{
			SacrificeInstead(1, "a nontoken blue creature", 0, Creature(), OfColor("U"), Not(IsTokenPredicate())),
		},
		OnResolve: counterTheTargetSpell,
	})
}
