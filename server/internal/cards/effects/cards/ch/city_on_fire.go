package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// City on Fire — Enchantment {5}{R}{R}{R} (EDHREC rank 870):
//
//	"Convoke
//	 If a source you control would deal damage to a permanent or
//	 player, it deals triple that damage instead."
//
// Fiery Emancipation at eight mana with convoke to pay for it — the
// same replacement, and the S22 tap-permanents cost on the cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41eec4e8-92d3-4f98-9346-3a3e3cc602ce",
		Name:         "City on Fire",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Replacements: []game.ReplacementEffect{effects.B07TripleDamageFromYourSources()},
	})
}
