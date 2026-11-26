package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hellfire Mongrel — Creature — Elemental Dog {2}{R} 2/2 (Zendikar, S82 OLD
// batch 492, #5583):
//
//	"At the beginning of each opponent's upkeep, if that player has two or
//	 fewer cards in hand, this creature deals 2 damage to that player."
//
// An intervening "if" (CR 603.4): the upkeep's player has two or fewer cards in
// hand as their upkeep begins and again as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "633a0919-9ac3-4645-8f0b-3f00c198db79",
		Name:         "Hellfire Mongrel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "that player has two or fewer cards in hand",
				Test: func(ctx *effects.Context) bool {
					return ctx.Item != nil && ctx.Item.Trigger != nil && effects.B14HandSize(ctx.Game, ctx.Item.Trigger.Event.Actor) <= 2
				},
			}, effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ByAnOpponent(ev, source, lki, g) && effects.B14HandSize(g, ev.Actor) <= 2
			}, "Hellfire Mongrel — 2 damage to that player", effects.ScgB1DamageThatPlayer(2))),
		},
	})
}
