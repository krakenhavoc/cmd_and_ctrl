package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jackal Pup — Creature — Jackal {R}, 2/1 (Tempest, S82 OLD batch 758,
// #5628):
//
//	"Whenever this creature is dealt damage, it deals that much damage to
//	 you."
//
// Brash Taunter's enrage trigger without the target: the damage dealt to the
// Pup at once (one batch, CR 603.2c) is one trigger, and the Pup, the source
// of the reflected damage, deals it to its controller even if it has died of
// the damage (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "993e41e6-0643-4a1f-8ada-b7a4aae486d9",
		Name:         "Jackal Pup",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Jackal Pup — it deals that much damage to you", effects.FiredrinkerSatyrReflect),
		},
	})
}
