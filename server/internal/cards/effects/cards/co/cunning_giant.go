package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cunning Giant — Creature — Giant {5}{R}, 4/4 (Portal Second Age, S82 older
// sets batch 733, #5625):
//
//	"If this creature is unblocked, you may have it assign its combat damage
//	 to a creature defending player controls."
//
// The #6192 seam: its controller is asked in the combat damage step for one
// creature the defending player controls, or the player (CR 510.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                    "4f94ce97-5b2d-42fb-ac8e-c64075fe1af2",
		Name:                        "Cunning Giant",
		Completeness:                effects.CompletenessFull,
		UnblockedAssignsToACreature: true,
	})
}
