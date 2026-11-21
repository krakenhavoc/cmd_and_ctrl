package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "ea55886c-48a1-4339-81fd-03dbaa116c28",
		Name:            "Molten Monstrosity",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.GreatestPowerAmongCreaturesYouControl(),
				"This spell costs {X} less to cast, where X is the greatest power among creatures you control."),
		},
	})
}
