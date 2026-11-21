package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cinder Barrens — Land (Oath of the Gatewatch, #5545):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {R}."
//
// The Guildgates' shape: it enters tapped (a replacement, CR 614.1c) and
// taps for either colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "f155a05f-9f5e-4875-a407-103b85ce30ee",
		Name:          "Cinder Barrens",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("B", "R")},
	})
}
