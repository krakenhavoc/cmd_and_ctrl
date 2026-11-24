package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Joel, Resolute Survivor — Legendary Creature — Human Survivor {3}{B}{G},
// 4/4 (Secret Lair Drop):
//
//	"Menace
//	 Whenever a creature token dies, put a +1/+1 counter on Joel and draw a
//	 card. This ability triggers only once each turn.
//	 Partner—Survivors (You can have two commanders if both have this
//	 ability.)"
//
// Yahenni's dies condition (the permanent's type as it last existed, CR 603.10a)
// narrowed to a token, whoever controls it. The once-a-turn limit is the
// trigger's (TriggersOnlyOnceEachTurn), so the second token to die in a turn
// does nothing. Both halves are one resolution: the counter, then the draw (a
// Joel that has left only draws, CR 608.2b). Partner—Survivors is a
// deck-construction rule read off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "64611260-c9f7-4375-b67a-f88fee6d3bff",
		Name:            "Joel, Resolute Survivor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && dead.IsToken()
			}, "Joel, Resolute Survivor — put a +1/+1 counter on Joel and draw a card",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}, effects.DrawCards{N: 1}))),
		},
	})
}
