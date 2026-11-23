package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rakdos Cackler — Creature — Devil {B/R}, 1/1 (Return to Ravnica, #5566):
//
//	"Unleash (You may have this creature enter with a +1/+1 counter on
//	 it. It can't block as long as it has a +1/+1 counter on it.)"
//
// Unleash is the engine's keyword (ADR 0109 §10): the entering "may" and
// the can't-block while it has a +1/+1 counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "763a2d27-52a8-43d9-b649-0f6c0cf87a07",
		Name:            "Rakdos Cackler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordUnleash},
	})
}
