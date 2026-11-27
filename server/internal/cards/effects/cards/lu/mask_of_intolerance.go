package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mask of Intolerance — Artifact {2} (Apocalypse, #5614):
//
//	"At the beginning of each player's upkeep, if there are four or more
//	 basic land types among lands that player controls, this artifact deals 3
//	 damage to that player."
//
// Every player's upkeep triggers it; the count is that player's lands'
// basic land types (domain), read layered, with the intervening if (CR 603.4)
// checked as it triggers and on resolution. The damage goes to the player
// whose upkeep it is, whoever controls the Mask.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "990d4798-3f59-462d-952c-777da5af1c41",
		Name:         "Mask of Intolerance",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB633MaskOfIntolerance("Mask of Intolerance — 3 damage to that player"),
		},
	})
}
