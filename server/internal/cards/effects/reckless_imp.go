package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reckless Imp — Creature — Imp {2}{B}, 2/2 (Dragons of Tarkir,
// #5551):
//
//	"Flying
//	 This creature can't block.
//	 Dash {1}{B} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)"
//
// Dash is the shared alternative cost (Ragavan's, CR 702.109): cast for
// its dash cost it gains haste and returns to its owner's hand at the
// beginning of the next end step.
// "Can't block" is Aesthir Glider's self restriction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "e8e88e9a-0097-49ac-8b23-482a3190f21b",
		Name:             "Reckless Imp",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		Static:           []game.StaticAbility{RestrictSelf(game.CantBlock)},
		AlternativeCosts: []game.AlternativeCost{Dash("{1}{B}")},
	})
}
