package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scalding Tongs — Artifact {2} (Tempest, S82 OLD batch 757, #5628):
//
//	"At the beginning of your upkeep, if you have three or fewer cards in
//	 hand, this artifact deals 1 damage to target opponent or
//	 planeswalker."
//
// An intervening "if" (CR 603.4): the hand is counted when the upkeep
// begins and again as the ability resolves, so it never even targets with
// four cards in hand. The target is chosen as it goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "939a0cdf-b5e8-4d7d-adb3-f7b2fd7aab8f",
		Name:         "Scalding Tongs",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.InterveningIfYou("you have three or fewer cards in hand",
				func(g *game.Game, you uuid.UUID) bool { return effects.CardsInHand(g, you) <= 3 },
				effects.AtYourUpkeep("Scalding Tongs — 1 damage to target opponent or planeswalker", effects.DealDamageToTheTarget(1))),
				effects.TargetOpponentOrPlaneswalker()),
		},
	})
}
