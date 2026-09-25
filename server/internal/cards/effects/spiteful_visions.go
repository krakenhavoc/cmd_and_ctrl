package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spiteful Visions — Enchantment {2}{B/R}{B/R} (EDHREC rank 2303):
//
//	"At the beginning of each player's draw step, that player draws
//	 an additional card.
//	 Whenever a player draws a card, this enchantment deals 1 damage
//	 to that player."
//
// Howling Mine with a tax on every card. Two triggers:
//
//   - The extra draw is Howling Mine's draw-step trigger without
//     the intervening-if: EACH player's draw step, the drawer is the
//     event's Actor rather than the controller, and the trigger
//     arrives after the turn-based draw (CR 504.1 does not use the
//     stack), which is what "an additional card" means.
//   - The damage is "whenever a player draws" — anyone's draw, the
//     controller's included, one event per card, so a draw-three is
//     three damage. Dealt by the enchantment, so it is noncombat
//     damage from a black-and-red source and a prevention shield
//     sees it. The extra draw above is itself a draw and is taxed,
//     as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "922cf963-2b1b-43ad-819e-6e49133e6aae",
		Name:         "Spiteful Visions",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventBeginDrawStep},
				AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor != uuid.Nil
				},
				Key: "Spiteful Visions — that player draws an additional card",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Trigger.Event.Actor, N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Watches: []game.EventKind{game.EventDrawCard},
				AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor != uuid.Nil
				},
				Key: "Spiteful Visions — 1 damage to the player who drew",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DealDamage{Source: item.SourceCardID, Target: item.Trigger.Event.Actor, Amount: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
