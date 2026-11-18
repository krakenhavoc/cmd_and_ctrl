package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Molten Monstrosity — Creature — Hellion {7}{R}, 5/5 (Dominaria United,
// S82 DMU batch 4):
//
//	"This spell costs {X} less to cast, where X is the greatest power among
//	 creatures you control.
//	 Trample"
//
// Ghalta the Unstoppable's self cost modifier over the greatest power among
// the creatures you control; it spends generic mana only. Trample is a
// printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ea55886c-48a1-4339-81fd-03dbaa116c28",
		Name:            "Molten Monstrosity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(greatestPowerAmongCreaturesYouControl(),
				"This spell costs {X} less to cast, where X is the greatest power among creatures you control."),
		},
	})
}
