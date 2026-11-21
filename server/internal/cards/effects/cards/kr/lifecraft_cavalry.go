package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lifecraft Cavalry — Creature — Elf Warrior {4}{G}, 4/4 (Aether Revolt,
// #5536):
//
//	"Trample
//	 Revolt — This creature enters with two +1/+1 counters on it if a
//	 permanent left the battlefield under your control this turn."
//
// Narnam Renegade's revolt self-replacement with two counters
// (revolt.go, CR 614.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "13db0230-b57e-47fd-9025-580d06cc88ee",
		Name:            "Lifecraft Cavalry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersWithCountersIfRevolt(game.CounterPlusOne, 2)},
	})
}
