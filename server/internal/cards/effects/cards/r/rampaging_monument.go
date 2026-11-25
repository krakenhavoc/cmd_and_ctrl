package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rampaging Monument — Artifact Creature — Cleric {4}, 0/0 (Guilds of
// Ravnica):
//
//	"Trample
//	 This creature enters with three +1/+1 counters on it.
//	 Whenever you cast a multicolored spell, put a +1/+1 counter on this
//	 creature."
//
// Mindless Automaton's entry counters (CR 614.1c) and a cast trigger over
// multicolored spells; the counter is put on only while the Monument is still
// on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3984478b-debb-4a65-ab5b-4bc134e03d82",
		Name:            "Rampaging Monument",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 3, "Rampaging Monument: enters with three +1/+1 counters on it"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Multicolored(), "Rampaging Monument — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
