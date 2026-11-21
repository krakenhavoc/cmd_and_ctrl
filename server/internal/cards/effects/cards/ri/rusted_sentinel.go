package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rusted Sentinel — Artifact Creature — Golem {4}, 3/4 (Magic 2012, #5573):
//
//	"This creature enters tapped."
//
// Diregraf Ghoul's entry replacement (SelfEntersTapped).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bab28608-b388-4c48-92a4-689943bbbf4a",
		Name:         "Rusted Sentinel",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
	})
}
