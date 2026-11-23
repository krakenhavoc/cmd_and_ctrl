package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pensive Professor — Creature — Human Wizard {1}{U}{U}, 0/2 (Secrets of
// Strixhaven, #4456):
//
//	"Increment (Whenever you cast a spell, if the amount of mana you spent is
//	 greater than this creature's power or toughness, put a +1/+1 counter on
//	 this creature.)
//	 Whenever one or more +1/+1 counters are put on this creature, draw a
//	 card."
//
// Increment is the keyword of soc_b19_helpers.go. The draw is Scurry Oak's
// reading of "one or more counters are put on": a counter-placement event
// that raised this creature's +1/+1 total (b33CountersPlacedDelta), so a
// removal does not count and several counters in one event are one draw. The
// increment's own counter, a counter put by another card, and a doubling
// effect all draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4b8736e2-fc41-4a25-bb6c-7351afc40981",
		Name:            "Pensive Professor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.KeywordIncrement},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered: []game.TriggeredAbility{
			effects.Increment(),
			effects.On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Target == source.InstanceID && effects.B33CountersPlacedDelta(ev, game.CounterPlusOne, g) > 0
			}, "Pensive Professor — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
