package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hulk, Strongest There Is — Legendary Creature — Gamma Berserker Hero
// {2}{G}{G}, 3/3 (Marvel Super Heroes Commander):
//
//	"Trample
//	 Hulk enters with a +1/+1 counter on him.
//	 At the beginning of your upkeep, double the number of +1/+1 counters on
//	 each Gamma creature you control."
//
// The entry counter is Afiya Grove's self-replacement, so it is on him before
// any state check. The upkeep trigger is Bristly Bill's doubling
// (doublePlusOneCountersOnEach) over the Gamma creatures: the set and the
// counts are read first, then each gets as many +1/+1 counters as it had.
// Hulk is a Gamma himself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "35daea08-655d-4588-8430-b078769d22b6",
		Name:            "Hulk, Strongest There Is",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 1, "Hulk: enters with a +1/+1 counter on him"),
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Hulk — double the +1/+1 counters on each Gamma creature you control", func(g *game.Game, item *game.StackItem) error {
				gamma := NewContext(g, item).TextSubtype("Gamma")
				return doublePlusOneCountersOnEach(g, item, func(c game.Card) bool {
					return c.Controller == item.Controller && c.IsCreature() && c.HasSubtype(gamma)
				})
			}),
		},
	})
}
