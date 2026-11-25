package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treetop Ambusher — Creature — Elf Berserker {1}{G}, 2/1 (Modern Horizons):
//
//	"Dash {1}{G} (You may cast this spell for its dash cost. If you do, it
//	 gains haste, and it's returned from the battlefield to its owner's hand at
//	 the beginning of the next end step.)
//	 Whenever this creature attacks, target creature you control gets +1/+1
//	 until end of turn."
//
// Dash is the shared alternative cost (CR 702.109). The attack trigger
// targets a creature its controller controls as it goes on the stack (CR
// 603.3d), the Ambusher itself included, and a target gone by resolution
// does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "72949d87-76cd-472c-86b3-2b7d5fce3159",
		Name:             "Treetop Ambusher",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Dash("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Treetop Ambusher — target creature you control gets +1/+1 until end of turn",
				effects.BoostTargetUntilEOT("Treetop Ambusher — +1/+1", 1, 1)),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
