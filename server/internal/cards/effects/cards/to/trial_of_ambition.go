package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trial of Ambition — {1}{B} Enchantment (Amonkhet, S82 old batch 248):
//
//	"When this enchantment enters, target opponent sacrifices a creature of
//	 their choice.
//	 When a Cartouche you control enters, return this enchantment to its owner's
//	 hand."
//
// The first trigger is a targeted edict: the opponent chooses the creature as
// it resolves, and a target that is no longer legal sacrifices nothing (CR
// 608.2b). The second is the Trial cycle's Cartouche return.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf614953-bd62-4000-a960-e2c26e1f7bb0",
		Name:         "Trial of Ambition",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.Targeting(
				effects.WhenThisEnters("Trial of Ambition — target opponent sacrifices a creature of their choice",
					effects.TargetPlayerSacrificesACreature("Trial of Ambition — sacrifice a creature")),
				effects.TargetPlayer("target opponent", effects.Opponent()))),
			effects.WhenACartoucheYouControlEntersReturnThis("Trial of Ambition"),
		},
	})
}
