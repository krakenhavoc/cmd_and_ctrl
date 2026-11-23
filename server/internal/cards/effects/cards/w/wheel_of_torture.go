package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wheel of Torture — Artifact {3} (Urza's Legacy, #5623):
//
//	"At the beginning of each opponent's upkeep, this artifact deals X
//	 damage to that player, where X is 3 minus the number of cards in
//	 their hand."
//
// Iron Maiden's trigger with The Rack's amount: on each opponent's upkeep it
// deals that player 3 minus their hand size. X is counted as the trigger
// resolves (CR 608.2h), so a player with three or more cards takes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a3e6e84-a8c2-472d-99fb-7cd14850b5ce",
		Name:         "Wheel of Torture",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachOpponentsUpkeep("Wheel of Torture — deals damage to that player equal to 3 minus their hand size",
				effects.DamageUpkeepPlayerByHand(func(hand int) int { return 3 - hand })),
		},
	})
}
