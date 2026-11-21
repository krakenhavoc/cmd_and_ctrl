package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mortus Strider — Creature — Skeleton {1}{U}{B}, 1/1 (Gatecrash, #5565):
//
//	"When this creature dies, return it to its owner's hand."
//
// Endless Cockroaches' death trigger: the card returns from its owner's
// graveyard, and not if it has left the graveyard first (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15af3bd0-022b-472e-b21e-29b925b7692b",
		Name:         "Mortus Strider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Mortus Strider — return it to its owner's hand", effects.ReturnSelfFromGraveyardToHand),
		},
	})
}
