package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shrieking Affliction — Enchantment {B} (Return to Ravnica, #5566):
//
//	"At the beginning of each opponent's upkeep, if that player has one or fewer
//	 cards in hand, they lose 3 life."
//
// An intervening "if" (CR 603.4): the opponent's hand is checked as their upkeep
// begins and again as the ability resolves, so a player who draws or discards in
// response changes the answer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e3f5d0b-103f-40e2-beb5-26b1fcb136f3",
		Name:         "Shrieking Affliction",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.Actor != source.Controller && effects.CardsInHand(g, ev.Actor) <= 1
			}, "Shrieking Affliction — that player loses 3 life", func(g *game.Game, item *game.StackItem) error {
				player := effects.NewContext(g, item).Trigger().Event.Actor
				if player == uuid.Nil || effects.CardsInHand(g, player) > 1 {
					return nil
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, player, -3)
			}),
		},
	})
}
