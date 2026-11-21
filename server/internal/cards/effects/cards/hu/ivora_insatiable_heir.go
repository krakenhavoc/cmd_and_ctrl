package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ivora, Insatiable Heir — Legendary Creature — Vampire Warrior {1}{R}, 1/1
// (Foundations Jumpstart):
//
//	"Trample
//	 When Ivora enters and whenever it deals combat damage to a player,
//	 create a Blood token.
//	 Whenever you discard a card, put a +1/+1 counter on Ivora."
//
// The Blood line is one printed ability with two conditions, so it is one
// registered trigger over two events (OnAny, as "enters or attacks" is). The
// discard line counts every card YOU discard, whatever made you, and Ivora
// herself must be on the battlefield to grow.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d847542b-7522-43ec-91b0-f0aba5d7dddd",
		Name:            "Ivora, Insatiable Heir",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventDealDamage},
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventETB {
						return effects.Self(ev, source, lki, g)
					}
					return effects.ThisDealtCombatDamageToAPlayer(ev, source, lki, g)
				},
				"Ivora, Insatiable Heir — create a Blood token", effects.Do(effects.CreateToken{Template: effects.BloodToken(), N: 1})),
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, "Ivora, Insatiable Heir — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
