package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avalanche Tusker — Creature — Elephant Warrior {2}{G}{U}{R}, 6/4 (Khans of
// Tarkir):
//
//	"Whenever this creature attacks, target creature defending player controls
//	 blocks it this combat if able."
//
// Provoke without the untap and without the "may": the target is chosen from
// the defending player's creatures as the trigger goes on the stack, and the
// requirement (CR 509.1c) names the Tusker as it attacked, so it asks nothing
// if the target can't block it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ccad8b03-5ff1-4895-94e0-0c96883cb875",
		Name:         "Avalanche Tusker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WithTargetsFrom(
				effects.WheneverThisAttacks("Avalanche Tusker — target creature defending player controls blocks it this combat if able",
					effects.TargetBlocksTheAttacker),
				effects.TargetCreatureDefendingPlayerControls),
		},
	})
}
