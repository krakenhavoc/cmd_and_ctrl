package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steelbane Hydra — {X}{G}{G} Creature — Turtle Hydra 0/0 (reprinted in
// Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"This creature enters with X +1/+1 counters on it.
//	 {2}{G}, Remove a +1/+1 counter from this creature: Destroy target
//	 artifact or enchantment."
//
// The entry counters are the cast's X (XCounters, Feral Hydra's shape);
// the removal cost is the CR 602 counter-removal cost District Mascot
// pays (RemoveCountersFromThis), so a Hydra with no counter left cannot
// activate, and paying the last one leaves a 0/0 that dies to state-based
// actions with the ability still on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "e38e3723-05f5-4a51-8364-1cda19f9cc49",
		Name:                       "Steelbane Hydra",
		Completeness:               effects.CompletenessFull,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, Remove a +1/+1 counter from this creature: Destroy target artifact or enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
