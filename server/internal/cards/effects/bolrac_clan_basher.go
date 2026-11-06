package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bolrac-Clan Basher — Creature — Cyclops Warrior {4}{R}{R}, 3/2
// (Murders at Karlov Manor, #4469):
//
//	"Double strike, trample
//	 Disguise {3}{R}{R} (You may cast this card face down for {3} as a
//	 2/2 creature with ward {2}. Turn it face up any time for its
//	 disguise cost.)"
//
// Two printed keywords and the disguise keyword (Forum Familiar's
// shape, ADR 0082). The face-down 2/2 has neither keyword (CR 708.2a);
// both come back when it is turned face up.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "c23617c9-53f9-4063-8241-b504e9b80cbe",
		Name:             "Bolrac-Clan Basher",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"double strike", "trample"},
		AlternativeCosts: []game.AlternativeCost{Disguise("{3}{R}{R}")},
	})
}
