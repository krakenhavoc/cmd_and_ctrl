package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guardian Seraph — Creature — Angel {2}{W}{W}, 3/4 (Magic 2010, #5585):
//
//	"Flying
//	 If a source an opponent controls would deal damage to you, prevent 1 of
//	 that damage."
//
// Heart-Shaped Herb's prevention (CR 615): damage of 1 is prevented
// entirely, larger damage loses 1, read against the source's last-known
// controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fe9106e3-c637-4995-9dc1-cae0be31f23b",
		Name:            "Guardian Seraph",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{effects.OldB499PreventOneOfTheDamageToYou("Guardian Seraph")},
	})
}
