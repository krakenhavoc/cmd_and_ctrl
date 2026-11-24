package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Davriel, Rogue Shadowmage — Legendary Planeswalker — Davriel {2}{B},
// loyalty 3:
//
//	"At the beginning of each opponent's upkeep, if that player has one or
//	 fewer cards in hand, Davriel deals 2 damage to them.
//	 −1: Target player discards a card."
//
// The intervening "if" (CR 603.4) is checked when the upkeep begins and again
// as the trigger resolves, against the upkeep's player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1bd5e51f-f9a5-4eb7-bda2-957310602c20",
		Name:            "Davriel, Rogue Shadowmage",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.AllOf(effects.ByAnOpponent,
				func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.CardsInHand(g, ev.Actor) <= 1
				}),
				"Davriel, Rogue Shadowmage — 2 damage to that player",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Trigger().Event.Actor
					if player == [16]byte{} || effects.CardsInHand(g, player) > 1 {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: player, Amount: 2}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "−1: Target player discards a card.",
			Cost:    effects.LoyaltyCost(-1),
			Targets: effects.TargetPlayer("target player"),
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
			Effect:  effects.TargetPlayerDiscards(1),
		}},
	})
}
