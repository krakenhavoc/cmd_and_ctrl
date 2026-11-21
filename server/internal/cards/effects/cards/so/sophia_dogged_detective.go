package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sophia, Dogged Detective — Legendary Creature — Human Detective
// {1}{G}{W}{U}, 3/4 (Murders at Karlov Manor Commander):
//
//	"When Sophia enters, create Tiny, a legendary 2/2 green Dog Detective
//	 creature token with trample.
//	 {1}, Sacrifice an artifact token: Put a +1/+1 counter on each Dog you
//	 control.
//	 Whenever a Dog you control deals combat damage to a player, create a
//	 Food token, then investigate."
//
// Three pieces. The token is the printed Tiny. The sacrifice cost takes any
// artifact token (a Clue, a Food, a Treasure), paid up front, and the
// counters go on every Dog read as the ability resolves
// (avB30aPutCounterOnEachOfSubtypeYouControl, Lo and Li's shape). The
// third is per Dog, not per batch: each Dog that connects triggers once,
// so two Dogs hitting the same player make two Foods and two Clues.
// The Food comes first, then the Clue (CR 701.16a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eaf8027a-017a-4471-ae7b-a372fd8e0ff3",
		Name:         "Sophia, Dogged Detective",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sophia, Dogged Detective — create Tiny, a legendary 2/2 Dog Detective with trample",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Legendary Dog Detective named Tiny with trample"), N: 1})),
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.CombatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.Source)
				return ok && dealer.HasSubtype(source.TextSubtype("Dog"))
			}, "Sophia, Dogged Detective — create a Food token, then investigate",
				effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1}, effects.Investigate{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice an artifact token: Put a +1/+1 counter on each Dog you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeN(1, "an artifact token", effects.Artifact(), effects.IsTokenPredicate())),
			Effect:  effects.AvB30aPutCounterOnEachOfSubtypeYouControl("Dog"),
		}},
	})
}
