package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coastal Peak — Land — Island Mountain (Marvel Super Heroes Commander,
// #4455): "({T}: Add {U} or {R}.) This land enters tapped. Cycling {2}".
// The blue-red member of the cycling-land cycle, the same shape as
// Canyon Slough.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "0e25faa2-efb0-4ca5-b280-18c38faa860c",
		Name:          "Coastal Peak",
		Completeness:  effects.CompletenessFull,
		Activated:     []effects.ActivatedAbility{effects.Cycling("{2}")},
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("U", "R")},
	})
}
