package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ebony Owl Netsuke — Artifact {2} (Saviors of Kamigawa, S82 OLD b583):
//
//	"At the beginning of each opponent's upkeep, if that player has seven or
//	 more cards in hand, this artifact deals 4 damage to that player."
//
// Iron Maiden's trigger with an intervening "if" (CR 603.4): the upkeep's
// player (the event's actor) must have seven or more cards as the upkeep
// begins, and the hand is counted again as the ability resolves. It is damage
// from the artifact, so prevention and damage triggers see a source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcacb717-4ca4-47c0-9b23-26e4eb4ffdc8",
		Name:         "Ebony Owl Netsuke",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.AllOf(effects.ByAnOpponent, ebonyOwlNetsukeHandIsFull),
				"Ebony Owl Netsuke — deals 4 damage to that player",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DamageUpkeepPlayerByHand(func(hand int) int {
						if hand < 7 {
							return 0
						}
						return 4
					})(g, item)
				}),
		},
	})
}

func ebonyOwlNetsukeHandIsFull(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	return effects.CardsInHand(g, ev.Actor) >= 7
}
