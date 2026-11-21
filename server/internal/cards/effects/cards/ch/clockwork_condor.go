package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clockwork Condor — Artifact Creature — Bird {4}, 0/0 (Mirrodin,
// #5606):
//
//	"Flying
//	 This creature enters with three +1/+1 counters on it.
//	 Whenever this creature attacks or blocks, remove a +1/+1 counter
//	 from it at end of combat."
//
// Clockwork Beetle with flying and three counters (clockwork.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3eff8f21-9a1b-4afc-8a7b-75a740d0a4ea",
		Name:            "Clockwork Condor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 3, "Clockwork Condor: enters with three +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			effects.ClockworkLosesACounterAtEndOfCombat("Clockwork Condor"),
		},
	})
}
