package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Somberwald Spider — Creature — Spider {4}{G}, 2/4 (Innistrad, S82 OLD
// batch 450, #5572):
//
//	"Reach
//	 Morbid — This creature enters with two +1/+1 counters on it if a
//	 creature died this turn."
//
// Festerhide Boar's self-replacement (CR 614.1c) with two counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8b3526c4-efc8-42be-a0ab-de10c4cc986d",
		Name:            "Somberwald Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersWithCountersIfMorbid(game.CounterPlusOne, 2)},
	})
}
