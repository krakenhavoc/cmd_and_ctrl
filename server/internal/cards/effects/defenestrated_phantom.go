package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Defenestrated Phantom — Creature — Spirit {4}{W}{W}, 4/3 (Murders at
// Karlov Manor, #4469):
//
//	"Flying
//	 Disguise {4}{W} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)"
//
// Flying and the disguise keyword (CR 702.168a). Face down it is a 2/2
// with ward {2} and no flying (CR 708.2a); turned face up it is the 4/3
// flier again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "e73addfa-d18a-46f7-882e-7bc2f292230b",
		Name:             "Defenestrated Phantom",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Disguise("{4}{W}")},
	})
}
