package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heat Wave — Enchantment {2}{R} (Visions, S82 older sets batch 789, #5631):
//
//	"Cumulative upkeep {R}
//	 Blue creatures can't block creatures you control.
//	 Nonblue creatures can't block creatures you control unless their
//	 controller pays 1 life for each blocking creature they control."
//
// A block rule bars blue creatures from blocking any creature the Wave's
// controller controls (CR 509.1b); a block tax (#6191) charges the life, once
// per nonblue creature that blocks at least one of them (CR 509.1d, CR 119.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8f08699-6210-41c3-bfa0-c4b3bcbbaf1c",
		Name:         "Heat Wave",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnMatching(effects.And(effects.Creature(), effects.OfColor("U"))), effects.ControlledBySourceController(),
				"Heat Wave: blue creatures can't block creatures you control"),
		},
		BlockTaxes: []game.BlockTax{
			effects.Old789NonblueCantBlockYourCreaturesUnlessPayLife(1,
				"Nonblue creatures can't block creatures you control unless their controller pays 1 life for each blocking creature they control."),
		},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Heat Wave — cumulative upkeep {R}", "{R}"),
		},
	})
}
