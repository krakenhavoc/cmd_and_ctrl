package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rotting Legion — Creature — Zombie {4}{B}, 4/5 (Magic 2011, #5579):
//
//	"This creature enters tapped."
//
// Shambling Cie'th's CR 614 self-replacement: it is never untapped on
// the battlefield, rather than tapped a beat after it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e5fd8b4-d81a-4ed6-b389-21445351bb43",
		Name:         "Rotting Legion",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
	})
}
