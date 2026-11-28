package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slith Strider — Creature — Slith {1}{U}{U}, 1/1 (Mirrodin, S82 OLD,
// #5606):
//
//	"Whenever this creature becomes blocked, draw a card.
//	 Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it."
//
// Two triggers. "Becomes blocked" is the once-per-attacker EventBecomesBlocked
// (CR 509.1h): a double block draws once, and the card is drawn whatever
// happens to the blockers. The counter trigger is the slith cycle's, once
// for each combat damage event to a player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e52e48f-b7a0-4195-8fff-8baed7b7d3fc",
		Name:         "Slith Strider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBecomesBlocked && ev.CardID == source.InstanceID
			}, "Slith Strider — draw a card", effects.Do(effects.DrawCards{N: 1})),
			effects.WheneverThisDealsCombatDamageToAPlayer("Slith Strider — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: "+1/+1", N: 1})),
		},
	})
}
