package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Highly Illogical — Instant {1}{U} (Star Trek, #4973):
//
//	"Kicker {1}{U} (You may pay an additional {1}{U} as you cast this
//	 spell.)
//	 Counter target spell with mana value 2 or less. If this spell was
//	 kicked, instead counter target spell."
//
// Bloodchief's Thirst's shape (WhenPaid, #1716): the kicker is announced
// at CR 601.2b, before targets at CR 601.2c, so the kicked cast has the
// wider clause and an unkicked one may not name a three-drop. The mana
// value is read on the stack (CR 202.3), so an X spell counts its X.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "38c43f30-0e01-4dcd-88fc-f1de318a5e29",
		Name:          "Highly Illogical",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{WhenPaid(Kicker("{1}{U}"), TargetSpell("target spell"))},
		Targets:       TargetSpell("target spell with mana value 2 or less", ManaValueLE(2)),
		OnResolve:     counterTheTargetSpell,
	})
}
