package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Falcon and Redwing — Legendary Creature — Human Bird Hero {2}{W}, 1/2
// (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever Falcon and Redwing deal combat damage to a player, create
//	 that many 1/1 white Bird creature tokens with flying, then put a
//	 +1/+1 counter on Falcon and Redwing."
//
// "That many" is the damage the triggering event dealt, read as the
// trigger resolves. The counter comes after the tokens, and only while
// the source is still the same permanent on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0f7c7035-e395-4da5-8536-9f192d03e3ba",
		Name:            "Falcon and Redwing",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Falcon and Redwing — create that many Birds, then a +1/+1 counter on this creature",
				effects.Do(
					effects.CreateTokensForEach{
						Template: effects.TokenCard("1/1 white Bird with flying"),
						Name:     "the combat damage dealt",
						Count:    func(ctx *effects.Context) int { return ctx.Trigger().Event.Amount },
					},
					effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1},
				)),
		},
	})
}
