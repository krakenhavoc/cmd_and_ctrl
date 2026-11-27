package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mishra's Helix — Artifact {5} (Urza's Saga, #5624):
//
//	"{X}, {T}: Tap X target lands."
//
// The number of targets is the announced X (game.TargetSpec.CountFromX, as
// Icy Blast's). X of zero taps nothing.
//
// No simplification.
func init() {
	targets := effects.TargetPermanent("X target lands", effects.Land())
	targets.CountFromX = true
	effects.Register(effects.Spec{
		OracleID:     "6b1552f2-df64-485a-8509-b1916edf19cf",
		Name:         "Mishra's Helix",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}: Tap X target lands.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
			Targets: targets,
			Effect:  effects.S82OldB726TapEachLegalTarget,
		}},
	})
}
