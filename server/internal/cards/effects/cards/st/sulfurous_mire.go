package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sulfurous Mire — Snow Land — Swamp Mountain (Kaldheim, #5505):
//
//	"({T}: Add {B} or {R}.)
//	 This land enters tapped."
//
// The Rakdos member of Kaldheim's snow-dual cycle, written as Ice
// Tunnel is: the snow supertype rides the printed type line, and the
// two colours are declared because two basic land types on a nonbasic
// land produce nothing on their own in this engine.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "77cf536e-246e-4c18-8e04-c904fcad3f40",
		Name:          "Sulfurous Mire",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("B", "R")},
	})
}
