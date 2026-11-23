package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guiding Hydra — Creature — Hydra Horror {X}{W}, 1/0:
//
//	"This creature enters with X +1/+1 counters on it.
//	 At the beginning of combat on your turn, you may remove a +1/+1
//	 counter from this creature. If you do, put a +1/+1 counter on each
//	 other creature you control."
//
// The X counters are the printed entry clause (XCounters, the Voracious
// Hydra shape). The combat trigger is a "you may" asked before the
// ability goes on the stack; on resolution it removes a counter and only
// then, if one was removed, spreads one to each other creature. The set
// is snapshotted before any counter lands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "24f1445c-16c9-45c0-be56-0266e6c78cdb",
		Name:                       "Guiding Hydra",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtBeginningOfYourCombat("Guiding Hydra — remove a +1/+1 counter to put one on each other creature you control", effects.RfCreatureCGuidingHydraEffect),
				"Remove a +1/+1 counter from Guiding Hydra to put a +1/+1 counter on each other creature you control?"),
		},
	})
}
