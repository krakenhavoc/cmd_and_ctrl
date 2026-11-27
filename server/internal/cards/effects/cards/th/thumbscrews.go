package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thumbscrews — Artifact {2} (Tempest, S82 OLD batch 758, #5628):
//
//	"At the beginning of your upkeep, if you have five or more cards in
//	 hand, this artifact deals 1 damage to target opponent or
//	 planeswalker."
//
// An intervening "if" (CR 603.4): the hand is counted when the upkeep
// begins and again as the ability resolves; the target is chosen as it goes
// on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8f1f55a-7a85-4338-b4d1-3d28f10f69a7",
		Name:         "Thumbscrews",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.InterveningIfYou("you have five or more cards in hand",
				func(g *game.Game, you uuid.UUID) bool { return effects.CardsInHand(g, you) >= 5 },
				effects.AtYourUpkeep("Thumbscrews — 1 damage to target opponent or planeswalker", effects.DealDamageToTheTarget(1))),
				effects.TargetOpponentOrPlaneswalker()),
		},
	})
}
