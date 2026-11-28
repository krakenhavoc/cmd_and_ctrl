package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Larceny — Enchantment {3}{B}{B} (Eighth Edition, S82 OLD batch 611, #5607):
//
//	"Whenever a creature you control deals combat damage to a player, that
//	 player discards a card."
//
// One trigger per creature that connects (each damage event names its source,
// Sphere Grid's shape); the player it hurt, read off the carried event, chooses
// and discards a card (CR 701.9a: the discarding player picks), and discards
// nothing with an empty hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f356087e-e1d2-4c25-aa1d-215d09535e7c",
		Name:         "Larceny",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CombatDamageToPlayerBy(ev, source.Controller, g)
			}, "Larceny — that player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: item.Trigger.Event.Target, Source: item.SourceCardID, N: 1})
					return nil
				}),
		},
	})
}
