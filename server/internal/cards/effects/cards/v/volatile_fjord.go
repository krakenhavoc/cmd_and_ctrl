package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volatile Fjord — Snow Land — Island Mountain (Kaldheim, #5505):
//
//	"({T}: Add {U} or {R}.)
//	 This land enters tapped."
//
// The Izzet member of Kaldheim's snow-dual cycle, written as Ice
// Tunnel is: the snow supertype rides the printed type line, and the
// two colours are declared because two basic land types on a nonbasic
// land produce nothing on their own in this engine.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "96262c15-9130-409c-8242-71f84be719b7",
		Name:          "Volatile Fjord",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("U", "R")},
	})
}
